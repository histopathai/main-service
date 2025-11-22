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
		if gin.Mode() == gin.DebugMode {
			r.config.Logger.Info("Running in debug mode, applying debug user middleware")
			v1.Use(debugUserMiddleware())
		} else {
			v1.Use(r.authMiddleware.RequireAuth())
		}

		r.setupWorkspaceRoutes(v1)
		r.setupPatientRoutes(v1)
		r.setupImageRoutes(v1)
		r.setupAnnotationRoutes(v1)
		r.setupAnnotationTypeRoutes(v1)

		v1.GET("/proxy/*objectPath", r.gcsProxyHandler.ProxyObject)
	}

	return r.engine
}

func (r *Router) setupWorkspaceRoutes(rg *gin.RouterGroup) {
	workspaces := rg.Group("/workspaces")
	{
		workspaces.POST("", r.workspaceHandler.CreateNewWorkspace)
		workspaces.GET("", r.workspaceHandler.ListWorkspaces)
		workspaces.GET("/:workspace_id", r.workspaceHandler.GetWorkspaceByID)
		workspaces.PUT("/:workspace_id", r.workspaceHandler.UpdateWorkspace)
		workspaces.DELETE("/:workspace_id", r.workspaceHandler.DeleteWorkspace)
		workspaces.GET("/:workspace_id/patients", r.patientHandler.GetPatientsByWorkspaceID)
		workspaces.DELETE("/:workspace_id/cascade-delete", r.workspaceHandler.CascadeDeleteWorkspace)
		workspaces.DELETE("/batch-delete", r.workspaceHandler.BatchDeleteWorkspaces)
		workspaces.GET("/count-v1", r.workspaceHandler.CountV1Workspaces)
	}
}

func (r *Router) setupPatientRoutes(rg *gin.RouterGroup) {
	patients := rg.Group("/patients")
	{
		patients.POST("", r.patientHandler.CreateNewPatient)
		patients.GET("", r.patientHandler.ListPatients)
		patients.GET("/:patient_id", r.patientHandler.GetPatientByID)
		patients.PUT("/:patient_id", r.patientHandler.UpdatePatientByID)
		patients.DELETE("/:patient_id", r.patientHandler.DeletePatientByID)
		patients.POST("/:patient_id/transfer/:workspace_id", r.patientHandler.TransferPatientWorkspace)
		patients.GET("/:patient_id/images", r.imageHandler.ListImageByPatientID)
		patients.GET("/count-v1", r.patientHandler.CountV1Patients)
		patients.DELETE("/batch-delete", r.patientHandler.BatchDeletePatients)
		patients.DELETE("/:patient_id/cascade-delete", r.patientHandler.CascadeDeletePatient)
		patients.POST("/batch-transfer", r.patientHandler.BatchTransferPatients)

	}
}

func (r *Router) setupImageRoutes(rg *gin.RouterGroup) {
	images := rg.Group("/images")
	{
		images.POST("", r.imageHandler.UploadImage)
		images.GET("/:image_id", r.imageHandler.GetImageByID)
		images.DELETE("/:image_id", r.imageHandler.DeleteImage)
		images.DELETE("/batch-delete", r.imageHandler.BatchDeleteImages)
		images.GET("/count-v1", r.imageHandler.CountV1Images)
		images.POST("/batch-transfer", r.imageHandler.BatchTransferImages)
	}
}

func (r *Router) setupAnnotationRoutes(rg *gin.RouterGroup) {
	annotations := rg.Group("/annotations")
	{
		annotations.POST("", r.annotationHandler.CreateNewAnnotation)
		annotations.GET("/:id", r.annotationHandler.GetAnnotationByID)
		annotations.DELETE("/:id", r.annotationHandler.DeleteAnnotation)
		annotations.GET("/image/:image_id", r.annotationHandler.GetAnnotationsByImageID)
		annotations.DELETE("/batch-delete", r.annotationHandler.BatchDeleteAnnotations)
		annotations.GET("/count-v1", r.annotationHandler.CountV1Annotations)
	}
}

func (r *Router) setupAnnotationTypeRoutes(rg *gin.RouterGroup) {
	annotationTypes := rg.Group("/annotation-types")
	{
		annotationTypes.POST("", r.annotationTypeHandler.CreateNewAnnotationType)
		annotationTypes.GET("", r.annotationTypeHandler.ListAnnotationTypes)
		annotationTypes.GET("/:annotation_type_id", r.annotationTypeHandler.GetAnnotationType)
		annotationTypes.PUT("/:annotation_type_id", r.annotationTypeHandler.UpdateAnnotationType)
		annotationTypes.DELETE("/:annotation_type_id", r.annotationTypeHandler.DeleteAnnotationType)
		annotationTypes.GET("/classification-enabled", r.annotationTypeHandler.GetClassificationOptionedAnnotationTypes)
		annotationTypes.GET("/score-enabled", r.annotationTypeHandler.GetScoreOptionedAnnotationTypes)
		annotationTypes.GET("/count-v1", r.annotationTypeHandler.CountV1AnnotationTypes)
		annotationTypes.DELETE("/batch-delete", r.annotationTypeHandler.BatchDeleteAnnotationTypes)
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
