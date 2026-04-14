package usecase_test

import (
	"context"
	"testing"
	"time"

	appusecase "github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fakes
// ─────────────────────────────────────────────────────────────────────────────

type fakeAnnotationRepo struct {
	annotations map[string]*model.Annotation
	updateCalls map[string]map[string]interface{}
}

func newFakeAnnotationRepo(annotations ...*model.Annotation) *fakeAnnotationRepo {
	m := make(map[string]*model.Annotation)
	for _, a := range annotations {
		m[a.ID] = a
	}
	return &fakeAnnotationRepo{annotations: m, updateCalls: make(map[string]map[string]interface{})}
}

func (r *fakeAnnotationRepo) Create(_ context.Context, e *model.Annotation) (*model.Annotation, error) {
	r.annotations[e.ID] = e
	return e, nil
}
func (r *fakeAnnotationRepo) Read(_ context.Context, id string) (*model.Annotation, error) {
	return r.annotations[id], nil
}
func (r *fakeAnnotationRepo) Update(_ context.Context, id string, updates map[string]interface{}) error {
	r.updateCalls[id] = updates
	return nil
}
func (r *fakeAnnotationRepo) SoftDelete(_ context.Context, _ string) error        { return nil }
func (r *fakeAnnotationRepo) Transfer(_ context.Context, _, _ string) error       { return nil }
func (r *fakeAnnotationRepo) SoftDeleteMany(_ context.Context, _ []string) error  { return nil }
func (r *fakeAnnotationRepo) UpdateMany(_ context.Context, _ []string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeAnnotationRepo) TransferMany(_ context.Context, _ []string, _ string) error { return nil }
func (r *fakeAnnotationRepo) Delete(_ context.Context, _ string) error                   { return nil }
func (r *fakeAnnotationRepo) Find(_ context.Context, _ query.Specification) (*query.Result[*model.Annotation], error) {
	return nil, nil
}
func (r *fakeAnnotationRepo) Count(_ context.Context, _ query.Specification) (int64, error) {
	return 0, nil
}

type fakeAnnotationReviewRepo struct {
	reviews    map[string]*model.AnnotationReview
	softDeleted []string
}

func newFakeAnnotationReviewRepo(reviews ...*model.AnnotationReview) *fakeAnnotationReviewRepo {
	m := make(map[string]*model.AnnotationReview)
	for _, r := range reviews {
		m[r.ID] = r
	}
	return &fakeAnnotationReviewRepo{reviews: m}
}

func (r *fakeAnnotationReviewRepo) Create(_ context.Context, e *model.AnnotationReview) (*model.AnnotationReview, error) {
	r.reviews[e.ID] = e
	return e, nil
}
func (r *fakeAnnotationReviewRepo) Read(_ context.Context, id string) (*model.AnnotationReview, error) {
	return r.reviews[id], nil
}
func (r *fakeAnnotationReviewRepo) Update(_ context.Context, _ string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeAnnotationReviewRepo) SoftDelete(_ context.Context, id string) error {
	r.softDeleted = append(r.softDeleted, id)
	return nil
}
func (r *fakeAnnotationReviewRepo) Transfer(_ context.Context, _, _ string) error       { return nil }
func (r *fakeAnnotationReviewRepo) SoftDeleteMany(_ context.Context, _ []string) error  { return nil }
func (r *fakeAnnotationReviewRepo) UpdateMany(_ context.Context, _ []string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeAnnotationReviewRepo) TransferMany(_ context.Context, _ []string, _ string) error {
	return nil
}
func (r *fakeAnnotationReviewRepo) Delete(_ context.Context, _ string) error { return nil }
func (r *fakeAnnotationReviewRepo) Find(_ context.Context, _ query.Specification) (*query.Result[*model.AnnotationReview], error) {
	return nil, nil
}
func (r *fakeAnnotationReviewRepo) Count(_ context.Context, _ query.Specification) (int64, error) {
	return 0, nil
}

type fakeUoW struct {
	annotationRepo       *fakeAnnotationRepo
	annotationReviewRepo *fakeAnnotationReviewRepo
}

func newFakeUoW(ar *fakeAnnotationRepo, arr *fakeAnnotationReviewRepo) *fakeUoW {
	return &fakeUoW{annotationRepo: ar, annotationReviewRepo: arr}
}

func (u *fakeUoW) WithTx(_ context.Context, fn func(context.Context) error) error {
	return fn(context.Background())
}
func (u *fakeUoW) GetWorkspaceRepo() port.WorkspaceRepository               { return nil }
func (u *fakeUoW) GetPatientRepo() port.PatientRepository                   { return nil }
func (u *fakeUoW) GetImageRepo() port.ImageRepository                       { return nil }
func (u *fakeUoW) GetContentRepo() port.ContentRepository                   { return nil }
func (u *fakeUoW) GetAnnotationRepo() port.AnnotationRepository             { return u.annotationRepo }
func (u *fakeUoW) GetAnnotationReviewRepo() port.AnnotationReviewRepository { return u.annotationReviewRepo }
func (u *fakeUoW) GetAnnotationTypeRepo() port.AnnotationTypeRepository     { return nil }

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func makeAnnotation(id, creatorID string, reviewIDs []string) *model.Annotation {
	return &model.Annotation{
		Entity: vobj.Entity{
			ID:         id,
			CreatorID:  creatorID,
			EntityType: vobj.EntityTypeAnnotation,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		AnnotationTypeID: "at-1",
		WsID:             "ws-1",
		TagType:          vobj.NumberTag,
		Value:            5.0,
		IsGlobal:         true,
		Resource:         fields.AnnotationResourceModel,
		ReviewIDs:        reviewIDs,
	}
}

func makeReview(id, reviewerID, annotationID string) *model.AnnotationReview {
	parent, _ := vobj.NewParentRef(annotationID, vobj.ParentTypeAnnotation)
	return &model.AnnotationReview{
		Entity: vobj.Entity{
			ID:         id,
			CreatorID:  reviewerID,
			EntityType: vobj.EntityTypeAnnotationReview,
			Parent:     *parent,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		ReviewerID: reviewerID,
		Status:     fields.ReviewStatusApproved,
		ReviewedAt: time.Now(),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Delete — ownership
// ─────────────────────────────────────────────────────────────────────────────

func TestAnnotationReviewUseCase_Delete_OwnReview_Success(t *testing.T) {
	annotation := makeAnnotation("anno-1", "creator-1", []string{"review-1"})
	review := makeReview("review-1", "reviewer-A", "anno-1")

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo(review)
	uow := newFakeUoW(ar, arr)
	uc := appusecase.NewAnnotationReviewUseCase(arr, uow)

	err := uc.Delete(context.Background(), "review-1", "reviewer-A")
	require.NoError(t, err)

	// review should be soft-deleted
	assert.Contains(t, arr.softDeleted, "review-1")

	// ReviewIDs should be cleaned up
	assert.Empty(t, ar.updateCalls["anno-1"][fields.AnnotationReviewIDs.DomainName()])
}

func TestAnnotationReviewUseCase_Delete_OtherUserReview_Forbidden(t *testing.T) {
	annotation := makeAnnotation("anno-1", "creator-1", []string{"review-1"})
	review := makeReview("review-1", "reviewer-A", "anno-1")

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo(review)
	uow := newFakeUoW(ar, arr)
	uc := appusecase.NewAnnotationReviewUseCase(arr, uow)

	// user-B tries to delete reviewer-A's review
	err := uc.Delete(context.Background(), "review-1", "user-B")

	require.Error(t, err)
	var appErr *errors.Err
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, errors.ErrorTypeForbidden, appErr.Type)

	// review must NOT be soft-deleted
	assert.Empty(t, arr.softDeleted)
}

func TestAnnotationReviewUseCase_Delete_ReviewNotFound(t *testing.T) {
	ar := newFakeAnnotationRepo()
	arr := newFakeAnnotationReviewRepo() // empty
	uow := newFakeUoW(ar, arr)
	uc := appusecase.NewAnnotationReviewUseCase(arr, uow)

	err := uc.Delete(context.Background(), "nonexistent", "reviewer-A")

	require.Error(t, err)
	var appErr *errors.Err
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, errors.ErrorTypeNotFound, appErr.Type)
}

func TestAnnotationReviewUseCase_Delete_CleansUpReviewIDFromAnnotation(t *testing.T) {
	// Annotation has multiple reviews, only one gets deleted
	annotation := makeAnnotation("anno-1", "creator-1", []string{"review-1", "review-2", "review-3"})
	review := makeReview("review-2", "reviewer-B", "anno-1")

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo(review)
	uow := newFakeUoW(ar, arr)
	uc := appusecase.NewAnnotationReviewUseCase(arr, uow)

	err := uc.Delete(context.Background(), "review-2", "reviewer-B")
	require.NoError(t, err)

	// The annotation's ReviewIDs should no longer contain "review-2"
	updatedIDs, ok := ar.updateCalls["anno-1"][fields.AnnotationReviewIDs.DomainName()].([]string)
	require.True(t, ok, "ReviewIDs update should have been called")
	assert.NotContains(t, updatedIDs, "review-2")
	assert.Contains(t, updatedIDs, "review-1")
	assert.Contains(t, updatedIDs, "review-3")
	assert.Len(t, updatedIDs, 2)
}
