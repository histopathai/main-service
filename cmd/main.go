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

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/adapter"
	"github.com/histopathai/main-service/config"
	"github.com/histopathai/main-service/internal/handler"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/main-service/internal/router"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gin.SetMode(cfg.Server.GINMode)

	// Setup logger
	loglevel := slog.LevelInfo
	if cfg.Env == "LOCAL" {
		loglevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: loglevel,
	}))
	slog.SetDefault(logger)

	ctx := context.Background()

	// Initialize Firestore client
	firestoreClient, err := firestore.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer firestoreClient.Close()
	slog.Info("Firestore client initialized")

	storageClient, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Storage client: %v", err)
	}
	defer storageClient.Close()
	slog.Info("Storage client initialized")

	pubsubClient, err := pubsub.NewClient(ctx, cfg.ProjectID)
	if err != nil {
		log.Fatalf("Failed to create Pub/Sub client: %v", err)
	}
	defer pubsubClient.Close()
	slog.Info("Pub/Sub client initialized")

	//initialize adapters
	firestoreAdapter := adapter.NewFirestoreAdapter(firestoreClient)
	pubsubAdapter := adapter.NewGooglePubSubAdapter(pubsubClient)

	//initialize repositories
	mainRepo := repository.NewMainRepository(firestoreAdapter)
	messageBroker := repository.NewMessageBroker(pubsubAdapter)

	//initialize handlers
	uploadHandler := handler.NewUploadHandler(
		storageClient,
		cfg.OriginalBucketName,
		mainRepo,
		logger,
	)
	workspaceHandler := handler.NewWorkspaceHandler(
		mainRepo,
		logger,
	)

	patientHandler := handler.NewPatientHandler(
		mainRepo,
		logger,
	)

	UploadCompletionHandler := handler.NewUploadCompletionHandler(
		messageBroker,
		mainRepo,
		logger,
		cfg.MsgTopics.UploadStatusTopicID,
		cfg.MsgTopics.ImageProcessingTopicID,
	)

	UploadCompletionHandler.StartListening(ctx)
	// Setup router
	r := router.SetupRouter(&router.RouterConfig{
		UploadHandler:    uploadHandler,
		WorkspaceHandler: workspaceHandler,
		PatientHandler:   patientHandler,
	})

	//Create HTTP server
	src := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server", "port", cfg.Server.Port, "env", cfg.Env)
		if err := src.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down server...")

	// Gracefully shutdown the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := src.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	slog.Info("Server exiting")
}
