package queries

import (
	"context"
	stderrors "errors"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ImageQuery struct {
	*BaseQuery[*model.Image]
	*HierarchicalQueries[*model.Image]

	tissueMaskRepo port.TissueMaskRepository
}

func NewImageQuery(repo port.ImageRepository, tissueMaskRepo port.TissueMaskRepository) *ImageQuery {
	return &ImageQuery{
		BaseQuery: &BaseQuery[*model.Image]{
			repo: repo,
		},
		HierarchicalQueries: &HierarchicalQueries[*model.Image]{
			repo: repo,
		},
		tissueMaskRepo: tissueMaskRepo,
	}
}

// SoftDelete soft deletes the image and its tissue mask. ML reads the
// tissue_masks collection straight from Firestore, so a mask left behind would
// keep a deleted image in the training set. The mask is written first: if that
// fails the image stays visible and the call can be retried.
func (s *ImageQuery) SoftDelete(ctx context.Context, id string) error {
	if err := s.softDeleteTissueMasks(ctx, []string{id}); err != nil {
		return err
	}
	return s.BaseQuery.SoftDelete(ctx, id)
}

// SoftDeleteMany soft deletes the images and their tissue masks, see SoftDelete.
func (s *ImageQuery) SoftDeleteMany(ctx context.Context, ids []string) error {
	if err := s.softDeleteTissueMasks(ctx, ids); err != nil {
		return err
	}
	return s.BaseQuery.SoftDeleteMany(ctx, ids)
}

// softDeleteTissueMasks soft deletes tissue_masks/{image_id} for the images
// that have one. Repository writes merge into the document and would create it
// when missing, so images without a mask are read first and skipped.
func (s *ImageQuery) softDeleteTissueMasks(ctx context.Context, imageIDs []string) error {
	existing := make([]string, 0, len(imageIDs))
	for _, id := range imageIDs {
		mask, err := s.tissueMaskRepo.Read(ctx, id)
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return err
		}
		if mask == nil {
			continue
		}
		existing = append(existing, id)
	}

	if len(existing) == 0 {
		return nil
	}
	return s.tissueMaskRepo.SoftDeleteMany(ctx, existing)
}

func isNotFound(err error) bool {
	var appErr *errors.Err
	return stderrors.As(err, &appErr) && appErr.Type == errors.ErrorTypeNotFound
}
