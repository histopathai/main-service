package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	apperrors "github.com/histopathai/main-service/internal/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func handleError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	var validationErrors validator.ValidationErrors

	if errors.As(err, &appErr) {
	} else if st, ok := status.FromError(err); ok {
		appErr = convertGRPCErrToAppErr(st)
	} else if errors.As(err, &validationErrors) {
		details := make(map[string]interface{})
		for _, fieldErr := range validationErrors {
			details[fieldErr.Field()] = "failed validation: " + fieldErr.Tag()
		}
		appErr = apperrors.NewValidationError("validation failed", details)
	} else {
		appErr = apperrors.NewInternalError("internal server error", err)
	}
	statusCode := getHTTPStatusFromErrorType(appErr.Type)

	response := gin.H{
		"error": appErr.Message,
		"type":  appErr.Type,
	}

	if appErr.Details != nil {
		response["details"] = appErr.Details
	}

	c.JSON(statusCode, response)
}

func convertGRPCErrToAppErr(st *status.Status) *apperrors.AppError {
	message := st.Message()

	switch st.Code() {
	case codes.NotFound:
		return apperrors.NewNotFoundError(message)
	case codes.AlreadyExists:
		return apperrors.NewConflictError(message)
	case codes.PermissionDenied:
		return apperrors.NewForbiddenError(message)
	case codes.Unauthenticated:
		return apperrors.NewUnauthorizedError(message)
	case codes.InvalidArgument:
		return apperrors.NewValidationError(message, nil)
	default:
		return apperrors.NewInternalError(message, st.Err())
	}
}

func getHTTPStatusFromErrorType(errType apperrors.ErrorType) int {
	switch errType {
	case apperrors.ErrorTypeValidation:
		return http.StatusBadRequest
	case apperrors.ErrorTypeNotFound:
		return http.StatusNotFound
	case apperrors.ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case apperrors.ErrorTypeForbidden:
		return http.StatusForbidden
	case apperrors.ErrorTypeConflict:
		return http.StatusConflict
	case apperrors.ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
