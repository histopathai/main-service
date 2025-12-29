package firestoreMappers

import (
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationMapper struct{}

func (am *AnnotationMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Annotation, error) {
	fa := &model.Annotation{}

	data := doc.Data()

	if data == nil {
		return nil, fmt.Errorf("firestore document data is nil")
	}

	beMapper := &BaseEntityMapper{}
	baseEntity, _ := beMapper.FromFirestoreDoc(doc)

	if baseEntity == nil {
		return nil, fmt.Errorf("failed to map base entity from firestore document")
	}

	points := data["polygon"].([]any)
	fa.Polygon = make([]model.Point, len(points))
	for i, p := range points {
		pointMap := p.(map[string]any)
		fa.Polygon[i] = model.Point{
			X: pointMap["x"].(float64),
			Y: pointMap["y"].(float64),
		}
	}

	if description, ok := data["description"].(string); ok {
		fa.Description = &description
	}

	if score, ok := data["score"].(float64); ok {
		fa.Score = &score
	}

	if class, ok := data["class"].(string); ok {
		fa.Class = &class
	}

	return fa, nil
}

func (am *AnnotationMapper) ToFirestoreMap(a *model.Annotation) map[string]interface{} {
	beMapper := &BaseEntityMapper{}
	m := beMapper.ToFirestoreMap(&a.BaseEntity)

	points := make([]map[string]float64, len(a.Polygon))
	for i, p := range a.Polygon {
		points[i] = map[string]float64{
			"x": p.X,
			"y": p.Y,
		}
	}
	m["polygon"] = points

	if a.Description != nil {
		m["description"] = *a.Description
	}

	if a.Score != nil {
		m["score"] = *a.Score
	}
	if a.Class != nil {
		m["class"] = *a.Class
	}

	return m
}

func (am *AnnotationMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	beMapper := &BaseEntityMapper{}
	firestoreUpdates, _ := beMapper.MapUpdates(updates)

	for k, v := range updates {
		switch k {
		case "polygon":
			points := v.([]model.Point)
			firestorePoints := make([]map[string]float64, len(points))
			for i, p := range points {
				firestorePoints[i] = map[string]float64{
					"x": p.X,
					"y": p.Y,
				}
			}
			firestoreUpdates["polygon"] = firestorePoints
		case "description":
			desc := v.(string)
			firestoreUpdates["description"] = desc
		case "score":
			score := v.(float64)
			firestoreUpdates["score"] = score
		case "class":
			class := v.(string)
			firestoreUpdates["class"] = class
		}
	}
	return firestoreUpdates, nil
}

func (am *AnnotationMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	beMapper := &BaseEntityMapper{}
	firestoreFilters, _ := beMapper.MapFilters(filters)

	for _, f := range filters {
		switch f.Field {
		case constants.AnnotationScoreField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "score",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.AnnotationClassField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "class",
				Operator: f.Operator,
				Value:    f.Value,
			})
		default:
			return nil, fmt.Errorf("unknown filter field: %s", f.Field)
		}
	}

	return firestoreFilters, nil
}
