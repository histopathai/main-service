// internal/api/http/middleware/auth.go
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type AuthMiddleware struct {
	logger *slog.Logger
}

func NewAuthMiddleware(logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{logger: logger}
}

func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				ErrorType: string(errors.ErrorTypeUnauthorized),
				Message:   "Authentication required",
			})
			c.Abort()
			return
		}

		// Type assertion with safety check
		userIDStr, ok := userID.(string)
		if !ok || userIDStr == "" {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				ErrorType: string(errors.ErrorTypeUnauthorized),
				Message:   "Invalid user ID format",
			})
			c.Abort()
			return
		}

		// Set typed value back to context
		c.Set("authenticated_user_id", userIDStr)
		c.Next()
	}
}

// Helper function for handlers
func GetAuthenticatedUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("authenticated_user_id")
	if !exists {
		return "", errors.NewUnauthorizedError("user not authenticated")
	}
	return userID.(string), nil
}
