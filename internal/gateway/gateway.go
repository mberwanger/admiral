package gateway

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/uber-go/tally/v4"
	tallyprom "github.com/uber-go/tally/v4/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.admiral.io/admiral/internal/config"
	"go.admiral.io/admiral/internal/endpoint"
	"go.admiral.io/admiral/internal/gateway/mux"
	"go.admiral.io/admiral/internal/gateway/stats"
	"go.admiral.io/admiral/internal/middleware"
	"go.admiral.io/admiral/internal/middleware/accesslog"
	"go.admiral.io/admiral/internal/middleware/errorintercept"
	"go.admiral.io/admiral/internal/middleware/timeouts"
	"go.admiral.io/admiral/internal/service"
)

type ComponentFactory struct {
	Services   service.Factory
	Middleware middleware.Factory
	Endpoints  endpoint.Factory
}

func Run(cfg *config.Config, cf *ComponentFactory, assets http.FileSystem) {
	// Init the server's logger.
	logger, err := newLogger(cfg.Server.Logger)
	if err != nil {
		panic(err)
	}
	// nolint
	defer logger.Sync()

	// Init stats.
	scopeOpts, metricsHandler := getStatsReporterConfiguration(cfg, logger)

	scope, scopeCloser := tally.NewRootScope(
		scopeOpts,
		cfg.Server.Stats.FlushInterval,
	)
	defer func() {
		if err := scopeCloser.Close(); err != nil {
			panic(err)
		}
	}()

	initScope := scope.SubScope("gateway")
	initScope.Counter("start").Inc(1)

	// Create the error interceptor so services can register error interceptors if desired.
	errorInterceptMiddleware, err := errorintercept.NewMiddleware(nil, logger, initScope)
	if err != nil {
		logger.Fatal("could not create error interceptor middleware", zap.Error(err))
	}

	// Instantiate and register services.
	for name, factory := range cf.Services {
		logger := logger.With(zap.String("serviceName", name))

		logger.Info("registering service")
		svc, err := factory(cfg, logger, scope.SubScope(name))
		if err != nil {
			logger.Fatal("service instantiation failed", zap.Error(err))
		}
		service.Registry[name] = svc

		if ei, ok := svc.(errorintercept.Interceptor); ok {
			logger.Info("service registered an error conversion interceptor")
			errorInterceptMiddleware.AddInterceptor(ei.InterceptError)
		}
	}

	var interceptors []grpc.UnaryServerInterceptor

	// Error interceptors should be first on the stack (last in chain).
	interceptors = append(interceptors, errorInterceptMiddleware.UnaryInterceptor())

	// Access log.
	if cfg.Server.AccessLog != nil {
		a, err := accesslog.New(cfg.Server.AccessLog, logger, scope)
		if err != nil {
			logger.Fatal("could not create accesslog interceptor", zap.Error(err))
		}
		interceptors = append(interceptors, a.UnaryInterceptor())
	}

	// Timeouts.
	timeoutInterceptor, err := timeouts.New(&cfg.Server.Timeouts, logger, scope)
	if err != nil {
		logger.Fatal("could not create timeout interceptor", zap.Error(err))
	}
	interceptors = append(interceptors, timeoutInterceptor.UnaryInterceptor())

	// All other configured middleware.
	for name, factory := range cf.Middleware {
		logger := logger.With(zap.String("middlewareName", name))

		logger.Info("registering middleware")
		m, err := factory(cfg, logger, scope)
		if err != nil {
			logger.Fatal("middleware instantiation failed", zap.Error(err))
		}

		interceptors = append(interceptors, m.UnaryInterceptor())
	}

	// Instantiate and register modules listed in the configuration.
	rpcMux, err := mux.New(interceptors, assets, metricsHandler, cfg.Server)
	if err != nil {
		panic(err)
	}
	ctx := context.TODO()

	// Create a client connection for the registrar to make grpc-gateway's handlers available.
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if cfg.Server.MaxResponseSizeBytes > 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(int(cfg.Server.MaxResponseSizeBytes))))
	}
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", cfg.Server.Listener.Address, cfg.Server.Listener.Port), opts...)
	if err != nil {
		logger.Fatal("failed to bring up gRPC transport for grpc-gateway handlers", zap.Error(err))
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				logger.Warn("failed to close gRPC transport connection after err", zap.Error(err))
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				logger.Warn("failed to close gRPC transport connection when done", zap.Error(err))
			}
		}()
	}()

	reg := newRegistrar(ctx, rpcMux.JSONGateway, rpcMux.GRPCServer, conn)
	for name, factory := range cf.Endpoints {
		logger := logger.With(zap.String("handlerName", name))

		logger.Info("registering handler")
		h, err := factory(cfg, logger, scope.SubScope(name))
		if err != nil {
			logger.Fatal("handler instantiation failed", zap.Error(err))
		}

		if err := h.Register(reg); err != nil {
			logger.Fatal("registration to gateway failed", zap.Error(err))
		}
	}

	// Now that everything is registered, enable gRPC reflection.
	rpcMux.EnableGRPCReflection()

	// Save metadata on what RPCs being served for fast-lookup by internal services.
	//if err := meta.GenerateGRPCMetadata(rpcMux.GRPCServer); err != nil {
	//	logger.Fatal("reflection on grpc server failed", zap.Error(err))
	//}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Listener.Address, cfg.Server.Listener.Port)
	logger.Info("listening", zap.Namespace("tcp"), zap.String("addr", addr))

	// Figure out the maximum global timeout and set as a backstop (with 1s buffer).
	timeout := computeMaximumTimeout(&cfg.Server.Timeouts)
	if timeout > 0 {
		timeout += time.Second
	}

	// Start collecting go runtime stats if enabled
	if cfg.Server.Stats != nil && cfg.Server.Stats.GoRuntimeStats != nil {
		runtimeStats := stats.NewRuntimeStats(scope, cfg.Server.Stats.GoRuntimeStats)
		go runtimeStats.Collect(ctx)
	}

	srv := &http.Server{
		Handler:      mux.InsecureHandler(rpcMux),
		Addr:         addr,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(
		sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go func() {
		if err = srv.ListenAndServe(); !errors.Is(http.ErrServerClosed, err) {
			// Only log an error if it's not due to shutdown or close
			logger.Fatal("error bringing up listener", zap.Error(err))
		}
	}()

	<-sc

	signal.Stop(sc)

	// Shutdown timeout should be max request timeout (with 1s buffer).
	ctxShutDown, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err = srv.Shutdown(ctxShutDown); err != nil {
		logger.Fatal("server shutdown failed", zap.Error(err))
	}

	logger.Debug("server shutdown gracefully")
}

// Returns maximum timeout, where 0 is considered maximum (i.e. no timeout).
func computeMaximumTimeout(cfg *config.Timeouts) time.Duration {
	if cfg == nil {
		return timeouts.DefaultTimeout
	}

	ret := cfg.Default
	for _, e := range cfg.Overrides {
		override := e.Timeout
		if ret == 0 || override == 0 {
			return 0
		}

		if override > ret {
			ret = override
		}
	}

	return ret
}

func getStatsReporterConfiguration(cfg *config.Config, logger *zap.Logger) (tally.ScopeOptions, http.Handler) {
	var metricsHandler http.Handler
	var scopeOpts tally.ScopeOptions

	statsPrefix := "admiral_api"
	if cfg.Server.Stats.Prefix != "" {
		statsPrefix = cfg.Server.Stats.Prefix
	}

	switch cfg.Server.Stats.ReporterType {
	case config.ReporterTypeNull:
		scopeOpts = tally.ScopeOptions{
			Reporter: tally.NullStatsReporter,
		}
		return scopeOpts, nil
	case config.ReporterTypeLog:
		scopeOpts = tally.ScopeOptions{
			Reporter: stats.NewLogReporter(logger),
			Prefix:   statsPrefix,
		}
		return scopeOpts, nil
	case config.ReporterTypePrometheus:
		reporter, err := stats.NewPrometheusReporter()
		if err != nil {
			logger.Fatal("error creating prometheus reporter", zap.Error(err))
		}
		scopeOpts = tally.ScopeOptions{
			CachedReporter:  reporter,
			Prefix:          statsPrefix,
			SanitizeOptions: &tallyprom.DefaultSanitizerOpts,
		}
		metricsHandler = reporter.HTTPHandler()
		return scopeOpts, metricsHandler
	default:
		return tally.ScopeOptions{}, nil
	}
}
