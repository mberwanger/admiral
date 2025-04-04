package timeouts

import (
	"context"
	"fmt"
	"time"

	"github.com/uber-go/tally/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.admiral.io/admiral/internal/config"
	"go.admiral.io/admiral/internal/middleware"
)

const (
	DefaultTimeout = time.Second * 15

	// Boost is added to the timeout to give a handler that respects the deadline the opportunity to return an error.
	boost = time.Millisecond * 50
)

func New(cfg *config.Timeouts, logger *zap.Logger, scope tally.Scope) (middleware.Middleware, error) { //
	if cfg == nil {
		cfg = &config.Timeouts{Default: DefaultTimeout}
	}

	defaultTimeout := cfg.Default

	overrides := make(map[string]time.Duration, len(cfg.Overrides))
	for _, entry := range cfg.Overrides {
		overrides[join(entry.Service, entry.Method)] = entry.Timeout
	}

	return &mid{
		logger:         logger,
		scope:          scope,
		defaultTimeout: defaultTimeout,
		overrides:      overrides,
	}, nil
}

type mid struct {
	logger         *zap.Logger
	scope          tally.Scope
	defaultTimeout time.Duration
	overrides      map[string]time.Duration
}

func (m *mid) getDuration(service, method string) time.Duration {
	if override, ok := m.overrides[join(service, method)]; ok {
		return override
	}
	return m.defaultTimeout
}

type unaryHandlerReturn struct {
	resp interface{}
	err  error
}

func (m *mid) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		service, method, ok := middleware.SplitFullMethod(info.FullMethod)
		if !ok {
			m.logger.Warn("could not parse gRPC method", zap.String("fullMethod", info.FullMethod))
		}

		// Create a return channel for the goroutine.
		resultChan := make(chan unaryHandlerReturn)
		defer close(resultChan)

		// Compute timeout, and if not infinite set up timer and context.
		timeout := m.getDuration(service, method)
		done := make(chan struct{}) // Channel to track when timeout error has been returned and return channel closed.
		if timeout != 0 {
			// Set-up a context with timeout.
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()

			// If timeout is not infinite, return after timeout plus boost. Boost gives the goroutine a chance to return
			// if it's respecting the deadline.
			timer := time.AfterFunc(timeout+boost, func() { close(done) })
			defer timer.Stop() // Channel will still be garbage collected if close never occurs.
		} else {
			defer close(done)
		}

		// Spawn the handler in a goroutine so we can return early on timeout if it doesn't complete.
		go func() {
			resp, err := handler(ctx, req)
			select {
			case <-done:
				m.logger.Error(
					"handler completed after timeout",
					zap.String("service", service),
					zap.String("method", method),
					zap.Error(err))
			default:
				resultChan <- unaryHandlerReturn{resp: resp, err: err}
			}
		}()

		// Wait for timeout or handler to send result.
		select {
		case ret := <-resultChan:
			return ret.resp, ret.err
		case <-done:
			return nil, status.New(codes.DeadlineExceeded, "timeout exceeded").Err()
		}
	}
}

func join(service, method string) string {
	const pattern = "/%s/%s"
	return fmt.Sprintf(pattern, service, method)
}
