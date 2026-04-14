package command

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateAnnotationReviewCommand struct {
	CreateEntityCommand
	AnnotationID    string
	ReviewerID      string
	Status          string
	Comments        *string
	ModifiedPolygon *[]CommandPoint
	ModifiedValue   any
}

func (c *CreateAnnotationReviewCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if c.AnnotationID == "" {
		details["annotation_id"] = "AnnotationID is required"
	}
	if c.ReviewerID == "" {
		details["reviewer_id"] = "ReviewerID is required"
	}
	
	if c.Status == "" {
		details["status"] = "Status is required"
	} else {
		status := fields.ReviewStatusField(c.Status)
		if !status.IsValid() {
			details["status"] = "Invalid status"
		}
	}

	if c.Status == string(fields.ReviewStatusModified) {
		if c.ModifiedValue == nil && c.ModifiedPolygon == nil {
			details["modified"] = "ModifiedValue or ModifiedPolygon is required when status is modified"
		}
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *CreateAnnotationReviewCommand) ToEntity() (*model.AnnotationReview, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	// Make sure EntityType is annotation_review
	c.CreateEntityCommand.EntityType = string(vobj.EntityTypeAnnotationReview)
	// Parent of Review is Annotation
	c.CreateEntityCommand.ParentID = c.AnnotationID
	c.CreateEntityCommand.ParentType = string(vobj.ParentTypeAnnotation)
	c.CreateEntityCommand.Name = "Review for " + c.AnnotationID
	// Reviewer is the creator of this Review entity
	c.CreateEntityCommand.CreatorID = c.ReviewerID

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	var points *[]vobj.Point
	if c.ModifiedPolygon != nil {
		pList := make([]vobj.Point, len(*c.ModifiedPolygon))
		for i, p := range *c.ModifiedPolygon {
			pList[i] = vobj.Point{X: p.X, Y: p.Y}
		}
		points = &pList
	}

	status := fields.ReviewStatusField(c.Status)

	annotationReviewEntity := model.AnnotationReview{
		Entity:          *baseEntity, // Entity.Parent.ID = AnnotationID, Entity.Parent.Type = annotation
		ReviewerID:      c.ReviewerID,
		Status:          status,
		Comments:        c.Comments,
		ReviewedAt:      time.Now(),
		ModifiedValue:   c.ModifiedValue,
		ModifiedPolygon: points,
	}

	return &annotationReviewEntity, nil
}
