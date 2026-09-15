package port

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/model"
)

type WorkspaceUseCase interface {
	Create(ctx context.Context, cmd command.CreateWorkspaceCommand) (*model.Workspace, error)
	Update(ctx context.Context, cmd command.UpdateWorkspaceCommand) error
}

type PatientUseCase interface {
	Create(ctx context.Context, cmd command.CreatePatientCommand) (*model.Patient, error)
	Update(ctx context.Context, cmd command.UpdatePatientCommand) error
	Transfer(ctx context.Context, cmd command.TransferCommand) error
	TransferMany(ctx context.Context, cmd command.TransferManyCommand) error
}

type AnnotationTypeUseCase interface {
	Create(ctx context.Context, cmd command.CreateAnnotationTypeCommand) (*model.AnnotationType, error)
	Update(ctx context.Context, cmd command.UpdateAnnotationTypeCommand) error
}

type AnnotationUseCase interface {
	Create(ctx context.Context, cmd command.CreateAnnotationCommand) (*model.Annotation, error)
	Update(ctx context.Context, cmd command.UpdateAnnotationCommand) error
}

type AnnotationReviewUseCase interface {
	Create(ctx context.Context, cmd command.CreateAnnotationReviewCommand) (*model.AnnotationReview, error)
	Delete(ctx context.Context, reviewID string, requesterID string) error
	Update(ctx context.Context, cmd command.UpdateAnnotationReviewCommand) error
}

type TissueMaskUseCase interface {
	Save(ctx context.Context, cmd command.SaveTissueMaskCommand) (*model.TissueMask, error)
	Approve(ctx context.Context, cmd command.ReviewTissueMaskCommand) (*model.TissueMask, error)
	Reject(ctx context.Context, cmd command.ReviewTissueMaskCommand) (*model.TissueMask, error)
	ApplyWorkerResult(ctx context.Context, cmd command.ApplyWorkerTissueMaskCommand) (bool, error)
}

type ImageUseCase interface {
	Upload(ctx context.Context, cmd command.UploadImageCommand) ([]PresignedURLPayload, error)
	Update(ctx context.Context, cmd command.UpdateImageCommand) error
	Transfer(ctx context.Context, cmd command.TransferCommand) error
	TransferMany(ctx context.Context, cmd command.TransferManyCommand) error
}
