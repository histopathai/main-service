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

	//Handlers
	workspaceHandler      *handler.WorkspaceHandler
	patientHandler        *handler.PatientHandler
	imageHandler          *handler.ImageHandler
	annotationHandler     *handler.AnnotationHandler
	annotationTypeHandler *handler.AnnotationTypeHandler
	gcsProxyHandler       *handler.GCSProxyHandler

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
	gcsProxyHandler *handler.GCSProxyHandler,
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
		gcsProxyHandler:       gcsProxyHandler,
		authMiddleware:        authMiddleware,
		timeoutMiddleware:     timeoutMiddleware,
	}
}

func (r *Router) SetHealthChecker(hc HealthChecker) {
	r.healthChecker = hc
}

func debugUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", "local-debug-user-uuid")
		c.Next()
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	//Global Middlewares
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
		//if gin.Mode() == gin.DebugMode {
		//	r.config.Logger.Info("Running in debug mode, applying debug user middleware")
		//	v1.Use(debugUserMiddleware())
		//} else {
		v1.Use(r.authMiddleware.RequireAuth())
		//}

		r.setupWorkspaceRoutes(v1)
		r.setupPatientRoutes(v1)
		r.setupImageRoutes(v1)
		r.setupAnnotationRoutes(v1)
		r.setupAnnotationTypeRoutes(v1)

		v1.GET("/proxy/:imageId/*objectPath", r.gcsProxyHandler.ProxyObject)
	}

	return r.engine
}

func (r *Router) setupWorkspaceRoutes(rg *gin.RouterGroup) {
	workspaces := rg.Group("/workspaces")
	{
		workspaces.POST("", r.workspaceHandler.Create)
		workspaces.GET("", r.workspaceHandler.List)
		workspaces.GET("/:id", r.workspaceHandler.Get)
		workspaces.PATCH("/:id", r.workspaceHandler.Update)
		workspaces.DELETE("/:id/soft-delete", r.workspaceHandler.SoftDelete)
		workspaces.GET("/:id/patients", r.patientHandler.GetByParentID)
		workspaces.GET("/count", r.workspaceHandler.Count)
		workspaces.DELETE("/soft-delete-many", r.workspaceHandler.SoftDeleteMany)
	}
}

func (r *Router) setupPatientRoutes(rg *gin.RouterGroup) {
	patients := rg.Group("/patients")
	{
		patients.POST("", r.patientHandler.CreateNewPatient)
		patients.POST("/list", r.patientHandler.List)
		patients.GET("/:id", r.patientHandler.Get)
		patients.PUT("/:id", r.patientHandler.Update)
		patients.DELETE("/:id/soft-delete", r.patientHandler.SoftDelete)
		patients.PUT("/:id/transfer/:workspace_id", r.patientHandler.Transfer)
		patients.GET("/:id/images", r.imageHandler.List)
		patients.POST("/count", r.patientHandler.Count)
		patients.DELETE("/soft-delete-many", r.patientHandler.SoftDeleteMany)
		patients.PUT("/transfer-many/:workspace_id", r.patientHandler.TransferMany)
	}
}

func (r *Router) setupImageRoutes(rg *gin.RouterGroup) {
	images := rg.Group("/images")
	{
		images.POST("", r.imageHandler.UploadImage)
		images.GET("/:id", r.imageHandler.Get)
		images.PUT("/:id", r.imageHandler.Update)
		images.DELETE("/:id/soft-delete", r.imageHandler.SoftDelete)
		images.GET("/parent/:parent_id", r.imageHandler.GetByParentID)
		images.GET("/workspace/:workspace_id", r.imageHandler.GetByWorkspaceID)
		images.POST("/list", r.imageHandler.List)
		images.POST("/count", r.imageHandler.Count)
		images.PUT("/:id/transfer/:patient_id", r.imageHandler.Transfer)
		images.PUT("/transfer-many/:workspace_id", r.imageHandler.TransferMany)
		images.DELETE("/soft-delete-many", r.imageHandler.SoftDeleteMany)
	}
}

func (r *Router) setupAnnotationRoutes(rg *gin.RouterGroup) {
	annotations := rg.Group("/annotations")
	{
		annotations.POST("", r.annotationHandler.Create)
		annotations.GET("/:id", r.annotationHandler.Get)
		annotations.PUT("/:id", r.annotationHandler.Update)
		annotations.DELETE("/:id/soft-delete", r.annotationHandler.SoftDelete)
		annotations.GET("/image/:image_id", r.annotationHandler.GetByParentID)
		annotations.GET("/workspace/:workspace_id", r.annotationHandler.GetByWsID)
		annotations.POST("/count", r.annotationHandler.Count)
		annotations.DELETE("/soft-delete-many", r.annotationHandler.SoftDeleteMany)
	}
}

func (r *Router) setupAnnotationTypeRoutes(rg *gin.RouterGroup) {
	annotationTypes := rg.Group("/annotation-types")
	{
		annotationTypes.POST("", r.annotationTypeHandler.Create)
		annotationTypes.GET("", r.annotationTypeHandler.List)
		annotationTypes.GET("/:id", r.annotationTypeHandler.Get)
		annotationTypes.PUT("/:id", r.annotationTypeHandler.Update)
		annotationTypes.DELETE("/:id/soft-delete", r.annotationTypeHandler.SoftDelete)
		annotationTypes.POST("/count", r.annotationTypeHandler.Count)
		annotationTypes.DELETE("/soft-delete-many", r.annotationTypeHandler.SoftDeleteMany)
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
