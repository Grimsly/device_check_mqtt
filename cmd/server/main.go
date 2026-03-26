package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"device_check_mqtt/internal/config"
	"device_check_mqtt/internal/loader"
	"device_check_mqtt/internal/router"
	"device_check_mqtt/internal/service"
	"device_check_mqtt/internal/store"
)

func main() {
	// Load up any all config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	inMemoryStore := store.NewInMemoryStore()

	// NOTE: Remove loading devices on server statup in actual production. Instead create a seed script if devices need to be added.
	count, err := loader.LoadDevicesFromCSV(cfg.DevicesCSVFilepath, inMemoryStore)
	if err != nil {
		slog.Error("Failed to load devices", "error", err)
		os.Exit(1)
	}
	slog.Info("Devices loaded", "count", count)

	// Inject in-memory storage to the device service
	service := service.NewDeviceService(inMemoryStore)
	handler := router.NewRouter(service)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to receive server errors without calling os.Exit inside a goroutine
	serverErr := make(chan error, 1)

	go func() {
		slog.Info("Device check MQTT server starting", "port", cfg.Port)
		// Run the server and check for any errors
		// If the error returned is ErrServerClosed, then it means that the user manually shut it down
		// therefore, don't output an error
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	// Wait for a terminate signal or a server error
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	case <-terminate:
	}

	slog.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server stopped with errors", "error", err)
		os.Exit(1)
	}
	slog.Info("Server stopped")
}
