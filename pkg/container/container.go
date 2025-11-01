// pkg/container/container.go
package container

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service-refactor/internal/api/http/handler"
	"github.com/histopathai/main-service-refactor/internal/api/http/middleware"
	"github.com/histopathai/main-service-refactor/internal/api/http/router"
	"github.com/histopathai/main-service-refactor/internal/api/http/validator"
	eventhandlers "github.com/histopathai/main-service-refactor/internal/application/event_handlers"
	"github.com/histopathai/main-service-refactor/internal/application/orchestrator"
	"github.com/histopathai/main-service-refactor/internal/domain/events"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	infraEvents "github.com/histopathai/main-service-refactor/internal/infrastructure/events"
	"github.com/histopathai/main-service-refactor/internal/infrastructure/events/pubsub"
	firestoreRepo "github.com/histopathai/main-service-refactor/internal/infrastructure/storage/firestore"
	"github.com/histopathai/main-service-refactor/internal/infrastructure/storage/gcs"
	"github.com/histopathai/main-service-refactor/internal/service"
	"github.com/histopathai/main-service-refactor/pkg/config"
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
	WorkspaceService      *service.WorkspaceService
	PatientService        *service.PatientService
	ImageService          *service.ImageService
	AnnotationService     *service.AnnotationService
	AnnotationTypeService *service.AnnotationTypeService
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
		events.EventTypeImageProcessingRequested: c.Config.PubSub.ImageProcessingTopicID,
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
		c.Logger,
	)

	eventSerializer := events.NewJSONEventSerializer()
	c.ProcessResultHandler = eventhandlers.NewProcessResultHandler(
		c.Repos.ImageRepo,
		eventSerializer,
		c.EventPublisher,
	)

	c.Logger.Info("Event handlers initialized")
	return nil
}

func (c *Container) initOrchestrator() error {
	orchestratorConfig := orchestrator.SubscriberConfig{
		UploadStatusSubscriptionID:       c.Config.PubSub.UploadStatusSubscriptionID,
		ImageProcessResultSubscriptionID: c.Config.PubSub.ImageProcessResultSubscriptionID,
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
		c.AuthMiddleware,
		c.TimeoutMiddleware,
	)

	c.Logger.Info("HTTP layer initialized")
	return nil
}

func (c *Container) Close() error {
	c.Logger.Info("Closing container resources...")

	if c.FirestoreClient != nil {
		if err := c.FirestoreClient.Close(); err != nil {
			c.Logger.Error("Failed to close Firestore client", slog.String("error", err.Error()))
		}
	}

	if c.GCSClient != nil {
		if err := c.GCSClient.Close(); err != nil {
			c.Logger.Error("Failed to close GCS client", slog.String("error", err.Error()))
		}
	}

	if c.PubSubClient != nil {
		if err := c.PubSubClient.Stop(); err != nil {
			c.Logger.Error("Failed to close PubSub client", slog.String("error", err.Error()))
		}
	}

	if c.ImageOrchestrator != nil {
		if err := c.ImageOrchestrator.Stop(); err != nil {
			c.Logger.Error("Failed to stop orchestrator", slog.String("error", err.Error()))
		}
	}

	c.Logger.Info("Container resources closed")
	return nil
}
