package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/histopathai/main-service/internal/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func handleError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		statusCode := getHTTPStatusFromErrorType(appErr.Type)

		response := gin.H{
			"error": appErr.Message,
			"type":  appErr.Type,
		}

		if appErr.Details != nil {
			response["details"] = appErr.Details
		}

		c.JSON(statusCode, response)
		return
	}

	if st, ok := status.FromError(err); ok {
		statusCode, message := getHTTPStatusFromGRPCCode(st.Code())
		c.JSON(statusCode, gin.H{
			"error": message,
			"type":  "INTERNAL_ERROR",
		})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{
		"error": "internal server error",
		"type":  "internal_error",
	})
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

func getHTTPStatusFromGRPCCode(code codes.Code) (int, string) {
	switch code {
	case codes.NotFound:
		return http.StatusNotFound, "resource not found"
	case codes.AlreadyExists:
		return http.StatusConflict, "resource already exists"
	case codes.PermissionDenied:
		return http.StatusForbidden, "permission denied"
	case codes.Unauthenticated:
		return http.StatusUnauthorized, "unauthenticated"
	case codes.InvalidArgument:
		return http.StatusBadRequest, "invalid argument"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
