package firestore

import (
	"errors"

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
