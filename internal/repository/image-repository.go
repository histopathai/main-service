package repository

import (
	"context"
	"time"

	"github.com/histopathai/models"
)

const ImagesCollection = "images"

type ImageRepository struct {
	repo Repository
}

func NewImageRepository(repo Repository) *ImageRepository {
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

func (ir *ImageRepository) QueryImages(ctx context.Context, filters map[string]interface{}) ([]*models.Image, error) {
	results, err := ir.repo.Query(ctx, ImagesCollection, filters)
	if err != nil {
		return nil, err
	}
	images := make([]*models.Image, 0, len(results))
	for _, data := range results {
		image := &models.Image{}
		image.FromMap(data)
		images = append(images, image)
	}
	return images, nil
}

func (ir *ImageRepository) Exists(ctx context.Context, imageID string) (bool, error) {
	return ir.repo.Exists(ctx, ImagesCollection, imageID)
}
