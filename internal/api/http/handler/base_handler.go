package handler

import (
	"net/http"

	stderr "errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type BaseHandler struct {
	logger   *slog.Logger
	response *ResponseHelper
}

func NewBaseHandler(logger *slog.Logger) BaseHandler {
	return BaseHandler{
		logger:   logger,
		response: &ResponseHelper{},
	}
}

func (bh *BaseHandler) handleError(c *gin.Context, err error) {
	requestID, _ := c.Get("request_id")
	var customErr *errors.Err

	if stderr.As(err, &customErr) {
		statusCode, errResponse := bh.mapCustomError(customErr)

		bh.logger.Error("Request failed",
			slog.String("request_id", requestID.(string)),
			slog.String("error_type", string(customErr.Type)),
			slog.String("message", customErr.Message),
			slog.String("path", c.Request.URL.Path),
		)
		c.JSON(statusCode, errResponse)
		return
	}

	// Unknown error
	bh.logger.Error("Unexpected error",
		slog.String("error", err.Error()),
		slog.String("path", c.Request.URL.Path),
	)
	c.JSON(http.StatusInternalServerError, response.ErrorResponse{
		ErrorType: string(errors.ErrorTypeInternal),
		Message:   "An unexpected error occurred",
	})
}

func (bh *BaseHandler) mapCustomError(err *errors.Err) (int, response.ErrorResponse) {
	statusMap := map[errors.ErrorType]int{
		errors.ErrorTypeValidation:   http.StatusBadRequest,
		errors.ErrorTypeNotFound:     http.StatusNotFound,
		errors.ErrorTypeConflict:     http.StatusConflict,
		errors.ErrorTypeUnauthorized: http.StatusUnauthorized,
		errors.ErrorTypeForbidden:    http.StatusForbidden,
		errors.ErrorTypeInternal:     http.StatusInternalServerError,
	}

	statusCode, exists := statusMap[err.Type]
	if !exists {
		statusCode = http.StatusInternalServerError
	}

	return statusCode, response.ErrorResponse{
		ErrorType: string(err.Type),
		Message:   err.Message,
		Details:   err.Details,
	}
}
