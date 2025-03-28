package main

import (
	"context"
	"crypto/tls"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	admiral "github.com/mberwanger/admiral/client"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
}

type service interface {
	Start(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error)
	Shutdown(ctx context.Context) error
}

func main() {
	var (
		metricsAddr                                      string
		metricsCertPath, metricsCertName, metricsCertKey string
		enableLeaderElection                             bool
		probeAddr                                        string
		secureMetrics                                    bool
		enableHTTP2                                      bool
		tlsOpts                                          []func(*tls.Config)
	)

	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true, "If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	flag.StringVar(&metricsCertPath, "metrics-cert-path", "", "The directory that contains the metrics server certificate.")
	flag.StringVar(&metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	flag.StringVar(&metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	flag.BoolVar(&enableHTTP2, "enable-http2", false, "If set, HTTP/2 will be enabled for the metrics and webhook servers")

	// Setup logging
	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Setup admiral client placeholder
	_, _ = admiral.New(context.Background(), admiral.Config{})

	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}

	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	// Create a cancellable context for the application lifetime
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancellation happens if main exits unexpectedly

	// Set up signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Define services to run
	services := []service{
		&poller{},
		&controller{},
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(services))

	// Start all services in goroutines
	for i, svc := range services {
		wg.Add(1)
		go svc.Start(ctx, &wg, errCh)
		setupLog.Info("started service", "index", i)
	}

	// Handle shutdown in a separate goroutine
	go func() {
		// Wait for a termination signal
		sig := <-sigCh
		setupLog.Info("received signal, initiating graceful shutdown", "signal", sig)

		// Cancel the main context to signal all services to stop
		cancel()

		// Create a timeout context for shutdown (e.g., 5 seconds)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// Call Shutdown on each service with the timeout context
		for i, svc := range services {
			if err := svc.Shutdown(shutdownCtx); err != nil {
				setupLog.Error(err, "failed to shutdown service", "index", i)
			}
		}
	}()

	// Monitor for startup errors
	select {
	case err := <-errCh:
		setupLog.Error(err, "service failed to start, triggering shutdown")
		cancel() // Trigger shutdown if a service fails
	case <-ctx.Done():
		// Context was cancelled (e.g., by signal handler)
	}

	// Wait for all services to finish
	wg.Wait()
	setupLog.Info("graceful shutdown complete, all services stopped")
}
