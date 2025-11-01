// internal/api/http/middleware/timeout.go
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/response"
)

type TimeoutMiddleware struct {
	timeout time.Duration
	logger  *slog.Logger
}

func NewTimeoutMiddleware(timeout time.Duration, logger *slog.Logger) *TimeoutMiddleware {
	return &TimeoutMiddleware{
		timeout: timeout,
		logger:  logger,
	}
}

func (tm *TimeoutMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), tm.timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			close(finished)
		}()

		select {
		case <-finished:
			return

		case p := <-panicChan:
			panic(p)

		case <-ctx.Done():
			c.Header("Connection", "close")
			c.AbortWithStatusJSON(http.StatusRequestTimeout, response.ErrorResponse{
				ErrorType: "TIMEOUT_ERROR",
				Message:   "Request processing timeout exceeded",
				Details: map[string]interface{}{
					"timeout": tm.timeout.String(),
				},
			})

			tm.logger.Warn("Request timeout",
				slog.String("path", c.Request.URL.Path),
				slog.String("method", c.Request.Method),
				slog.Duration("timeout", tm.timeout),
			)
		}
	}
}
