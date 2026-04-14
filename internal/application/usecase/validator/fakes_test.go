package validator_test

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fake AnnotationRepository
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

func (r *fakeAnnotationRepo) Create(_ context.Context, entity *model.Annotation) (*model.Annotation, error) {
	r.annotations[entity.ID] = entity
	return entity, nil
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

// ─────────────────────────────────────────────────────────────────────────────
// Fake AnnotationReviewRepository
// ─────────────────────────────────────────────────────────────────────────────

type fakeAnnotationReviewRepo struct {
	reviews map[string]*model.AnnotationReview
}

func newFakeAnnotationReviewRepo() *fakeAnnotationReviewRepo {
	return &fakeAnnotationReviewRepo{reviews: make(map[string]*model.AnnotationReview)}
}

func (r *fakeAnnotationReviewRepo) Create(_ context.Context, entity *model.AnnotationReview) (*model.AnnotationReview, error) {
	r.reviews[entity.ID] = entity
	return entity, nil
}
func (r *fakeAnnotationReviewRepo) Read(_ context.Context, id string) (*model.AnnotationReview, error) {
	return r.reviews[id], nil
}
func (r *fakeAnnotationReviewRepo) Update(_ context.Context, _ string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeAnnotationReviewRepo) SoftDelete(_ context.Context, _ string) error        { return nil }
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

// ─────────────────────────────────────────────────────────────────────────────
// Fake AnnotationTypeRepository
// ─────────────────────────────────────────────────────────────────────────────

type fakeAnnotationTypeRepo struct {
	types map[string]*model.AnnotationType
}

func newFakeAnnotationTypeRepo(types ...*model.AnnotationType) *fakeAnnotationTypeRepo {
	m := make(map[string]*model.AnnotationType)
	for _, at := range types {
		m[at.ID] = at
	}
	return &fakeAnnotationTypeRepo{types: m}
}

func (r *fakeAnnotationTypeRepo) Create(_ context.Context, entity *model.AnnotationType) (*model.AnnotationType, error) {
	return entity, nil
}
func (r *fakeAnnotationTypeRepo) Read(_ context.Context, id string) (*model.AnnotationType, error) {
	return r.types[id], nil
}
func (r *fakeAnnotationTypeRepo) Update(_ context.Context, _ string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeAnnotationTypeRepo) SoftDelete(_ context.Context, _ string) error        { return nil }
func (r *fakeAnnotationTypeRepo) Transfer(_ context.Context, _, _ string) error       { return nil }
func (r *fakeAnnotationTypeRepo) SoftDeleteMany(_ context.Context, _ []string) error  { return nil }
func (r *fakeAnnotationTypeRepo) UpdateMany(_ context.Context, _ []string, _ map[string]interface{}) error {
	return nil
}
func (r *fakeAnnotationTypeRepo) TransferMany(_ context.Context, _ []string, _ string) error {
	return nil
}
func (r *fakeAnnotationTypeRepo) Delete(_ context.Context, _ string) error { return nil }
func (r *fakeAnnotationTypeRepo) Find(_ context.Context, _ query.Specification) (*query.Result[*model.AnnotationType], error) {
	return nil, nil
}
func (r *fakeAnnotationTypeRepo) Count(_ context.Context, _ query.Specification) (int64, error) {
	return 0, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Fake UnitOfWorkFactory
// ─────────────────────────────────────────────────────────────────────────────

type fakeUoW struct {
	annotationRepo       *fakeAnnotationRepo
	annotationReviewRepo *fakeAnnotationReviewRepo
	annotationTypeRepo   *fakeAnnotationTypeRepo
}

func newFakeUoW(ar *fakeAnnotationRepo, arr *fakeAnnotationReviewRepo, atr *fakeAnnotationTypeRepo) *fakeUoW {
	return &fakeUoW{
		annotationRepo:       ar,
		annotationReviewRepo: arr,
		annotationTypeRepo:   atr,
	}
}

func (u *fakeUoW) WithTx(_ context.Context, fn func(context.Context) error) error {
	return fn(context.Background())
}
func (u *fakeUoW) GetWorkspaceRepo() port.WorkspaceRepository         { return nil }
func (u *fakeUoW) GetPatientRepo() port.PatientRepository             { return nil }
func (u *fakeUoW) GetImageRepo() port.ImageRepository                 { return nil }
func (u *fakeUoW) GetContentRepo() port.ContentRepository             { return nil }
func (u *fakeUoW) GetAnnotationRepo() port.AnnotationRepository       { return u.annotationRepo }
func (u *fakeUoW) GetAnnotationReviewRepo() port.AnnotationReviewRepository {
	return u.annotationReviewRepo
}
func (u *fakeUoW) GetAnnotationTypeRepo() port.AnnotationTypeRepository { return u.annotationTypeRepo }
