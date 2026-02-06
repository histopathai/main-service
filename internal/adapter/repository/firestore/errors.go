package firestore

import (
	"errors"
	"strings"

	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound           = errors.New("document not found")
	ErrAlreadyExists      = errors.New("document already exists")
	ErrTransactionAborted = errors.New("transaction aborted")
	ErrConflict           = errors.New("conflict occurred")
	ErrInternal           = errors.New("internal error occurred")
	ErrInvalidInput       = errors.New("invalid input or nil entity")
)

func mapFirestoreError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, iterator.Done) {
		return nil
	}

	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.NotFound:
			return ErrNotFound
		case codes.AlreadyExists:
			return ErrAlreadyExists
		case codes.Aborted:
			return ErrTransactionAborted
		default:
			return errors.Join(ErrInternal, err)
		}
	}

	return errors.Join(ErrInternal, err)
}

// isCollectionNotFoundError checks if the error is due to a non-existent collection
// Firestore doesn't explicitly return "collection not found" errors, but we can infer
// from certain error patterns that the collection doesn't exist
func isCollectionNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check for gRPC status codes
	st, ok := status.FromError(err)
	if ok {
		// NotFound can indicate collection doesn't exist
		if st.Code() == codes.NotFound {
			return true
		}
		// FailedPrecondition with certain messages can also indicate missing collection
		if st.Code() == codes.FailedPrecondition {
			msg := strings.ToLower(st.Message())
			if strings.Contains(msg, "collection") || strings.Contains(msg, "not found") {
				return true
			}
		}
	}

	// Check error message for collection-related errors
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "collection") && strings.Contains(errMsg, "not found") {
		return true
	}

	return false
}
