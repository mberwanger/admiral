package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Loop stopped: shutting down")
				return
			case t := <-ticker.C:
				fmt.Println("Tick:", t.Format(time.RFC3339))
			}
		}
	}()

	<-sigChan
	fmt.Println("\nReceived Ctrl+C, initiating graceful shutdown...")
	cancel()

	time.Sleep(1 * time.Second)
	fmt.Println("Program exited gracefully")
}
