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

type ImageUseCase interface {
	Upload(ctx context.Context, cmd command.UploadImageCommand) (*PresignedURLPayload, error)
	Update(ctx context.Context, cmd command.UpdateImageCommand) error
	Transfer(ctx context.Context, cmd command.TransferCommand) error
	TransferMany(ctx context.Context, cmd command.TransferManyCommand) error
}

type ContentUseCase interface {
	Upload(ctx context.Context, cmd command.UploadContentCommand) (*PresignedURLPayload, error)
}
