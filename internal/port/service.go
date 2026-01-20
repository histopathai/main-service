package port

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/shared/query"
)

type Service[T Entity] interface {
	Create(ctx context.Context, cmd command.CreateCommand) (T, error)
	Get(ctx context.Context, cmd command.ReadCommand) (T, error)
	Update(ctx context.Context, cmd command.UpdateCommand) error
	Delete(ctx context.Context, cmd command.DeleteCommand) error
	DeleteMany(ctx context.Context, cmd command.DeleteCommands) error
	List(ctx context.Context, cmd command.ListCommand) (*[]query.Result[T], error)
	Count(ctx context.Context, cmd command.CountCommand) (int64, error)
	GetByParentID(ctx context.Context, cmd command.ReadByParentIDCommand) (*[]query.Result[T], error)
}

type TelemetryStats struct {
	TotalErrors  int64                          `json:"total_errors"`
	BySeverity   map[events.ErrorSeverity]int64 `json:"by_severity"`
	ByCategory   map[events.ErrorCategory]int64 `json:"by_category"`
	RecentErrors []*events.TelemetryMessage     `json:"recent_errors"`
	ErrorTrend   []ErrorTrendPoint              `json:"error_trend"`
}

type ErrorTrendPoint struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
}
type ITelemetryService interface {
	RecordDLQMessage(ctx context.Context, event *events.DLQMessageEvent) error
	RecordError(ctx context.Context, event *events.TelemetryErrorEvent) error
	ListMessages(ctx context.Context, filters []query.Filter, pagination *query.Pagination) (*query.Result[*events.TelemetryMessage], error)
	GetMessageByID(ctx context.Context, id string) (*events.TelemetryMessage, error)
	GetErrorStats(ctx context.Context) (*TelemetryStats, error)
}
