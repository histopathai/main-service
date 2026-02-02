package router

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/handler"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouterConfig struct {
	Logger         *slog.Logger
	RequestTimeout time.Duration
}

type HealthChecker interface {
	IsHealthy() bool
}

type Router struct {
	engine *gin.Engine
	config *RouterConfig

	// Handlers
	workspaceHandler      *handler.WorkspaceHandler
	patientHandler        *handler.PatientHandler
	imageHandler          *handler.ImageHandler
	annotationHandler     *handler.AnnotationHandler
	annotationTypeHandler *handler.AnnotationTypeHandler
	tileProxyHandler      *handler.TileProxyHandler

	// Middleware
	authMiddleware    *middleware.AuthMiddleware
	timeoutMiddleware *middleware.TimeoutMiddleware

	// Health checker (optional, can be nil)
	healthChecker HealthChecker
}

func NewRouter(
	config *RouterConfig,
	workspaceHandler *handler.WorkspaceHandler,
	patientHandler *handler.PatientHandler,
	imageHandler *handler.ImageHandler,
	annotationHandler *handler.AnnotationHandler,
	annotationTypeHandler *handler.AnnotationTypeHandler,
	tileProxyHandler *handler.TileProxyHandler,
	authMiddleware *middleware.AuthMiddleware,
	timeoutMiddleware *middleware.TimeoutMiddleware,
) *Router {
	return &Router{
		engine:                gin.Default(),
		config:                config,
		workspaceHandler:      workspaceHandler,
		patientHandler:        patientHandler,
		imageHandler:          imageHandler,
		annotationHandler:     annotationHandler,
		annotationTypeHandler: annotationTypeHandler,
		tileProxyHandler:      tileProxyHandler,
		authMiddleware:        authMiddleware,
		timeoutMiddleware:     timeoutMiddleware,
	}
}

func (r *Router) SetHealthChecker(hc HealthChecker) {
	r.healthChecker = hc
}

func (r *Router) SetupRoutes() *gin.Engine {
	// Global Middlewares
	r.engine.Use(middleware.RequestIDMiddleware())
	r.engine.Use(r.timeoutMiddleware.Handler())

	// Health check endpoints (no auth required) - MUST be before any other routes
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/ready", r.readinessCheck)
	r.engine.GET("/readiness", r.readinessCheck) // Alternative endpoint

	// Swagger documentation endpoint
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{
		v1.Use(r.authMiddleware.RequireAuth())

		r.setupWorkspaceRoutes(v1)
		r.setupPatientRoutes(v1)
		r.setupImageRoutes(v1)
		r.setupAnnotationRoutes(v1)
		r.setupAnnotationTypeRoutes(v1)

		// Tile Proxy
		v1.GET("/proxy/:imageId/*objectPath", r.tileProxyHandler.ProxyTile)
	}

	return r.engine
}

func (r *Router) setupWorkspaceRoutes(rg *gin.RouterGroup) {
	workspaces := rg.Group("/workspaces")
	{
		// CRUD Operations
		workspaces.POST("", r.workspaceHandler.Create)    // Create
		workspaces.GET("", r.workspaceHandler.List)       // List (with query params)
		workspaces.GET("/:id", r.workspaceHandler.Get)    // Get by ID
		workspaces.PUT("/:id", r.workspaceHandler.Update) // Update (changed from PATCH to PUT)

		// Soft Delete
		workspaces.DELETE("/:id/soft-delete", r.workspaceHandler.SoftDelete)
		workspaces.DELETE("/soft-delete-many", r.workspaceHandler.SoftDeleteMany)

		// Queries
		workspaces.GET("/count", r.workspaceHandler.Count) // Count (changed from POST to GET)

		// Sub-resources
		workspaces.GET("/:parent_id/patients", r.patientHandler.GetByParentID)
	}
}

func (r *Router) setupPatientRoutes(rg *gin.RouterGroup) {
	patients := rg.Group("/patients")
	{
		// CRUD Operations
		patients.POST("", r.patientHandler.Create)    // Create
		patients.GET("", r.patientHandler.List)       // List (changed from POST to GET, query params)
		patients.GET("/:id", r.patientHandler.Get)    // Get by ID
		patients.PUT("/:id", r.patientHandler.Update) // Update

		// Soft Delete
		patients.DELETE("/:id/soft-delete", r.patientHandler.SoftDelete)
		patients.DELETE("/soft-delete-many", r.patientHandler.SoftDeleteMany)

		// Transfer
		patients.PUT("/:id/transfer/:workspace_id", r.patientHandler.Transfer)
		patients.PUT("/transfer-many/:workspace_id", r.patientHandler.TransferMany)

		// Queries
		patients.GET("/count", r.patientHandler.Count) // Count (changed from POST to GET)

		// Sub-resources
		patients.GET("/:id/images", r.imageHandler.List)
	}
}

func (r *Router) setupImageRoutes(rg *gin.RouterGroup) {
	images := rg.Group("/images")
	{
		// CRUD Operations
		images.POST("", r.imageHandler.UploadImage) // Upload
		images.GET("/:id", r.imageHandler.Get)      // Get by ID
		images.PUT("/:id", r.imageHandler.Update)   // Update

		// Soft Delete
		images.DELETE("/:id/soft-delete", r.imageHandler.SoftDelete)
		images.DELETE("/soft-delete-many", r.imageHandler.SoftDeleteMany)

		// Transfer
		images.PUT("/:id/transfer/:patient_id", r.imageHandler.Transfer)
		images.PUT("/transfer-many/:patient_id", r.imageHandler.TransferMany)

		// Queries
		images.GET("/parent/:parent_id", r.imageHandler.GetByParentID)
		images.GET("/workspace/:workspace_id", r.imageHandler.GetByWorkspaceID)
		images.GET("/count", r.imageHandler.Count) // Count (changed from POST to GET)
	}
}

func (r *Router) setupAnnotationRoutes(rg *gin.RouterGroup) {
	annotations := rg.Group("/annotations")
	{
		// CRUD Operations
		annotations.POST("", r.annotationHandler.Create)    // Create
		annotations.GET("/:id", r.annotationHandler.Get)    // Get by ID
		annotations.PUT("/:id", r.annotationHandler.Update) // Update

		// Soft Delete
		annotations.DELETE("/:id/soft-delete", r.annotationHandler.SoftDelete)
		annotations.DELETE("/soft-delete-many", r.annotationHandler.SoftDeleteMany)

		// Queries
		annotations.GET("/image/:image_id", r.annotationHandler.GetByParentID)
		annotations.GET("/workspace/:workspace_id", r.annotationHandler.GetByWsID)
		annotations.GET("/count", r.annotationHandler.Count) // Count (changed from POST to GET)
	}
}

func (r *Router) setupAnnotationTypeRoutes(rg *gin.RouterGroup) {
	annotationTypes := rg.Group("/annotation-types")
	{
		// CRUD Operations
		annotationTypes.POST("", r.annotationTypeHandler.Create)    // Create
		annotationTypes.GET("", r.annotationTypeHandler.List)       // List (with query params)
		annotationTypes.GET("/:id", r.annotationTypeHandler.Get)    // Get by ID
		annotationTypes.PUT("/:id", r.annotationTypeHandler.Update) // Update

		// Soft Delete
		annotationTypes.DELETE("/:id/soft-delete", r.annotationTypeHandler.SoftDelete)
		annotationTypes.DELETE("/soft-delete-many", r.annotationTypeHandler.SoftDeleteMany)

		// Queries
		annotationTypes.GET("/count", r.annotationTypeHandler.Count) // Count (changed from POST to GET)
	}
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}

func (r *Router) readinessCheck(c *gin.Context) {
	// Check if health checker is available and healthy
	if r.healthChecker != nil && !r.healthChecker.IsHealthy() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "not ready",
			"time":   time.Now().UTC(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
		"time":   time.Now().UTC(),
	})
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
