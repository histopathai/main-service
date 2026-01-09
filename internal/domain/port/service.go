package port

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/query"
)

type CreateEntityInput struct {
	ID        *string
	Name      string
	Type      vobj.EntityType
	CreatorID string
	Parent    *vobj.ParentRef
}

type UpdateEntityInput struct {
	Name   *string
	Parent *vobj.ParentRef
}

type CreateWorkspaceInput struct {
	CreateEntityInput
	OrganType       string
	Organization    string
	Description     string
	License         string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

type UpdateWorkspaceInput struct {
	UpdateEntityInput
	OrganType       *string
	Organization    *string
	Description     *string
	License         *string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

type IWorkspaceService interface {
	CreateNewWorkspace(ctx context.Context, input *CreateWorkspaceInput) (*model.Workspace, error)
	UpdateWorkspace(ctx context.Context, id string, input UpdateWorkspaceInput) error
	GetWorkspaceByID(ctx context.Context, id string) (*model.Workspace, error)
	DeleteWorkspace(ctx context.Context, id string) error
	ListWorkspaces(ctx context.Context, pagination *query.Pagination) (*query.Result[*model.Workspace], error)
	BatchDeleteWorkspaces(ctx context.Context, workspaceIDs []string) error
	CascadeDeleteWorkspace(ctx context.Context, workspaceID string) error
	CountWorkspaces(ctx context.Context, filters []query.Filter) (int64, error)
}

type CreatePatientInput struct {
	CreateEntityInput
	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *int
	History *string
}

type UpdatePatientInput struct {
	UpdateEntityInput
	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *int
	History *string
}

type IPatientService interface {
	CreateNewPatient(ctx context.Context, input CreatePatientInput) (*model.Patient, error)
	GetPatientByID(ctx context.Context, patientID string) (*model.Patient, error)
	GetPatientsByWorkspaceID(ctx context.Context, workspaceID string, paginationOpts *query.Pagination) (*query.Result[*model.Patient], error)
	ListPatients(ctx context.Context, paginationOpts *query.Pagination) (*query.Result[*model.Patient], error)
	DeletePatientByID(ctx context.Context, patientId string) error
	UpdatePatient(ctx context.Context, patientID string, input UpdatePatientInput) error
	TransferPatientWorkspace(ctx context.Context, patientID string, newWorkspaceID string) error
	CascadeDelete(ctx context.Context, patientID string) error
	BatchDelete(ctx context.Context, patientIDs []string) error
	BatchTransfer(ctx context.Context, patientIDs []string, newWorkspaceID string) error
	CountPatients(ctx context.Context, filters []query.Filter) (int64, error)
}

type UploadImageInput struct {
	CreateEntityInput
	ContentType string
	Format      string
	Width       *int
	Height      *int
	Size        *int64
}

type ConfirmUploadInput struct {
	CreateEntityInput
	Format     string
	Width      *int
	Height     *int
	Size       *int64
	Status     model.ImageStatus
	OriginPath string
}

type IImageService interface {
	UploadImage(ctx context.Context, input *UploadImageInput) (*SignedURLPayload, error)
	ConfirmUpload(ctx context.Context, input *ConfirmUploadInput) error
	GetImageByID(ctx context.Context, imageID string) (*model.Image, error)
	ListImageByPatientID(ctx context.Context, patientID string, pagination *query.Pagination) (*query.Result[*model.Image], error)
	DeleteImageByID(ctx context.Context, imageID string) error
	BatchDeleteImages(ctx context.Context, imageIDs []string) error
	BatchTransferImages(ctx context.Context, imageIDs []string, newPatientID string) error
	CountImages(ctx context.Context, filters []query.Filter) (int64, error)
	TransferImage(ctx context.Context, imageID string, newPatientID string) error
}

type CreateAnnotationInput struct {
	CreateEntityInput
	vobj.TagValue
	Polygon *[]vobj.Point
}

type UpdateTagValue struct {
	TagType *vobj.TagType
	TagName *string
	Value   *any
	Color   *string
	Global  *bool
}
type UpdateAnnotationInput struct {
	UpdateEntityInput
	Polygon *[]vobj.Point
	UpdateTagValue
}

type IAnnotationService interface {
	CreateNewAnnotation(ctx context.Context, input *CreateAnnotationInput) (*model.Annotation, error)
	GetAnnotationByID(ctx context.Context, id string) (*model.Annotation, error)
	GetAnnotationsByImageID(ctx context.Context, imageID string, pagination *query.Pagination) (*query.Result[*model.Annotation], error)
	UpdateAnnotation(ctx context.Context, id string, input *UpdateAnnotationInput) error
	DeleteAnnotation(ctx context.Context, id string) error
	BatchDeleteAnnotations(ctx context.Context, ids []string) error
	CountAnnotations(ctx context.Context, filters []query.Filter) (int64, error)
}

type CreateAnnotationTypeInput struct {
	CreateEntityInput
	Tag vobj.Tag
}

type UpdateAnnotationTypeInput struct {
	UpdateEntityInput
	Tag *vobj.Tag
}

type IAnnotationTypeService interface {
	CreateNewAnnotationType(ctx context.Context, input *CreateAnnotationTypeInput) (*model.AnnotationType, error)
	GetAnnotationTypeByID(ctx context.Context, id string) (*model.AnnotationType, error)
	ListAnnotationTypes(ctx context.Context, paginationOpts *query.Pagination) (*query.Result[*model.AnnotationType], error)
	UpdateAnnotationType(ctx context.Context, id string, input *UpdateAnnotationTypeInput) error
	DeleteAnnotationType(ctx context.Context, id string) error
	CountAnnotationTypes(ctx context.Context, filters []query.Filter) (int64, error)
	BatchDeleteAnnotationTypes(ctx context.Context, ids []string) error
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
