package main

import (
	"context"
	"sync"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

type poller struct{}

func (p *poller) Start(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	logger := ctrl.Log.WithName("poller")
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutting down due to context cancellation")
			return
		case t := <-ticker.C:
			logger.Info("tick", "time", t.Format(time.RFC3339))
		}
	}
}

func (p *poller) Shutdown(ctx context.Context) error {
	ctrl.Log.WithName("poller").Info("shutdown called")
	return nil
}
