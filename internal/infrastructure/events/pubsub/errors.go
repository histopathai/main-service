// internal/infrastructure/events/pubsub/errors.go
package pubsub

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Basit, olay yayınlama altyapısına özel hatalar.
var (
	ErrInternal      = errors.New("pubsub internal error")
	ErrPublishFailed = errors.New("failed to publish message")
	ErrTopicNotFound = errors.New("topic not found for event")
)

func mapPubSubError(err error, context string) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if ok {
		switch st.Code() {
		case codes.NotFound:
			return fmt.Errorf("%s: %w: %v", context, ErrTopicNotFound, err)
		case codes.Unavailable, codes.DeadlineExceeded, codes.Internal, codes.Unknown:
			return fmt.Errorf("%s: %w: %v", context, ErrInternal, err)
		default:
			return fmt.Errorf("%s: %w: %v", context, ErrPublishFailed, err)
		}
	}

	return fmt.Errorf("%s: %w: %v", context, ErrInternal, err)
}
