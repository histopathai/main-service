package gcs

import (
	"errors"
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

func mapGCSError(err error) error {
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
			return ErrNotFound
		case http.StatusForbidden:
			return ErrForbidden
		case http.StatusUnauthorized:
			return ErrUnauthorized
		case http.StatusConflict:
			return ErrConflict
		case http.StatusServiceUnavailable, http.StatusInternalServerError:
			return errors.Join(ErrInternal, errors.New("service unavailable"))
		default:
			return errors.Join(ErrInternal, err)
		}
	}

	return errors.Join(ErrInternal, err)
}
