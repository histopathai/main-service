package errors

import (
	"context"
	"errors"
	"io"
	"net/http"

	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func FromFirestoreError(err error) *Err {
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.NotFound:
			return NewNotFoundError("resource not found")
		case codes.PermissionDenied:
			return NewForbiddenError("permission denied")
		case codes.Unauthenticated:
			return NewUnauthorizedError("unauthenticated access")
		case codes.Unavailable:
			return NewInternalError("service unavailable", err)
		case codes.AlreadyExists:
			return NewConflictError("resource already exists", nil)
		default:
			return NewInternalError("internal server error", err)
		}
	}
	return NewInternalError("internal server error", err)
}

func FromPubSubError(err error) *Err {
	if err == nil {
		return nil
	}

	// gRPC error
	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.NotFound:
			return NewNotFoundError("pubsub resource not found")
		case codes.PermissionDenied:
			return NewForbiddenError("pubsub permission denied")
		case codes.Unauthenticated:
			return NewUnauthorizedError("unauthenticated pubsub access")
		case codes.DeadlineExceeded:
			return NewInternalError("pubsub deadline exceeded", err)
		case codes.Unavailable:
			return NewInternalError("pubsub service unavailable", err)
		default:
			return NewInternalError("unexpected pubsub error", err)
		}
	}

	// Non-gRPC errors (timeout, EOF vs.)
	if errors.Is(err, context.DeadlineExceeded) {
		return NewInternalError("pubsub operation timeout", err)
	}
	if errors.Is(err, io.EOF) {
		return NewInternalError("pubsub stream closed unexpectedly", err)
	}

	return NewInternalError("unknown pubsub error", err)
}

func FromGCSError(err error) *Err {
	if err == nil {
		return nil
	}

	var gErr *googleapi.Error
	if errors.As(err, &gErr) {
		switch gErr.Code {
		case http.StatusNotFound:
			return NewNotFoundError("object not found in GCS")
		case http.StatusForbidden:
			return NewForbiddenError("forbidden: insufficient permissions for GCS resource")
		case http.StatusUnauthorized:
			return NewUnauthorizedError("unauthorized access to GCS resource")
		case http.StatusConflict:
			return NewConflictError("conflict in GCS resource operation", nil)
		case http.StatusServiceUnavailable:
			return NewInternalError("GCS service unavailable", err)
		default:
			return NewInternalError("unexpected GCS error", err)
		}
	}

	return NewInternalError("unknown GCS error", err)
}

func FromExternalError(err error, service string) *Err {
	if err == nil {
		return nil
	}

	switch service {
	case "firestore":
		return FromFirestoreError(err)
	case "pubsub":
		return FromPubSubError(err)
	case "gcs":
		return FromGCSError(err)
	default:
		return NewInternalError("unknown external service error", err)
	}
}
