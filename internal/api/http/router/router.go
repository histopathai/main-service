package router

import (
	"log/slog"
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
	return &Router{engine: gin.Default(),
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
func debugUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Test için sahte bir kullanıcı ID'si ekle
		c.Set("user_id", "local-debug-user-uuid")
		c.Next()
	}
}

func (r *Router) SetupRoutes() *gin.Engine {
	//Global Middlewares
	r.engine.Use(middleware.RequestIDMiddleware())
	r.engine.Use(r.timeoutMiddleware.Handler())

	// Health check endpoint (no auth required)
	r.engine.GET("/health", r.healthCheck)
	r.engine.GET("/ready", r.readinessCheck)

	// Swagger documentation endpoint
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// API v1 routes
	v1 := r.engine.Group("/api/v1")
	{

		if gin.Mode() == gin.DebugMode {
			r.config.Logger.Info("Running in debug mode, applying debug user middleware")
			v1.Use(debugUserMiddleware())
		} else {
			// Apply authentication middleware to all v1 routes
			v1.Use(r.authMiddleware.RequireAuth())
		}
		// Workspace routes
		r.setupWorkspaceRoutes(v1)

		// Patient routes
		r.setupPatientRoutes(v1)

		// Image routes
		r.setupImageRoutes(v1)

		// Annotation routes
		r.setupAnnotationRoutes(v1)

		// Annotation Type routes
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

		// Nested patient routes
		workspaces.GET("/:workspace_id/patients", r.patientHandler.GetPatientsByWorkspaceID)
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

		// Nested image routes
		patients.GET("/:patient_id/images", r.imageHandler.ListImageByPatientID)
	}
}

func (r *Router) setupImageRoutes(rg *gin.RouterGroup) {
	images := rg.Group("/images")
	{
		images.POST("", r.imageHandler.UploadImage)
		images.GET("/:image_id", r.imageHandler.GetImageByID)
		images.DELETE("/:image_id", r.imageHandler.DeleteImage)
	}
}

func (r *Router) setupAnnotationRoutes(rg *gin.RouterGroup) {
	annotations := rg.Group("/annotations")
	{
		annotations.POST("", r.annotationHandler.CreateNewAnnotation)
		annotations.GET("/:id", r.annotationHandler.GetAnnotationByID)
		annotations.DELETE("/:id", r.annotationHandler.DeleteAnnotation)
		annotations.GET("/image/:image_id", r.annotationHandler.GetAnnotationsByImageID)
	}
}

func (r *Router) setupAnnotationTypeRoutes(rg *gin.RouterGroup) {
	annotationTypes := rg.Group("/annotation-types")
	{
		annotationTypes.POST("", r.annotationTypeHandler.CreateNewAnnotationType)
		annotationTypes.GET("", r.annotationTypeHandler.ListAnnotationTypes)
		annotationTypes.GET("/:id", r.annotationTypeHandler.GetAnnotationType)
		annotationTypes.PUT("/:id", r.annotationTypeHandler.UpdateAnnotationType)
		annotationTypes.DELETE("/:id", r.annotationTypeHandler.DeleteAnnotationType)
		annotationTypes.GET("/classification-enabled", r.annotationTypeHandler.GetClassificationOptionedAnnotationTypes)
		annotationTypes.GET("/score-enabled", r.annotationTypeHandler.GetScoreOptionedAnnotationTypes)
	}
}

func (r *Router) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}

func (r *Router) readinessCheck(c *gin.Context) {
	// TODO: Add actual readiness checks (DB, PubSub, etc.)
	c.JSON(200, gin.H{
		"status": "ready",
		"time":   time.Now().UTC(),
	})
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
