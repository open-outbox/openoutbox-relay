// Package main is the entry point for the OpenOutbox Relay.
// It sets up signal handling, builds the dependency injection container,
// and manages the lifecycle of the relay engine and the HTTP server.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/open-outbox/relay/internal/container"
	"github.com/open-outbox/relay/internal/relay"
	"go.uber.org/zap"
)

// main initializes the application and starts the main execution loop.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := container.BuildContainer(ctx)

	err := c.Invoke(func(engine *relay.Engine, api *relay.Server, logger *zap.Logger) error {
		defer func() { _ = logger.Sync() }()

		logger.Info("Starting OpenOutbox Relay")
		api.Start()
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := api.Stop(shutdownCtx); err != nil {
				logger.Error("API shutdown error", zap.Error(err))
			}
		}()

		if err := engine.Start(ctx); err != nil && err != context.Canceled {
			return err
		}

		logger.Info("Relay process exited gracefully")
		return nil
	})

	if err != nil {
		log.Fatalf("Relay terminated with error: %v", err)
	}
}
