// internal/api/http/middleware/auth.go
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

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

		// **DEBUG LOG EKLE**
		am.logger.Debug("Auth middleware checking",
			"has_user_id_header", userID != "",
			"user_id", userID,
		)

		if userID == "" {
			debugUserID, exists := c.Get("user_id")
			if !exists {
				am.logger.Warn("No authentication found in headers or context")
				c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					ErrorType: string(errors.ErrorTypeUnauthorized),
					Message:   "user not authenticated",
				})
				c.Abort()
				return
			}
			userID = debugUserID.(string)
		}

		c.Set("user_id", userID)
		c.Set("authenticated_user_id", userID)

		// Get User Role from Header
		userRole := c.GetHeader("X-User-Role")

		// **DEBUG LOG EKLE**
		am.logger.Debug("Auth middleware role check",
			"has_role_header", userRole != "",
			"user_role", userRole,
		)

		if userRole == "" {
			debugUserRole, exists := c.Get("user_role")
			if !exists {
				am.logger.Warn("No role found in headers or context")
				c.JSON(http.StatusUnauthorized, response.ErrorResponse{
					ErrorType: string(errors.ErrorTypeUnauthorized),
					Message:   "user role not found in context",
				})
				c.Abort()
				return
			}
			userRole = debugUserRole.(string)
		}

		c.Set("user_role", userRole)
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

func GetAuthenticatedUserRole(c *gin.Context) (string, error) {
	userRole, exists := c.Get("user_role")
	if !exists {
		return "", errors.NewUnauthorizedError("user role not found in context")
	}
	return userRole.(string), nil
}

// RequireRole rejects requests whose X-User-Role (set by RequireAuth) is not
// one of roles, compared case-insensitively.
func (am *AuthMiddleware) RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetAuthenticatedUserRole(c)
		if err == nil {
			for _, allowed := range roles {
				if strings.EqualFold(role, allowed) {
					c.Next()
					return
				}
			}
		}
		am.logger.Warn("Role check failed", "path", c.Request.URL.Path, "role", role, "required", roles)
		c.JSON(http.StatusForbidden, response.ErrorResponse{
			ErrorType: string(errors.ErrorTypeForbidden),
			Message:   "insufficient role",
		})
		c.Abort()
	}
}
