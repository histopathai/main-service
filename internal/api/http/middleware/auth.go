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
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			debugUserID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					ErrorType: string(errors.ErrorTypeUnauthorized),
					Message:   "Authentication required (Header or context missing)",
				})
				c.Abort()
				return
			}
			userID = debugUserID.(string)
		}

		c.Set("user_id", userID)

		// Ensure userID is a non-empty string
		if userID == "" {
			c.JSON(http.StatusUnauthorized, response.ErrorResponse{
				ErrorType: string(errors.ErrorTypeUnauthorized),
				Message:   "Invalid user ID format",
			})
			c.Abort()
			return
		}

		// Set typed value back to context
		c.Set("authenticated_user_id", userID)
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
