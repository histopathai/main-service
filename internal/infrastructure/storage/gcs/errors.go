package gcs

import (
	"errors"
	"fmt"
	"net/http"

	"cloud.google.com/go/storage"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
)

var (
	ErrNotFound     = errors.New("object not found in gcs")
	ErrForbidden    = errors.New("forbidden: insufficient permissions for gcs resource")
	ErrUnauthorized = errors.New("unauthorized access to gcs resource")
	ErrInternal     = errors.New("gcs internal error")
	ErrConflict     = errors.New("conflict in gcs operation")
)

func mapGCSError(err error, context string) error {
	if err == nil || errors.Is(err, iterator.Done) {
		return nil
	}

	if errors.Is(err, storage.ErrObjectNotExist) {
		return ErrNotFound
	}

	var gErr *googleapi.Error
	if errors.As(err, &gErr) {
		switch gErr.Code {
		case http.StatusNotFound:
			return fmt.Errorf("%s: %w: %v", context, ErrNotFound, err)
		case http.StatusForbidden:
			return fmt.Errorf("%s: %w: %v", context, ErrForbidden, err)
		case http.StatusUnauthorized:
			return fmt.Errorf("%s: %w: %v", context, ErrUnauthorized, err)
		case http.StatusConflict:
			return fmt.Errorf("%s: %w: %v", context, ErrConflict, err)
		case http.StatusServiceUnavailable, http.StatusInternalServerError:
			return fmt.Errorf("%s: %w: %v", context, ErrInternal, errors.New("service unavailable"))
		default:
			return fmt.Errorf("%s: %w: %v", context, ErrInternal, err)
		}
	}

	return fmt.Errorf("%s: %w: %v", context, ErrInternal, err)
}
