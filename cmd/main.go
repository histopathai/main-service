// cmd/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service-refactor/pkg/config"
	"github.com/histopathai/main-service-refactor/pkg/container"
	"github.com/histopathai/main-service-refactor/pkg/logger"
	"github.com/histopathai/main-service-refactor/pkg/seeder"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	loggerInstance := logger.New(&cfg.Logging)
	loggerInstance.Info("Starting main-service",
		slog.String("env", string(cfg.Env)),
		slog.String("version", getVersion()),
	)

	// Set Gin mode
	gin.SetMode(cfg.Server.GinMode)

	// Create application context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize dependency injection container
	app, err := container.New(ctx, cfg, loggerInstance.Logger)
	if err != nil {
		loggerInstance.Error("Failed to initialize application",
			slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer app.Close()

	// Run database seeder (only in local/dev environments)
	if cfg.IsLocal() || cfg.IsDevelopment() {
		if shouldSeed() {
			loggerInstance.Info("Running database seeder...")
			seederInstance := seeder.NewSeeder(app.Repos, loggerInstance.Logger)
			if err := seederInstance.Seed(ctx); err != nil {
				loggerInstance.Error("Failed to seed database",
					slog.String("error", err.Error()))
			}
		}
	}

	// Start event orchestrator
	if err := app.ImageOrchestrator.Start(ctx); err != nil {
		loggerInstance.Error("Failed to start image orchestrator",
			slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Setup HTTP router
	engine := app.Router.SetupRoutes()

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      engine,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start HTTP server in a goroutine
	go func() {
		loggerInstance.Info("Starting HTTP server",
			slog.String("addr", server.Addr))

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerInstance.Error("Failed to start HTTP server",
				slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	loggerInstance.Info("Shutting down server...")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown HTTP server
	if err := server.Shutdown(shutdownCtx); err != nil {
		loggerInstance.Error("Server forced to shutdown",
			slog.String("error", err.Error()))
	}

	// Stop orchestrator
	if err := app.ImageOrchestrator.Stop(); err != nil {
		loggerInstance.Error("Failed to stop orchestrator",
			slog.String("error", err.Error()))
	}

	loggerInstance.Info("Server exited successfully")
}

func getVersion() string {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		return "dev"
	}
	return version
}

func shouldSeed() bool {
	seed := os.Getenv("SEED_DATABASE")
	return seed == "true" || seed == "1"
}
