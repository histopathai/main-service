package router

import (
	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/handler"
	"github.com/histopathai/main-service/internal/middleware"
)

type RouterConfig struct {
	UploadHandler    *handler.UploadHandler
	WorkspaceHandler *handler.WorkspaceHandler
	PatientHandler   *handler.PatientHandler
}

func SetupRouter(cfg *RouterConfig) *gin.Engine {
	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// API v1 routes
	v1 := r.Group("/api/v1")

	{
		v1.Use(middleware.AuthMiddleware())

		workspaces := v1.Group("/workspaces")
		{
			cfg.WorkspaceHandler.RegisterRoutes(workspaces)
		}
		patients := v1.Group("/patients")
		{
			cfg.PatientHandler.RegisterRoutes(patients)
		}
		uploads := v1.Group("/uploads")
		{
			cfg.UploadHandler.RegisterRoutes(uploads)
		}

	}

	return r
}
