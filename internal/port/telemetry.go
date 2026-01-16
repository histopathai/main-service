package port

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/shared/query"
)

// TelemetryRepository handles persistence of telemetry messages
type TelemetryRepository interface {
	Create(ctx context.Context, message *events.TelemetryMessage) error
	Read(ctx context.Context, id string) (*events.TelemetryMessage, error)
	FindByFilters(ctx context.Context, filters []query.Filter, pagination *query.Pagination) (*query.Result[*events.TelemetryMessage], error)
	Count(ctx context.Context, filters []query.Filter) (int64, error)
	Delete(ctx context.Context, id string) error
	BatchDelete(ctx context.Context, ids []string) error
}
