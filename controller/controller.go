package main

import (
	"context"
	"sync"

	ctrl "sigs.k8s.io/controller-runtime"
)

type controller struct {
	mgr ctrl.Manager
}

func (c *controller) Start(ctx context.Context, wg *sync.WaitGroup, errCh chan<- error) {
	defer wg.Done()

	logger := ctrl.Log.WithName("controller")
	var err error
	c.mgr, err = ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		errCh <- err
		return
	}

	logger.Info("starting manager")
	if err := c.mgr.Start(ctx); err != nil {
		errCh <- err
	}
	logger.Info("manager stopped")
}

func (c *controller) Shutdown(ctx context.Context) error {
	ctrl.Log.WithName("controller").Info("shutdown called")
	return nil
}
