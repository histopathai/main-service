package validator_test

import (
	"context"
	"testing"
	"time"

	"github.com/histopathai/main-service/internal/application/usecase/validator"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/stretchr/testify/assert"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fixtures
// ─────────────────────────────────────────────────────────────────────────────

func makeAnnotationType(id string, tagType vobj.TagType, required bool, opts []string, min, max *float64) *model.AnnotationType {
	return &model.AnnotationType{
		Entity:     vobj.Entity{ID: id, CreatorID: "admin", EntityType: vobj.EntityTypeAnnotationType, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		TagType:    tagType,
		IsRequired: required,
		Options:    opts,
		Min:        min,
		Max:        max,
	}
}

func makeAnnotation(id, creatorID, annotationTypeID string, resource fields.AnnotationResourceField, value any) *model.Annotation {
	return &model.Annotation{
		Entity: vobj.Entity{
			ID:         id,
			CreatorID:  creatorID,
			EntityType: vobj.EntityTypeAnnotation,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		AnnotationTypeID: annotationTypeID,
		WsID:             "ws-1",
		TagType:          vobj.NumberTag,
		Value:            value,
		IsGlobal:         true,
		Resource:         resource,
	}
}

func makeReview(annotationID, reviewerID string, status fields.ReviewStatusField, modVal any, modPoly *[]vobj.Point) *model.AnnotationReview {
	parent, _ := vobj.NewParentRef(annotationID, vobj.ParentTypeAnnotation)
	return &model.AnnotationReview{
		Entity: vobj.Entity{
			ID:         "review-1",
			CreatorID:  reviewerID,
			EntityType: vobj.EntityTypeAnnotationReview,
			Parent:     *parent,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		ReviewerID:      reviewerID,
		Status:          status,
		ReviewedAt:      time.Now(),
		ModifiedValue:   modVal,
		ModifiedPolygon: modPoly,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidateCreate — access control
// ─────────────────────────────────────────────────────────────────────────────

func TestAnnotationReviewValidator_ValidateCreate_AnnotationNotFound(t *testing.T) {
	ar := newFakeAnnotationRepo() // empty — no annotations
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo()
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	review := makeReview("nonexistent-anno", "reviewer-1", fields.ReviewStatusApproved, nil, nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.Error(t, err, "should error when annotation does not exist")
}

func TestAnnotationReviewValidator_ValidateCreate_CannotReviewOwnManualAnnotation(t *testing.T) {
	min, max := 0.0, 10.0
	at := makeAnnotationType("at-1", vobj.NumberTag, false, nil, &min, &max)
	annotation := makeAnnotation("anno-1", "user-A", "at-1", fields.AnnotationResourceManual, 5.0)

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	// Same user (user-A) trying to review their own manual annotation
	review := makeReview("anno-1", "user-A", fields.ReviewStatusApproved, nil, nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.Error(t, err, "should reject when reviewer == creator && resource == manual")
}

func TestAnnotationReviewValidator_ValidateCreate_SameUserCanReviewOwnModelAnnotation(t *testing.T) {
	min, max := 0.0, 10.0
	at := makeAnnotationType("at-1", vobj.NumberTag, false, nil, &min, &max)
	// resource = model — so same user CAN review it
	annotation := makeAnnotation("anno-1", "user-A", "at-1", fields.AnnotationResourceModel, 5.0)

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	review := makeReview("anno-1", "user-A", fields.ReviewStatusApproved, nil, nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.NoError(t, err)
}

func TestAnnotationReviewValidator_ValidateCreate_DifferentUserCanReviewManualAnnotation(t *testing.T) {
	min, max := 0.0, 10.0
	at := makeAnnotationType("at-1", vobj.NumberTag, false, nil, &min, &max)
	annotation := makeAnnotation("anno-1", "user-A", "at-1", fields.AnnotationResourceManual, 5.0)

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	// Different reviewer
	review := makeReview("anno-1", "user-B", fields.ReviewStatusApproved, nil, nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidateCreate — Modified status value validation
// ─────────────────────────────────────────────────────────────────────────────

func TestAnnotationReviewValidator_ValidateCreate_ModifiedValue_OutOfRange(t *testing.T) {
	min, max := 0.0, 10.0
	at := makeAnnotationType("at-1", vobj.NumberTag, false, nil, &min, &max)
	annotation := makeAnnotation("anno-1", "user-A", "at-1", fields.AnnotationResourceModel, 5.0)

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	// Modified value exceeds max (10)
	review := makeReview("anno-1", "user-B", fields.ReviewStatusModified, float64(99), nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.Error(t, err, "should reject out-of-range modified value")
}

func TestAnnotationReviewValidator_ValidateCreate_ModifiedValue_InRange(t *testing.T) {
	min, max := 0.0, 10.0
	at := makeAnnotationType("at-1", vobj.NumberTag, false, nil, &min, &max)
	annotation := makeAnnotation("anno-1", "user-A", "at-1", fields.AnnotationResourceModel, 5.0)

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	review := makeReview("anno-1", "user-B", fields.ReviewStatusModified, float64(7), nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.NoError(t, err)
}

func TestAnnotationReviewValidator_ValidateCreate_ModifiedSelectValue_InvalidOption(t *testing.T) {
	at := makeAnnotationType("at-2", vobj.SelectTag, false, []string{"a", "b", "c"}, nil, nil)
	annotation := makeAnnotation("anno-2", "user-A", "at-2", fields.AnnotationResourceImported, "a")
	annotation.TagType = vobj.SelectTag

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	review := makeReview("anno-2", "user-B", fields.ReviewStatusModified, "invalid_option", nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.Error(t, err, "should reject invalid select option in modified value")
}

func TestAnnotationReviewValidator_ValidateCreate_ModifiedSelectValue_ValidOption(t *testing.T) {
	at := makeAnnotationType("at-2", vobj.SelectTag, false, []string{"a", "b", "c"}, nil, nil)
	annotation := makeAnnotation("anno-2", "user-A", "at-2", fields.AnnotationResourceImported, "a")
	annotation.TagType = vobj.SelectTag

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	review := makeReview("anno-2", "user-B", fields.ReviewStatusModified, "b", nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.NoError(t, err)
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidateCreate — Approved/Rejected (no modification checks)
// ─────────────────────────────────────────────────────────────────────────────

func TestAnnotationReviewValidator_ValidateCreate_Approved_NoModification(t *testing.T) {
	min, max := 0.0, 10.0
	at := makeAnnotationType("at-1", vobj.NumberTag, false, nil, &min, &max)
	annotation := makeAnnotation("anno-1", "user-A", "at-1", fields.AnnotationResourceImported, 5.0)

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	review := makeReview("anno-1", "user-B", fields.ReviewStatusApproved, nil, nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.NoError(t, err)
}

func TestAnnotationReviewValidator_ValidateCreate_Rejected_NoModification(t *testing.T) {
	min, max := 0.0, 10.0
	at := makeAnnotationType("at-1", vobj.NumberTag, false, nil, &min, &max)
	annotation := makeAnnotation("anno-1", "user-A", "at-1", fields.AnnotationResourceModel, 5.0)

	ar := newFakeAnnotationRepo(annotation)
	arr := newFakeAnnotationReviewRepo()
	atr := newFakeAnnotationTypeRepo(at)
	uow := newFakeUoW(ar, arr, atr)

	v := validator.NewAnnotationReviewValidator(arr, uow)
	review := makeReview("anno-1", "user-B", fields.ReviewStatusRejected, nil, nil)

	err := v.ValidateCreate(context.Background(), review)
	assert.NoError(t, err)
}
