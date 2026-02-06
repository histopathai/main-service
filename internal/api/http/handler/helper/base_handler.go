package helper

import (
	"net/http"

	stderr "errors"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type BaseHandler struct {
	logger   *slog.Logger
	Response *ResponseHelper
}

func NewBaseHandler(logger *slog.Logger) BaseHandler {
	return BaseHandler{
		logger:   logger,
		Response: &ResponseHelper{},
	}
}

func (bh *BaseHandler) HandleError(c *gin.Context, err error) {
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

func (bh *BaseHandler) ApplyVisibilityFilters(c *gin.Context, spec *query.Specification) {
	// 1. Get User Role
	role, err := middleware.GetAuthenticatedUserRole(c)
	if err != nil {
		// Log warning but proceed as normal user (restrictive)
		bh.logger.Warn("Could not get user role for visibility filter, defaulting to restrictive",
			slog.String("path", c.Request.URL.Path),
			slog.String("error", err.Error()))
		role = ""
	}

	isAdmin := role == "admin" || role == "ADMIN"

	isDeletedField := fields.EntityIsDeleted.DomainName()
	isDeletedAPIName := fields.EntityIsDeleted.APIName() // "is_deleted"

	// Check if "is_deleted" filter already exists
	hasDeletedFilter := false
	deletedFilterIndex := -1

	for i, f := range spec.Filters {
		// Check both Domain Name "IsDeleted" and API name "is_deleted" just in case
		if f.Field == isDeletedField || f.Field == isDeletedAPIName {
			hasDeletedFilter = true
			deletedFilterIndex = i
			break
		}
	}

	// 2. Logic Application
	if isAdmin {
		// ADMIN:
		// If NO filter provided -> Assume they want ACTIVE records only (IsDeleted = false)
		// If filter provided -> Respect it (they can ask for IsDeleted = true or false)
		if !hasDeletedFilter {
			spec.Filters = append(spec.Filters, query.Filter{
				Field:    isDeletedAPIName, // Use API Name "is_deleted"
				Operator: query.OpEqual,
				Value:    false,
			})
		}
	} else {
		// NORMAL USER:
		// FORCE IsDeleted = false.
		// If they provided a filter, REMOVE it or OVERRIDE it.
		// Removal + Append is safer to ensure correct value.

		if hasDeletedFilter {
			// Remove existing filter to override
			// Efficient removal from slice
			spec.Filters = append(spec.Filters[:deletedFilterIndex], spec.Filters[deletedFilterIndex+1:]...)
		}

		// Always append IsDeleted = false
		spec.Filters = append(spec.Filters, query.Filter{
			Field:    isDeletedAPIName, // Use API Name "is_deleted"
			Operator: query.OpEqual,
			Value:    false,
		})
	}
}
