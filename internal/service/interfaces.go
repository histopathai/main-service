package service

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/storage"
	"github.com/histopathai/main-service/internal/shared/query"
)

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

type IImageService interface {
	UploadImage(ctx context.Context, input *UploadImageInput) (*storage.SignedURLPayload, error)
	ConfirmUpload(ctx context.Context, input *ConfirmUploadInput) error
	GetImageByID(ctx context.Context, imageID string) (*model.Image, error)
	ListImageByPatientID(ctx context.Context, patientID string, pagination *query.Pagination) (*query.Result[*model.Image], error)
	DeleteImageByID(ctx context.Context, imageID string) error
	BatchDeleteImages(ctx context.Context, imageIDs []string) error
	BatchTransferImages(ctx context.Context, imageIDs []string, newPatientID string) error
	CountImages(ctx context.Context, filters []query.Filter) (int64, error)
	TransferImage(ctx context.Context, imageID string, newPatientID string) error
}

type IAnnotationService interface {
	CreateNewAnnotation(ctx context.Context, input *CreateAnnotationInput) (*model.Annotation, error)
	GetAnnotationByID(ctx context.Context, id string) (*model.Annotation, error)
	GetAnnotationsByImageID(ctx context.Context, imageID string, pagination *query.Pagination) (*query.Result[*model.Annotation], error)
	DeleteAnnotation(ctx context.Context, id string) error
	BatchDeleteAnnotations(ctx context.Context, ids []string) error
	CountAnnotations(ctx context.Context, filters []query.Filter) (int64, error)
}

type IAnnotationTypeService interface {
	ValidateAnnotationTypeCreation(ctx context.Context, input *CreateAnnotationTypeInput) error
	CreateNewAnnotationType(ctx context.Context, input *CreateAnnotationTypeInput) (*model.AnnotationType, error)
	GetAnnotationTypeByID(ctx context.Context, id string) (*model.AnnotationType, error)
	ListAnnotationTypes(ctx context.Context, paginationOpts *query.Pagination) (*query.Result[*model.AnnotationType], error)
	GetClassificationAnnotationTypes(ctx context.Context, paginationOpts *query.Pagination) (*query.Result[*model.AnnotationType], error)
	GetScoreAnnotationTypes(ctx context.Context, paginationOpts *query.Pagination) (*query.Result[*model.AnnotationType], error)
	UpdateAnnotationType(ctx context.Context, id string, input *UpdateAnnotationTypeInput) error
	DeleteAnnotationType(ctx context.Context, id string) error
	CountAnnotationTypes(ctx context.Context, filters []query.Filter) (int64, error)
	BatchDeleteAnnotationTypes(ctx context.Context, ids []string) error
}
