package orchestrator

import (
	"context"
	"log/slog"
	"sync"

	eventhandlers "github.com/histopathai/main-service/internal/application/event_handlers"
	"github.com/histopathai/main-service/internal/domain/events"
	infraEvents "github.com/histopathai/main-service/internal/infrastructure/events"
)

type SubscriberConfig struct {
	UploadStatusSubscriptionID       string
	ImageProcessResultSubscriptionID string
}

type ImageOrchestrator struct {
	pubsubClient   events.Subscriber
	registry       *infraEvents.EventRegistry
	uploadHandler  *eventhandlers.UploadStatusHandler
	processHandler *eventhandlers.ProcessResultHandler
	config         SubscriberConfig
	logger         *slog.Logger
	cancelFuncs    []context.CancelFunc
	wg             sync.WaitGroup
}

func NewImageOrchestrator(
	pubsubClient events.Subscriber,
	registry *infraEvents.EventRegistry,
	uploadHandler *eventhandlers.UploadStatusHandler,
	processHandler *eventhandlers.ProcessResultHandler,
	config SubscriberConfig,
	logger *slog.Logger,
) *ImageOrchestrator {

	registry.Register(events.EventTypeImageUploaded, uploadHandler.Handle)
	registry.Register(events.EventTypeImageProcessingCompleted, processHandler.Handle)
	registry.Register(events.EventTypeImageProcessingFailed, processHandler.Handle)

	return &ImageOrchestrator{
		pubsubClient:   pubsubClient,
		registry:       registry,
		uploadHandler:  uploadHandler,
		processHandler: processHandler,
		config:         config,
		logger:         logger,
		cancelFuncs:    make([]context.CancelFunc, 0),
	}
}

func (io *ImageOrchestrator) Start(ctx context.Context) error {
	io.logger.Info("Starting Image Orchestrator")

	if io.config.UploadStatusSubscriptionID != "" {
		if err := io.startSubscription(ctx, io.config.UploadStatusSubscriptionID, "upload-status"); err != nil {
			return err
		}
	}
	if io.config.ImageProcessResultSubscriptionID != "" {
		if err := io.startSubscription(ctx, io.config.ImageProcessResultSubscriptionID, "image-process-result"); err != nil {
			return err
		}
	}

	io.logger.Info("Image Orchestrator started successfully")
	return nil
}

func (io *ImageOrchestrator) startSubscription(ctx context.Context, subscriptionID, name string) error {
	subCtx, cancel := context.WithCancel(ctx)
	io.cancelFuncs = append(io.cancelFuncs, cancel)

	io.wg.Add(1)
	go func() {
		defer io.wg.Done()

		io.logger.Info("Starting subscription", "name", name, "subscriptionID", subscriptionID)

		if err := io.pubsubClient.Subscribe(subCtx, subscriptionID, io.registry.Handle); err != nil {
			if subCtx.Err() != context.Canceled {
				io.logger.Error("Subscription error",
					"name", name,
					"subscriptionID", subscriptionID,
					"error", err)
			}
		}

		io.logger.Info("Subscription stopped", "name", name, "subscriptionID", subscriptionID)
	}()

	return nil
}

func (io *ImageOrchestrator) Stop() error {
	io.logger.Info("Stopping Image Orchestrator")

	for _, cancel := range io.cancelFuncs {
		cancel()
	}

	io.wg.Wait()
	io.logger.Info("Image Orchestrator stopped")

	if err := io.pubsubClient.Stop(); err != nil {
		io.logger.Error("Error stopping Pub/Sub client", slog.String("error", err.Error()))
	}

	io.logger.Info("Pub/Sub client stopped")
	return nil
}
