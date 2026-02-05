package container

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	inmemorycache "github.com/histopathai/main-service/internal/adapter/cache"
	"github.com/histopathai/main-service/internal/adapter/events/pubsub"
	firestorerepo "github.com/histopathai/main-service/internal/adapter/repository/firestore"
	"github.com/histopathai/main-service/internal/adapter/storage/gcs"
	"github.com/histopathai/main-service/internal/adapter/worker"
	"github.com/histopathai/main-service/internal/api/http/handler"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/api/http/router"
	apphandler "github.com/histopathai/main-service/internal/application/handler"
	"github.com/histopathai/main-service/internal/application/proxy"
	appquery "github.com/histopathai/main-service/internal/application/queries"
	appusecase "github.com/histopathai/main-service/internal/application/usecase"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/port/cache"
	portevent "github.com/histopathai/main-service/internal/port/event"
	"github.com/histopathai/main-service/pkg/config"
)

type Container struct {
	Config *config.Config
	Logger *slog.Logger

	// Infrastructure
	FirestoreClient *firestore.Client
	StorageClient   *storage.Client
	Cache           cache.Cache

	// Repositories
	WorkspaceRepo      port.WorkspaceRepository
	PatientRepo        port.PatientRepository
	ImageRepo          port.ImageRepository
	ContentRepo        port.ContentRepository
	AnnotationRepo     port.AnnotationRepository
	AnnotationTypeRepo port.AnnotationTypeRepository
	UOW                port.UnitOfWorkFactory
	TileServer         *proxy.TileServer

	// Storages
	OriginStorage    port.Storage
	ProcessedStorage port.Storage

	// Use Cases
	WorkspaceUseCase      port.WorkspaceUseCase
	PatientUseCase        port.PatientUseCase
	ImageUseCase          port.ImageUseCase
	AnnotationUseCase     port.AnnotationUseCase
	AnnotationTypeUseCase port.AnnotationTypeUseCase

	// Queries
	WorkspaceQuery      port.WorkspaceQuery
	PatientQuery        port.PatientQuery
	ImageQuery          port.ImageQuery
	ContentQuery        port.ContentQuery
	AnnotationQuery     port.AnnotationQuery
	AnnotationTypeQuery port.AnnotationTypeQuery

	// Event Infrastructure
	EventPublisher     portevent.EventPublisher
	UploadSubscriber   portevent.EventSubscriber
	ProcessSubscriber  portevent.EventSubscriber
	CompleteSubscriber portevent.EventSubscriber

	// Event Handlers
	NewFileHandler              *apphandler.NewFileHandler
	ImageProcessHandler         *apphandler.ImageProcessHandler
	ImageProcessCompleteHandler *apphandler.ImageProcessCompleteHandler

	// Worker
	ImageProcessingWorker port.ImageProcessingWorker

	// HTTP Layer
	WorkspaceHandler      *handler.WorkspaceHandler
	PatientHandler        *handler.PatientHandler
	ImageHandler          *handler.ImageHandler
	AnnotationHandler     *handler.AnnotationHandler
	AnnotationTypeHandler *handler.AnnotationTypeHandler
	AuthMiddleware        *middleware.AuthMiddleware
	TimeoutMiddleware     *middleware.TimeoutMiddleware
	TileProxyHandler      *handler.TileProxyHandler
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

	if err := c.initRepositories(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize repositories: %w", err)
	}

	if err := c.initStorages(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize storages: %w", err)
	}

	if err := c.initUseCases(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize use cases: %w", err)
	}

	if err := c.initQueries(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize queries: %w", err)
	}

	if err := c.initEventInfrastructure(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize event infrastructure: %w", err)
	}

	if err := c.initWorkers(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize workers: %w", err)
	}

	if err := c.initEventHandlers(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize event handlers: %w", err)
	}

	if err := c.initSubscribers(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize subscribers: %w", err)
	}

	if err := c.initProxies(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize proxies: %w", err)
	}

	if err := c.initHTTPLayer(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP layer: %w", err)
	}

	return c, nil
}

func (c *Container) initInfrastructure(ctx context.Context) error {
	// Initialize Firestore Client
	var firestoreClient *firestore.Client
	var err error

	// Use configured database name (supports dev environment isolation)
	if c.Config.GCP.FirestoreDatabase != "" && c.Config.GCP.FirestoreDatabase != "(default)" {
		firestoreClient, err = firestore.NewClientWithDatabase(ctx, c.Config.GCP.ProjectID, c.Config.GCP.FirestoreDatabase)
	} else {
		firestoreClient, err = firestore.NewClient(ctx, c.Config.GCP.ProjectID)
	}

	if err != nil {
		return fmt.Errorf("failed to create firestore client: %w", err)
	}
	c.FirestoreClient = firestoreClient
	c.Logger.Info("Firestore client initialized", "database", c.Config.GCP.FirestoreDatabase)

	// Initialize GCSs
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}
	c.StorageClient = client
	c.Logger.Info("GCS client initialized")

	// Initialize Cache
	c.Cache = inmemorycache.NewMemoryCache(time.Minute * 10)
	c.Logger.Info("Cache initialized")
	return nil
}

func (c *Container) initRepositories(ctx context.Context) error {
	uowFactory := firestorerepo.NewFirestoreUnitOfWorkFactory(c.FirestoreClient)

	c.UOW = uowFactory
	c.WorkspaceRepo = uowFactory.GetWorkspaceRepo()
	c.PatientRepo = uowFactory.GetPatientRepo()
	c.ImageRepo = uowFactory.GetImageRepo()
	c.ContentRepo = uowFactory.GetContentRepo()
	c.AnnotationRepo = uowFactory.GetAnnotationRepo()
	c.AnnotationTypeRepo = uowFactory.GetAnnotationTypeRepo()
	c.Logger.Info("Repositories initialized")
	return nil
}

func (c *Container) initStorages(ctx context.Context) error {
	originStorage := gcs.NewGCSAdapter(c.StorageClient, c.Config.GCP.OriginalBucketName, c.Logger)
	processedStorage := gcs.NewGCSAdapter(c.StorageClient, c.Config.GCP.ProcessedBucketName, c.Logger)
	c.OriginStorage = originStorage
	c.ProcessedStorage = processedStorage
	c.Logger.Info("Storages initialized")
	return nil
}

func (c *Container) initUseCases(ctx context.Context) error {
	c.WorkspaceUseCase = appusecase.NewWorkspaceUseCase(c.WorkspaceRepo, c.UOW)
	c.PatientUseCase = appusecase.NewPatientUseCase(c.PatientRepo, c.UOW)
	c.ImageUseCase = appusecase.NewImageUseCase(c.ImageRepo, c.UOW, c.OriginStorage, c.ProcessedStorage)
	c.AnnotationUseCase = appusecase.NewAnnotationUseCase(c.AnnotationRepo, c.UOW)
	c.AnnotationTypeUseCase = appusecase.NewAnnotationTypeUseCase(c.AnnotationTypeRepo, c.UOW)
	c.Logger.Info("Use cases initialized")
	return nil
}

func (c *Container) initQueries(ctx context.Context) error {
	c.WorkspaceQuery = appquery.NewWorkspaceQuery(c.WorkspaceRepo)
	c.PatientQuery = appquery.NewPatientQuery(c.PatientRepo)
	c.ImageQuery = appquery.NewImageQuery(c.ImageRepo)
	c.ContentQuery = appquery.NewContentQuery(c.ContentRepo)
	c.AnnotationQuery = appquery.NewAnnotationQuery(c.AnnotationRepo)
	c.AnnotationTypeQuery = appquery.NewAnnotationTypeQuery(c.AnnotationTypeRepo)
	c.Logger.Info("Queries initialized")
	return nil
}

func (c *Container) initEventInfrastructure(ctx context.Context) error {

	// Topic mappings for publishers
	topicMapping := map[domainevent.EventType]string{
		domainevent.NewFileExistEventType:         c.Config.PubSub.UploadStatus.Topic,
		domainevent.ImageProcessReqEventType:      c.Config.PubSub.ImageProcessingRequest.Topic.Name,
		domainevent.ImageProcessCompleteEventType: c.Config.PubSub.ImageProcessingResult.Topic.Name,
	}

	// Create main event publisher
	publisher, err := pubsub.NewPubSubPublisher(ctx, c.Config.GCP.ProjectID, topicMapping)
	if err != nil {
		return fmt.Errorf("failed to create event publisher: %w", err)
	}
	c.EventPublisher = publisher

	// Create subscribers
	uploadSub, err := pubsub.NewPubSubSubscriber(
		ctx,
		c.Config.GCP.ProjectID,
		c.Config.PubSub.UploadStatus.Name,
		nil, // handler set later,
		c.Cache,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create upload subscriber: %w", err)
	}
	c.UploadSubscriber = uploadSub

	processSub, err := pubsub.NewPubSubSubscriber(
		ctx,
		c.Config.GCP.ProjectID,
		c.Config.PubSub.ImageProcessingRequest.Subscription.Name,
		nil, // handler set later,
		c.Cache,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create process subscriber: %w", err)
	}
	c.ProcessSubscriber = processSub

	completeSub, err := pubsub.NewPubSubSubscriber(
		ctx,
		c.Config.GCP.ProjectID,
		c.Config.PubSub.ImageProcessingResult.Subscription.Name,
		nil, // handler set later,
		c.Cache,
		c.Logger,
	)
	if err != nil {
		return fmt.Errorf("failed to create complete subscriber: %w", err)
	}
	c.CompleteSubscriber = completeSub

	c.Logger.Info("Event infrastructure initialized")
	return nil
}

func (c *Container) initWorkers(ctx context.Context) error {
	worker, err := worker.NewCloudRunWorker(ctx, c.Config.Worker, c.Config.GCP, c.Logger)
	if err != nil {
		return fmt.Errorf("failed to create Cloud Run worker: %w", err)
	}
	c.ImageProcessingWorker = worker
	c.Logger.Info("Workers initialized")
	return nil
}

func (c *Container) initEventHandlers(ctx context.Context) error {
	// Upload Handler
	c.NewFileHandler = apphandler.NewNewFileHandler(
		c.UploadSubscriber,
		c.UOW,
		c.EventPublisher,
		c.Logger.WithGroup("upload_handler"),
	)

	// Image Process Handler
	c.ImageProcessHandler = apphandler.NewImageProcessHandler(
		c.ProcessSubscriber,
		c.ImageProcessingWorker,
		c.ImageRepo,
		c.Logger.WithGroup("image_process_handler"),
	)

	// Image Process Complete Handler
	c.ImageProcessCompleteHandler = apphandler.NewImageProcessCompleteHandler(
		c.CompleteSubscriber,
		c.EventPublisher,
		c.ImageRepo,
		c.ContentRepo,
		c.Logger.WithGroup("image_process_complete_handler"),
	)

	c.Logger.Info("Event handlers initialized")
	return nil
}

func (c *Container) initSubscribers(ctx context.Context) error {
	// Start Upload Handler
	go func() {
		c.Logger.Info("Starting upload handler",
			slog.String("subscription", c.Config.PubSub.UploadStatus.Name))
		if err := c.NewFileHandler.Start(ctx); err != nil {
			c.Logger.Error("Upload handler error", slog.String("error", err.Error()))
		}
	}()

	// Start Image Process Handler
	go func() {
		c.Logger.Info("Starting image process handler",
			slog.String("subscription", c.Config.PubSub.ImageProcessingRequest.Subscription.Name))
		if err := c.ImageProcessHandler.Start(ctx); err != nil {
			c.Logger.Error("Image process handler error", slog.String("error", err.Error()))
		}
	}()

	// Start Image Process Complete Handler
	go func() {
		c.Logger.Info("Starting image process complete handler",
			slog.String("subscription", c.Config.PubSub.ImageProcessingResult.Subscription.Name))
		if err := c.ImageProcessCompleteHandler.Start(ctx); err != nil {
			c.Logger.Error("Image process complete handler error", slog.String("error", err.Error()))
		}
	}()

	c.Logger.Info("All subscribers started")
	return nil
}

func (c *Container) initProxies(ctx context.Context) error {
	// Initialize KeyBuilder for TileServer
	keyBuilder := cache.NewKeyBuilder("tile_server")

	// Initialize TileServer
	c.TileServer = proxy.NewTileServer(
		c.Cache,
		keyBuilder,
		c.ContentRepo,
		c.ImageRepo,
		c.ProcessedStorage,
	)

	c.Logger.Info("Proxies initialized")
	return nil
}

func (c *Container) initHTTPLayer(ctx context.Context) error {
	// Handlers

	// Handlers
	c.WorkspaceHandler = handler.NewWorkspaceHandler(
		c.WorkspaceQuery,
		c.WorkspaceUseCase,
		c.Logger,
	)

	c.PatientHandler = handler.NewPatientHandler(
		c.PatientQuery,
		c.PatientUseCase,
		c.Logger,
	)

	c.ImageHandler = handler.NewImageHandler(
		c.ImageQuery,
		c.ImageUseCase,
		c.Logger,
	)

	c.AnnotationHandler = handler.NewAnnotationHandler(
		c.AnnotationQuery,
		c.AnnotationUseCase,
		c.Logger,
	)

	c.AnnotationTypeHandler = handler.NewAnnotationTypeHandler(
		c.AnnotationTypeQuery,
		c.AnnotationTypeUseCase,
		c.Logger,
	)

	// Middleware
	c.AuthMiddleware = middleware.NewAuthMiddleware(c.Logger)
	c.TimeoutMiddleware = middleware.NewTimeoutMiddleware(
		30*time.Second,
		c.Logger,
	)

	// Tile Proxy Handler
	c.TileProxyHandler = handler.NewTileProxyHandler(
		c.TileServer,
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
		c.TileProxyHandler,
		c.AuthMiddleware,
		c.TimeoutMiddleware,
	)

	c.Logger.Info("HTTP layer initialized")
	return nil
}

func (c *Container) Close() error {
	c.Logger.Info("Closing container resources...")

	var errs []error

	// Stop all event handlers
	if c.NewFileHandler != nil {
		if err := c.NewFileHandler.Stop(); err != nil {
			c.Logger.Error("Error stopping upload handler", slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("upload handler stop: %w", err))
		}
	}

	if c.ImageProcessHandler != nil {
		if err := c.ImageProcessHandler.Stop(); err != nil {
			c.Logger.Error("Error stopping image process handler", slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("image process handler stop: %w", err))
		}
	}

	if c.ImageProcessCompleteHandler != nil {
		if err := c.ImageProcessCompleteHandler.Stop(); err != nil {
			c.Logger.Error("Error stopping image process complete handler", slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("image process complete handler stop: %w", err))
		}
	}

	// Close other resources
	if c.FirestoreClient != nil {
		if err := c.FirestoreClient.Close(); err != nil {
			c.Logger.Error("Failed to close Firestore client", slog.String("error", err.Error()))
			errs = append(errs, fmt.Errorf("firestore close: %w", err))
		}
	}

	if c.StorageClient != nil {
		if err := c.StorageClient.Close(); err != nil {
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
