package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGKILL,
		syscall.SIGABRT, os.Interrupt)
	defer cancel()

	log := newLogger()

	// We need a running geoclue agent
	isRunning, err := geoClueAgentIsRunning(ctx)
	if err != nil {
		log.Error("failed to check if geoclue agent is running", logError(err))
		os.Exit(1)
	}
	if !isRunning {
		log.Error("required geoclue agent is not running, shutting down")
		os.Exit(1)
	}

	// Initialize the service
	service, err := New()
	if err != nil {
		log.Error("failed to initialize waybar-weather service", logError(err))
		os.Exit(1)
	}

	// Start the service loop
	log.Info("starting waybar-weather service")
	if err = service.Run(ctx); err != nil {
		log.Error("failed to start waybar-weather service", logError(err))
	}
	log.Info("shutting down waybar-weather service")
}
