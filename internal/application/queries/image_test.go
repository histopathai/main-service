package queries_test

import (
	"context"
	"errors"
	"testing"

	"github.com/histopathai/main-service/internal/application/queries"
	"github.com/histopathai/main-service/internal/domain/model"
	apperrors "github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fake ImageRepository
// ─────────────────────────────────────────────────────────────────────────────

type fakeImageRepo struct {
	softDeleted   []string
	softDeleteErr error
}

func (r *fakeImageRepo) SoftDelete(_ context.Context, id string) error {
	if r.softDeleteErr != nil {
		return r.softDeleteErr
	}
	r.softDeleted = append(r.softDeleted, id)
	return nil
}

func (r *fakeImageRepo) SoftDeleteMany(_ context.Context, ids []string) error {
	if r.softDeleteErr != nil {
		return r.softDeleteErr
	}
	r.softDeleted = append(r.softDeleted, ids...)
	return nil
}

func (r *fakeImageRepo) Create(_ context.Context, entity *model.Image) (*model.Image, error) {
	return entity, nil
}
func (r *fakeImageRepo) Read(_ context.Context, _ string) (*model.Image, error) { return nil, nil }
func (r *fakeImageRepo) Update(_ context.Context, _ string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeImageRepo) Transfer(_ context.Context, _, _ string) error { return nil }
func (r *fakeImageRepo) UpdateMany(_ context.Context, _ []string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeImageRepo) TransferMany(_ context.Context, _ []string, _ string) error { return nil }
func (r *fakeImageRepo) Delete(_ context.Context, _ string) error                   { return nil }
func (r *fakeImageRepo) Find(_ context.Context, _ query.Specification) (*query.Result[*model.Image], error) {
	return nil, nil
}
func (r *fakeImageRepo) Count(_ context.Context, _ query.Specification) (int64, error) {
	return 0, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Fake TissueMaskRepository
// ─────────────────────────────────────────────────────────────────────────────

type fakeTissueMaskRepo struct {
	masks         map[string]*model.TissueMask
	readErr       error
	softDeleteErr error
	softDeleted   []string
}

func newFakeTissueMaskRepo(imageIDs ...string) *fakeTissueMaskRepo {
	masks := make(map[string]*model.TissueMask)
	for _, id := range imageIDs {
		mask := &model.TissueMask{}
		mask.SetID(id)
		masks[id] = mask
	}
	return &fakeTissueMaskRepo{masks: masks}
}

func (r *fakeTissueMaskRepo) Read(_ context.Context, id string) (*model.TissueMask, error) {
	if r.readErr != nil {
		return nil, r.readErr
	}
	mask, ok := r.masks[id]
	if !ok {
		return nil, apperrors.NewNotFoundError("document not found")
	}
	return mask, nil
}

func (r *fakeTissueMaskRepo) SoftDeleteMany(_ context.Context, ids []string) error {
	if r.softDeleteErr != nil {
		return r.softDeleteErr
	}
	r.softDeleted = append(r.softDeleted, ids...)
	return nil
}

func (r *fakeTissueMaskRepo) SoftDelete(_ context.Context, id string) error {
	if r.softDeleteErr != nil {
		return r.softDeleteErr
	}
	r.softDeleted = append(r.softDeleted, id)
	return nil
}

func (r *fakeTissueMaskRepo) Create(_ context.Context, entity *model.TissueMask) (*model.TissueMask, error) {
	return entity, nil
}
func (r *fakeTissueMaskRepo) Update(_ context.Context, _ string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeTissueMaskRepo) Transfer(_ context.Context, _, _ string) error { return nil }
func (r *fakeTissueMaskRepo) UpdateMany(_ context.Context, _ []string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeTissueMaskRepo) TransferMany(_ context.Context, _ []string, _ string) error { return nil }
func (r *fakeTissueMaskRepo) Delete(_ context.Context, _ string) error                   { return nil }
func (r *fakeTissueMaskRepo) Find(_ context.Context, _ query.Specification) (*query.Result[*model.TissueMask], error) {
	return nil, nil
}
func (r *fakeTissueMaskRepo) Count(_ context.Context, _ query.Specification) (int64, error) {
	return 0, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestImageSoftDeleteCascadesToTissueMask(t *testing.T) {
	imageRepo := &fakeImageRepo{}
	maskRepo := newFakeTissueMaskRepo("img-1")
	q := queries.NewImageQuery(imageRepo, maskRepo)

	if err := q.SoftDelete(context.Background(), "img-1"); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	if len(maskRepo.softDeleted) != 1 || maskRepo.softDeleted[0] != "img-1" {
		t.Errorf("tissue mask not soft deleted, got %v", maskRepo.softDeleted)
	}
	if len(imageRepo.softDeleted) != 1 || imageRepo.softDeleted[0] != "img-1" {
		t.Errorf("image not soft deleted, got %v", imageRepo.softDeleted)
	}
}

func TestImageSoftDeleteWithoutTissueMask(t *testing.T) {
	imageRepo := &fakeImageRepo{}
	maskRepo := newFakeTissueMaskRepo()
	q := queries.NewImageQuery(imageRepo, maskRepo)

	if err := q.SoftDelete(context.Background(), "img-1"); err != nil {
		t.Fatalf("SoftDelete: %v", err)
	}

	// Writing a missing mask document would create it, so nothing may be written.
	if len(maskRepo.softDeleted) != 0 {
		t.Errorf("wrote a mask that does not exist: %v", maskRepo.softDeleted)
	}
	if len(imageRepo.softDeleted) != 1 {
		t.Errorf("image not soft deleted, got %v", imageRepo.softDeleted)
	}
}

func TestImageSoftDeleteManyCascadesOnlyToExistingMasks(t *testing.T) {
	imageRepo := &fakeImageRepo{}
	maskRepo := newFakeTissueMaskRepo("img-1", "img-3")
	q := queries.NewImageQuery(imageRepo, maskRepo)

	ids := []string{"img-1", "img-2", "img-3"}
	if err := q.SoftDeleteMany(context.Background(), ids); err != nil {
		t.Fatalf("SoftDeleteMany: %v", err)
	}

	if len(maskRepo.softDeleted) != 2 {
		t.Fatalf("expected 2 masks soft deleted, got %v", maskRepo.softDeleted)
	}
	for i, want := range []string{"img-1", "img-3"} {
		if maskRepo.softDeleted[i] != want {
			t.Errorf("mask %d: got %q, want %q", i, maskRepo.softDeleted[i], want)
		}
	}
	if len(imageRepo.softDeleted) != 3 {
		t.Errorf("expected 3 images soft deleted, got %v", imageRepo.softDeleted)
	}
}

func TestImageSoftDeleteKeepsImageWhenMaskWriteFails(t *testing.T) {
	imageRepo := &fakeImageRepo{}
	maskRepo := newFakeTissueMaskRepo("img-1")
	maskRepo.softDeleteErr = errors.New("firestore unavailable")
	q := queries.NewImageQuery(imageRepo, maskRepo)

	if err := q.SoftDelete(context.Background(), "img-1"); err == nil {
		t.Fatal("expected the mask write error to be returned")
	}

	// The image must stay visible so the call can be retried.
	if len(imageRepo.softDeleted) != 0 {
		t.Errorf("image was soft deleted despite the mask failure: %v", imageRepo.softDeleted)
	}
}

func TestImageSoftDeleteKeepsImageWhenMaskReadFails(t *testing.T) {
	imageRepo := &fakeImageRepo{}
	maskRepo := newFakeTissueMaskRepo("img-1")
	maskRepo.readErr = errors.New("firestore unavailable")
	q := queries.NewImageQuery(imageRepo, maskRepo)

	if err := q.SoftDelete(context.Background(), "img-1"); err == nil {
		t.Fatal("expected the mask read error to be returned")
	}
	if len(imageRepo.softDeleted) != 0 {
		t.Errorf("image was soft deleted despite the mask failure: %v", imageRepo.softDeleted)
	}
}
