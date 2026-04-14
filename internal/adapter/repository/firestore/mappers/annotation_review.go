package mappers

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationReviewMapper struct {
	*EntityMapper[*model.AnnotationReview]
}

func NewAnnotationReviewMapper() *AnnotationReviewMapper {
	return &AnnotationReviewMapper{
		EntityMapper: NewEntityMapper[*model.AnnotationReview](),
	}
}

func (m *AnnotationReviewMapper) ToFirestoreMap(entity *model.AnnotationReview) map[string]interface{} {
	doc := m.EntityMapper.ToFirestoreMap(entity)

	doc["reviewer_id"] = entity.ReviewerID
	doc["status"] = entity.Status.FirestoreName()
	doc["reviewed_at"] = entity.ReviewedAt

	if entity.Comments != nil {
		doc["comments"] = *entity.Comments
	}
	if entity.ModifiedPolygon != nil {
		doc["modified_polygon"] = vobj.ToJSONPoints(*entity.ModifiedPolygon)
	}
	if entity.ModifiedValue != nil {
		doc["modified_value"] = entity.ModifiedValue
	}

	return doc
}

func (m *AnnotationReviewMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.AnnotationReview, error) {
	entity, err := m.EntityMapper.ParseEntity(doc)
	if err != nil {
		return nil, err
	}

	review := &model.AnnotationReview{
		Entity: *entity,
	}

	data := doc.Data()

	if v, ok := data["reviewer_id"].(string); ok {
		review.ReviewerID = v
	}
	if v, ok := data["status"].(string); ok {
		review.Status = fields.ReviewStatusField(v)
	}
	if v, ok := data["reviewed_at"].(time.Time); ok {
		review.ReviewedAt = v
	}
	if v, ok := data["comments"].(string); ok {
		review.Comments = &v
	}
	if polygonRaw, ok := data["modified_polygon"].([]interface{}); ok {
		jsonPoints := make([]map[string]float64, 0, len(polygonRaw))
		for _, p := range polygonRaw {
			if pointMap, ok := p.(map[string]interface{}); ok {
				jsonPoint := make(map[string]float64)
				if x, ok := pointMap["X"].(float64); ok {
					jsonPoint["X"] = x
				}
				if y, ok := pointMap["Y"].(float64); ok {
					jsonPoint["Y"] = y
				}
				jsonPoints = append(jsonPoints, jsonPoint)
			}
		}
		points := vobj.FromJSONPoints(jsonPoints)
		review.ModifiedPolygon = &points
	}
	if v, ok := data["modified_value"]; ok {
		review.ModifiedValue = v
	}

	return review, nil
}

func (m *AnnotationReviewMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mappedUpdates, err := m.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for k, v := range updates {
		switch k {
		case "Status":
			if status, ok := v.(fields.ReviewStatusField); ok {
				mappedUpdates["status"] = status.FirestoreName()
			} else if statusStr, ok := v.(string); ok {
				mappedUpdates["status"] = statusStr
			} else {
				return nil, errors.NewValidationError("invalid type for status field", nil)
			}
		case "Comments":
			if comments, ok := v.(*string); ok && comments != nil {
				mappedUpdates["comments"] = *comments
			} else if commentStr, ok := v.(string); ok {
				mappedUpdates["comments"] = commentStr
			}
		case "ModifiedPolygon":
			if polygon, ok := v.(*[]vobj.Point); ok {
				mappedUpdates["modified_polygon"] = vobj.ToJSONPoints(*polygon)
			} else if points, ok := v.([]vobj.Point); ok {
				mappedUpdates["modified_polygon"] = vobj.ToJSONPoints(points)
			} else {
				return nil, errors.NewValidationError("invalid type for modified_polygon field", nil)
			}
		case "ModifiedValue":
			mappedUpdates["modified_value"] = v
		}
	}

	return mappedUpdates, nil
}

func (m *AnnotationReviewMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters, err := m.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	for _, f := range filters {
		switch f.Field {
		case "annotation_id", "AnnotationID":
			mappedFilters = append(mappedFilters, query.Filter{
				Field: "annotation_id", Operator: f.Operator, Value: f.Value,
			})
		case "reviewer_id", "ReviewerID":
			mappedFilters = append(mappedFilters, query.Filter{
				Field: "reviewer_id", Operator: f.Operator, Value: f.Value,
			})
		case "status", "Status":
			mappedFilters = append(mappedFilters, query.Filter{
				Field: "status", Operator: f.Operator, Value: f.Value,
			})
		}
	}

	return mappedFilters, nil
}
