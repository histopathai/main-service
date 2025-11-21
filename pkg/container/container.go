// pkg/container/container.go
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
	"github.com/histopathai/main-service/internal/application/orchestrator"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/repository"
	infraEvents "github.com/histopathai/main-service/internal/infrastructure/events"
	"github.com/histopathai/main-service/internal/infrastructure/events/pubsub"
	firestoreRepo "github.com/histopathai/main-service/internal/infrastructure/storage/firestore"
	"github.com/histopathai/main-service/internal/infrastructure/storage/gcs"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/pkg/config"
)

type Container struct {
	Config *config.Config
	Logger *slog.Logger

	// Infrastructure
	FirestoreClient *firestore.Client
	GCSClient       *gcs.GCSClient
	PubSubClient    *pubsub.GooglePubSubClient

	// Repositories
	Repos *repository.Repositories
	UOW   repository.UnitOfWorkFactory

	// Services
	WorkspaceService      service.IWorkspaceService
	PatientService        service.IPatientService
	ImageService          service.IImageService
	AnnotationService     service.IAnnotationService
	AnnotationTypeService service.IAnnotationTypeService
	EventPublisher        events.ImageEventPublisher

	// Event Handlers
	UploadStatusHandler  *eventhandlers.UploadStatusHandler
	ProcessResultHandler *eventhandlers.ProcessResultHandler
	EventRegistry        *infraEvents.EventRegistry

	// Orchestrator
	ImageOrchestrator *orchestrator.ImageOrchestrator

	// HTTP Layer
	Validator             *validator.RequestValidator
	WorkspaceHandler      *handler.WorkspaceHandler
	PatientHandler        *handler.PatientHandler
	ImageHandler          *handler.ImageHandler
	AnnotationHandler     *handler.AnnotationHandler
	AnnotationTypeHandler *handler.AnnotationTypeHandler
	AuthMiddleware        *middleware.AuthMiddleware
	TimeoutMiddleware     *middleware.TimeoutMiddleware
	GCSProxyHandler       *handler.GCSProxyHandler
	Router                *router.Router
}

func New(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Container, error) {
	c := &Container{
		Config: cfg,
		Logger: logger,
	}

	if err := c.initInfrastructure(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize infrastructure: %w", err)
	}

	if err := c.initRepositories(); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if err := c.initServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	if err := c.initEventHandlers(); err != nil {
		return nil, fmt.Errorf("failed to initialize event handlers: %w", err)
	}

	if err := c.initOrchestrator(); err != nil {
		return nil, fmt.Errorf("failed to initialize orchestrator: %w", err)
	}

	if err := c.initHTTPLayer(); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP layer: %w", err)
	}

	return c, nil
}

func (c *Container) initInfrastructure(ctx context.Context) error {
	// Initialize Firestore
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

func (c *Container) initRepositories() error {
	uowFactory, allRepos := firestoreRepo.NewFirestoreUnitOfWorkFactory(c.FirestoreClient)
	c.UOW = uowFactory
	c.Repos = allRepos
	c.Logger.Info("Repositories initialized")
	return nil
}

func (c *Container) initServices() error {
	// Create event publisher
	topicMap := map[events.EventType]string{
		events.EventTypeImageProcessingRequested: c.Config.PubSub.ImageProcessingTopic,
	}
	c.EventPublisher = service.NewEventPublisher(c.PubSubClient, topicMap)

	// Create services
	c.WorkspaceService = service.NewWorkspaceService(
		c.Repos.WorkspaceRepo,
		c.UOW,
	)

	c.PatientService = service.NewPatientService(
		c.Repos.PatientRepo,
		c.Repos.WorkspaceRepo,
		c.UOW,
	)

	c.ImageService = service.NewImageService(
		c.Repos.ImageRepo,
		c.UOW,
		c.GCSClient,
		c.Config.GCP.OriginalBucketName,
		c.EventPublisher,
		c.Repos.PatientRepo,
	)

	c.AnnotationService = service.NewAnnotationService(
		c.Repos.AnnotationRepo,
		c.UOW,
	)

	c.AnnotationTypeService = service.NewAnnotationTypeService(
		c.Repos.AnnotationTypeRepo,
		c.UOW,
	)

	c.Logger.Info("Services initialized")
	return nil
}

func (c *Container) initEventHandlers() error {
	c.EventRegistry = infraEvents.NewEventRegistry()

	c.UploadStatusHandler = eventhandlers.NewUploadStatusHandler(
		c.ImageService,
		c.Logger.WithGroup("upload_status_handler"),
	)

	eventSerializer := events.NewJSONEventSerializer()
	c.ProcessResultHandler = eventhandlers.NewProcessResultHandler(
		c.Repos.ImageRepo,
		eventSerializer,
		c.EventPublisher,
		c.Logger.WithGroup("process_result_handler"),
	)

	c.Logger.Info("Event handlers initialized")
	return nil
}

func (c *Container) initOrchestrator() error {
	orchestratorConfig := orchestrator.SubscriberConfig{
		UploadStatusSubscriptionID:       c.Config.PubSub.UploadStatusSubscription,
		ImageProcessResultSubscriptionID: c.Config.PubSub.ProcessingCompletedSubscription,
	}

	c.ImageOrchestrator = orchestrator.NewImageOrchestrator(
		c.PubSubClient,
		c.EventRegistry,
		c.UploadStatusHandler,
		c.ProcessResultHandler,
		orchestratorConfig,
		c.Logger,
	)

	c.Logger.Info("Orchestrator initialized")
	return nil
}

func (c *Container) initHTTPLayer() error {
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

	if c.ImageOrchestrator != nil {
		c.Logger.Info("Stopping Image Orchestrator")
		if err := c.ImageOrchestrator.Stop(); err != nil {
			c.Logger.Error("Error stopping orchestrator", slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("orchestrator stop: %w", err))
		}
	}

	if c.PubSubClient != nil {
		c.Logger.Info("Stopping Pub/Sub client")
		if err := c.PubSubClient.Stop(); err != nil {
			c.Logger.Error("Error stopping Pub/Sub client", slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("pubsub stop: %w", err))
		}
	}

	if c.FirestoreClient != nil {
		if err := c.FirestoreClient.Close(); err != nil {
			c.Logger.Error("Failed to close Firestore client", slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("firestore close: %w", err))
		}
	}

	if c.GCSClient != nil {
		if err := c.GCSClient.Close(); err != nil {
			c.Logger.Error("Failed to close GCS client", slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("gcs close: %w", err))
		}
	}

	if len(errs) > 0 {
		c.Logger.Error("Container closed with errors", slog.Int("error_count", len(errs)))
		return fmt.Errorf("container close errors: %v", errs)
	}

	c.Logger.Info("Container resources closed successfully")
	return nil
}
