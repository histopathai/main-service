package repository

import (
	"context"
	"time"

	"github.com/histopathai/models"
)

const ImagesCollection = "images"

type ImageQueryResult struct {
	Data    []models.Image
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

type ImageRepository struct {
	repo *MainRepository
}

func NewImageRepository(repo *MainRepository) *ImageRepository {
	return &ImageRepository{
		repo: repo,
	}
}

func (ir *ImageRepository) CreateImage(ctx context.Context, image *models.Image) (string, error) {
	image.CreatedAt = time.Now()
	image.UpdatedAt = time.Now()
	return ir.repo.Create(ctx, ImagesCollection, image.ToMap())
}

func (ir *ImageRepository) ReadImage(ctx context.Context, imageID string) (*models.Image, error) {
	data, err := ir.repo.Read(ctx, ImagesCollection, imageID)
	if err != nil {
		return nil, err
	}
	image := &models.Image{}
	image.FromMap(data)
	return image, nil
}

func (ir *ImageRepository) UpdateImage(ctx context.Context, imageID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return ir.repo.Update(ctx, ImagesCollection, imageID, updates)
}

func (ir *ImageRepository) DeleteImage(ctx context.Context, imageID string) error {
	return ir.repo.Delete(ctx, ImagesCollection, imageID)
}
func (ir *ImageRepository) ListImages(ctx context.Context, filters []Filter, pagination Pagination) (*ImageQueryResult, error) {
	result, err := ir.repo.List(ctx, ImagesCollection, filters, pagination)
	if err != nil {
		return nil, err
	}

	images := make([]models.Image, len(result.Data))
	for i, data := range result.Data {
		image := models.Image{}
		image.FromMap(data)
		images[i] = image
	}

	return &ImageQueryResult{
		Data:    images,
		Total:   result.Total,
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}, nil
}

func (ir *ImageRepository) GetImagesByPatientID(ctx context.Context, patientID string, pagination Pagination) (*ImageQueryResult, error) {
	filters := []Filter{
		{
			Field: "patient_id",
			Op:    OpEqual,
			Value: patientID,
		},
	}
	return ir.ListImages(ctx, filters, pagination)
}

func (ir *ImageRepository) GetImagesByCreatorID(ctx context.Context, creatorID string, pagination Pagination) (*ImageQueryResult, error) {
	filters := []Filter{
		{
			Field: "creator_id",
			Op:    OpEqual,
			Value: creatorID,
		},
	}
	return ir.ListImages(ctx, filters, pagination)
}

func (ir *ImageRepository) GetImagesByWorkspaceID(ctx context.Context, workspaceID string, pagination Pagination) (*ImageQueryResult, error) {
	filters := []Filter{
		{
			Field: "workspace_id",
			Op:    OpEqual,
			Value: workspaceID,
		},
	}
	return ir.ListImages(ctx, filters, pagination)
}

func (ir *ImageRepository) Exists(ctx context.Context, imageID string) (bool, error) {
	return ir.repo.Exists(ctx, ImagesCollection, imageID)
}
