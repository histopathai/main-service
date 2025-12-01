package container

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/api/http/handler"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/api/http/router"
	"github.com/histopathai/main-service/internal/api/http/validator"
	eventhandlers "github.com/histopathai/main-service/internal/application/event_handlers"
	eventpublisher "github.com/histopathai/main-service/internal/application/event_publisher"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/infrastructure/events/pubsub"
	firestoreRepo "github.com/histopathai/main-service/internal/infrastructure/storage/firestore"
	"github.com/histopathai/main-service/internal/infrastructure/storage/gcs"
	"github.com/histopathai/main-service/internal/infrastructure/worker"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/pkg/config"
)

type Container struct {
	Config *config.Config
	Logger *slog.Logger

	//Infrastructure
	FirestoreClient *firestore.Client
	GCSClient       *gcs.GCSClient
	PubSubClient    *pubsub.GooglePubSubClient

	// Repositories
	Repos         *port.Repositories
	TelemetryRepo port.TelemetryRepository
	UOW           port.UnitOfWorkFactory
	Storage       port.ObjectStorage

	//Services
	WorkspaceService      port.IWorkspaceService
	PatientService        port.IPatientService
	ImageService          port.IImageService
	AnnotationService     port.IAnnotationService
	AnnotationTypeService port.IAnnotationTypeService
	TelemetryService      port.ITelemetryService

	// Event Publishers
	ImageEventPublisher     port.ImageEventPublisher
	TelemetryEventPublisher port.TelemetryEventPublisher

	// Worker
	ImageProcessingWorker port.ImageProcessingWorker

	// Event Handlers
	UploadStatusHandler           *eventhandlers.UploadStatusHandler
	ImageProcessingRequestHandler *eventhandlers.ImageProcessingRequestHandler
	ImageProcessingResultHandler  *eventhandlers.ImageProcessingResultHandler
	ImageDeletionHandler          *eventhandlers.ImageDeletionHandler
	TelemetryDLQHandler           *eventhandlers.TelemetryDLQHandler
	TelemetryErrorHandler         *eventhandlers.TelemetryErrorHandler

	// HTTP Layer
	Validator             *validator.RequestValidator
	WorkspaceHandler      *handler.WorkspaceHandler
	PatientHandler        *handler.PatientHandler
	ImageHandler          *handler.ImageHandler
	AnnotationHandler     *handler.AnnotationHandler
	AnnotationTypeHandler *handler.AnnotationTypeHandler
	// TelemetryHandler      *handler.TelemetryHandler
	AuthMiddleware    *middleware.AuthMiddleware
	TimeoutMiddleware *middleware.TimeoutMiddleware
	GCSProxyHandler   *handler.GCSProxyHandler
	Router            *router.Router
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Container, error) {
	c := &Container{
		Config: cfg,
		Logger: logger,
	}

	if err := c.initInfrastructure(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure : %w", err)
	}

	if err := c.initRepositories(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories : %w", err)
	}

	if err := c.initEventPublishers(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize event publishers : %w", err)
	}

	if err := c.initWorkers(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize workers : %w", err)
	}

	if err := c.initServices(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize services : %w", err)
	}

	if err := c.initEventHandlers(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize event handlers : %w", err)
	}

	if err := c.initSubscribers(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize subscribers : %w", err)
	}

	if err := c.initHTTPLayer(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP layer : %w", err)
	}

	return c, nil
}

func (c *Container) initInfrastructure(ctx context.Context) error {

	// Initialize Firestore Client
	firestoreClient, err := firestore.NewClient(ctx, c.Config.GCP.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to create firestore client: %w", err)
	}
	c.FirestoreClient = firestoreClient
	c.Logger.Info("Firestore client initialized")

	// Initialize GCS
	gcsClient, err := gcs.NewGCSClient(ctx, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}
	c.GCSClient = gcsClient
	c.Logger.Info("GCS client initialized")

	// Initialize PubSub
	pubsubClient, err := pubsub.NewGooglePubSubClient(ctx, c.Config.GCP.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to create pubsub client: %w", err)
	}
	c.PubSubClient = pubsubClient
	c.Logger.Info("PubSub client initialized")

	return nil
}

func (c *Container) initRepositories(ctx context.Context) error {
	uowFactory, allRepos := firestoreRepo.NewFirestoreUnitOfWorkFactory(c.FirestoreClient)
	c.UOW = uowFactory
	c.Repos = allRepos
	c.TelemetryRepo = firestoreRepo.NewTelemetryRepositoryImpl(c.FirestoreClient)
	c.Storage = c.GCSClient
	c.Logger.Info("Repositories initialized")
	return nil
}

func (c *Container) initEventPublishers(ctx context.Context) error {

	imageTopicMap := map[events.EventType]string{
		events.EventTypeImageProcessingRequested: c.Config.PubSub.ImageProcessingRequest.Topic.Name,
		events.EventTypeImageDeletionRequested:   c.Config.PubSub.ImageDeletion.Topic.Name,
	}

	c.ImageEventPublisher = eventpublisher.NewImageEventPublisher(
		c.PubSubClient,
		imageTopicMap,
	)

	telemetryTopicMap := map[events.EventType]string{
		events.EventTypeTelemetryDLQMessage: c.Config.PubSub.TelemetryDLQ.Topic.Name,
		events.EventTypeTelemetryError:      c.Config.PubSub.TelemetryError.Topic.Name,
	}

	c.TelemetryEventPublisher = eventpublisher.NewTelemetryEventPublisher(
		c.PubSubClient,
		telemetryTopicMap,
	)

	c.Logger.Info("Event publishers initialized successfully")
	return nil
}

func (c *Container) initEventHandlers(ctx context.Context) error {
	serializer := events.NewJSONEventSerializer()

	// Upload Status Handler
	c.UploadStatusHandler = eventhandlers.NewUploadStatusHandler(
		c.ImageService,
		c.Logger.WithGroup("upload_status_handler"),
		serializer,
		c.TelemetryEventPublisher,
	)

	// Image Processing Request Handler
	c.ImageProcessingRequestHandler = eventhandlers.NewImageProcessingRequestHandler(
		c.Repos.ImageRepo,
		c.ImageProcessingWorker,
		c.Storage,
		c.Config.GCP.OriginalBucketName,
		serializer,
		c.TelemetryEventPublisher,
		c.Logger.WithGroup("image_processing_request_handler"),
	)

	// Image Processing Result Handler
	c.ImageProcessingResultHandler = eventhandlers.NewImageProcessingResultHandler(
		c.Repos.ImageRepo,
		c.ImageEventPublisher,
		serializer,
		c.TelemetryEventPublisher,
		c.Logger.WithGroup("image_processing_result_handler"),
	)

	// Image Deletion Handler
	c.ImageDeletionHandler = eventhandlers.NewImageDeletionHandler(
		c.Repos.ImageRepo,
		c.GCSClient,
		c.Config.GCP.OriginalBucketName,
		c.Config.GCP.ProcessedBucketName,
		serializer,
		c.TelemetryEventPublisher,
		c.Logger.WithGroup("image_deletion_handler"),
	)

	// Telemetry DLQ Handler
	c.TelemetryDLQHandler = eventhandlers.NewTelemetryDLQHandler(
		c.TelemetryRepo,
		serializer,
		c.TelemetryEventPublisher,
		c.Logger.WithGroup("telemetry_dlq_handler"),
	)

	// Telemetry Error Handler
	c.TelemetryErrorHandler = eventhandlers.NewTelemetryErrorHandler(
		c.TelemetryRepo,
		serializer,
		c.TelemetryEventPublisher,
		c.Logger.WithGroup("telemetry_error_handler"),
	)

	c.Logger.Info("Event handlers initialized")
	return nil
}

func (c *Container) initServices(ctx context.Context) error {
	// Telemetry Service
	c.TelemetryService = service.NewTelemetryService(
		c.TelemetryRepo,
	)

	c.ImageService = service.NewImageService(
		c.Repos.ImageRepo,
		c.UOW,
		c.Storage,
		c.Config.GCP.OriginalBucketName,
		c.ImageEventPublisher,
		c.Repos.PatientRepo,
	)
	// Patient Service
	c.PatientService = service.NewPatientService(
		c.Repos.PatientRepo,
		c.Repos.WorkspaceRepo,
		c.Repos.ImageRepo,
		c.Repos.AnnotationRepo,
		c.ImageEventPublisher,
		c.UOW,
	)

	// Workspace Service
	c.WorkspaceService = service.NewWorkspaceService(
		c.Repos.WorkspaceRepo,
		c.Repos.PatientRepo,
		c.PatientService,
		c.UOW,
	)

	// Annotation Type Service
	c.AnnotationTypeService = service.NewAnnotationTypeService(
		c.Repos.AnnotationTypeRepo,
		c.UOW,
	)

	// Annotation Service
	c.AnnotationService = service.NewAnnotationService(
		c.Repos.AnnotationRepo,
		c.UOW,
	)

	c.Logger.Info("Services initialized")
	return nil
}

func (c *Container) initSubscribers(ctx context.Context) error {
	// Subscribe to Upload Status
	go func() {
		c.Logger.Info("Starting upload status subscriber",
			slog.String("subscription", c.Config.PubSub.UploadStatus.Name))

		if err := c.PubSubClient.Subscribe(
			ctx,
			c.Config.PubSub.UploadStatus.Name,
			c.UploadStatusHandler.Handle,
		); err != nil {
			c.Logger.Error("Upload status subscriber error",
				slog.String("error", err.Error()))
		}
	}()

	// Subscribe to Image Processing Requests
	go func() {
		c.Logger.Info("Starting image processing request subscriber",
			slog.String("subscription", c.Config.PubSub.ImageProcessingRequest.Subscription.Name))

		if err := c.PubSubClient.Subscribe(
			ctx,
			c.Config.PubSub.ImageProcessingRequest.Subscription.Name,
			c.ImageProcessingRequestHandler.Handle,
		); err != nil {
			c.Logger.Error("Image processing request subscriber error",
				slog.String("error", err.Error()))
		}
	}()

	// Subscribe to Image Processing Results
	go func() {
		c.Logger.Info("Starting image processing result subscriber",
			slog.String("subscription", c.Config.PubSub.ImageProcessingResult.Subscription.Name))

		if err := c.PubSubClient.Subscribe(
			ctx,
			c.Config.PubSub.ImageProcessingResult.Subscription.Name,
			c.ImageProcessingResultHandler.Handle,
		); err != nil {
			c.Logger.Error("Image processing result subscriber error",
				slog.String("error", err.Error()))
		}
	}()

	// Subscribe to Image Deletion
	go func() {
		c.Logger.Info("Starting image deletion subscriber",
			slog.String("subscription", c.Config.PubSub.ImageDeletion.Subscription.Name))

		if err := c.PubSubClient.Subscribe(
			ctx,
			c.Config.PubSub.ImageDeletion.Subscription.Name,
			c.ImageDeletionHandler.Handle,
		); err != nil {
			c.Logger.Error("Image deletion subscriber error",
				slog.String("error", err.Error()))
		}
	}()

	// Subscribe to Telemetry DLQ
	go func() {
		c.Logger.Info("Starting telemetry DLQ subscriber",
			slog.String("subscription", c.Config.PubSub.TelemetryDLQ.Subscription.Name))

		if err := c.PubSubClient.Subscribe(
			ctx,
			c.Config.PubSub.TelemetryDLQ.Subscription.Name,
			c.TelemetryDLQHandler.Handle,
		); err != nil {
			c.Logger.Error("Telemetry DLQ subscriber error",
				slog.String("error", err.Error()))
		}
	}()

	// Subscribe to Telemetry Errors
	go func() {
		c.Logger.Info("Starting telemetry error subscriber",
			slog.String("subscription", c.Config.PubSub.TelemetryError.Subscription.Name))

		if err := c.PubSubClient.Subscribe(
			ctx,
			c.Config.PubSub.TelemetryError.Subscription.Name,
			c.TelemetryErrorHandler.Handle,
		); err != nil {
			c.Logger.Error("Telemetry error subscriber error",
				slog.String("error", err.Error()))
		}
	}()

	c.Logger.Info("All subscribers started")
	return nil
}

func (c *Container) initWorkers(ctx context.Context) error {
	worker, err := worker.NewCloudRunWorker(ctx, c.Config.Worker, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to create Cloud Run worker: %w", err)
	}
	c.ImageProcessingWorker = worker
	c.Logger.Info("Workers initialized")
	return nil
}

func (c *Container) initHTTPLayer(ctx context.Context) error {
	// Validator
	c.Validator = validator.NewRequestValidator()

	// Handlers
	c.WorkspaceHandler = handler.NewWorkspaceHandler(
		c.WorkspaceService,
		c.Validator,
		c.Logger,
	)

	c.PatientHandler = handler.NewPatientHandler(
		c.PatientService,
		c.Validator,
		c.Logger,
	)

	c.ImageHandler = handler.NewImageHandler(
		c.ImageService,
		c.Validator,
		c.Logger,
	)

	c.AnnotationHandler = handler.NewAnnotationHandler(
		c.AnnotationService,
		c.Validator,
		c.Logger,
	)

	c.AnnotationTypeHandler = handler.NewAnnotationTypeHandler(
		c.AnnotationTypeService,
		c.Validator,
		c.Logger,
	)

	// Middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.Logger)
	c.TimeoutMiddleware = middleware.NewTimeoutMiddleware(
		30*time.Second,
		c.Logger,
	)

	gcsProxyhandler, err := handler.NewGCSProxyHandler(
		c.Config.GCP.ProjectID,
		c.Config.GCP.ProcessedBucketName,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create GCS proxy handler: %w", err)
	}
	c.GCSProxyHandler = gcsProxyhandler

	// Router
	routerConfig := &router.RouterConfig{
		Logger:         c.Logger,
		RequestTimeout: 30 * time.Second,
	}

	c.Router = router.NewRouter(
		routerConfig,
		c.WorkspaceHandler,
		c.PatientHandler,
		c.ImageHandler,
		c.AnnotationHandler,
		c.AnnotationTypeHandler,
		c.GCSProxyHandler,
		c.AuthMiddleware,
		c.TimeoutMiddleware,
	)

	c.Logger.Info("HTTP layer initialized")
	return nil
}

func (c *Container) Close() error {
	c.Logger.Info("Closing container resources...")

	var errs []error

	// Stop all subscribers
	if c.PubSubClient != nil {
		c.Logger.Info("Stopping Pub/Sub client")
		if err := c.PubSubClient.Stop(); err != nil {
			c.Logger.Error("Error stopping Pub/Sub client",
				slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("pubsub stop: %w", err))
		}
	}

	// Close other resources
	if c.FirestoreClient != nil {
		if err := c.FirestoreClient.Close(); err != nil {
			c.Logger.Error("Failed to close Firestore client",
				slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("firestore close: %w", err))
		}
	}

	if c.GCSClient != nil {
		if err := c.GCSClient.Close(); err != nil {
			c.Logger.Error("Failed to close GCS client",
				slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("gcs close: %w", err))
		}
	}

	if len(errs) > 0 {
		c.Logger.Error("Container closed with errors",
			slog.Int("error_count", len(errs)))
		return fmt.Errorf("container close errors: %v", errs)
	}

	c.Logger.Info("Container resources closed successfully")
	return nil
}
