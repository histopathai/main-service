package firestoreMappers

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type BaseEntityMapper struct{}

func (bem *BaseEntityMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.BaseEntity, error) {
	be := &model.BaseEntity{}

	data := doc.Data()

	be.SetID(doc.Ref.ID)
	be.SetCreatorID(data["creator_id"].(string))
	be.SetDeleted(data["deleted"].(bool))
	be.SetCreatedAt(data["created_at"].(time.Time))
	be.SetUpdatedAt(data["updated_at"].(time.Time))

	if data["name"] != nil {
		be.SetName(data["name"].(string))
	}
	return be, nil
}

func (bem *BaseEntityMapper) ToFirestoreMap(be *model.BaseEntity) map[string]interface{} {
	m := map[string]interface{}{
		"creator_id": be.GetCreatorID(),
		"deleted":    be.IsDeleted(),
		"created_at": be.GetCreatedAt(),
		"updated_at": be.GetUpdatedAt(),
	}

	if be.GetName() != "" {
		m["name"] = be.GetName()
	}

	return m
}

func (bem *BaseEntityMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {

	if len(updates) == 0 {
		return nil, nil
	}
	firestoreUpdates := make(map[string]interface{})

	for k, v := range updates {
		switch k {
		case constants.NameField:
			firestoreUpdates["name"] = v
		case constants.CreatorIDField:
			firestoreUpdates["creator_id"] = v
		case constants.DeletedField:
			firestoreUpdates["deleted"] = v
		case constants.UpdatedAtField:
			firestoreUpdates["updated_at"] = v

		default:
			// Ignore unknown fields
			continue
		}

	}

	return firestoreUpdates, nil
}

func (bem *BaseEntityMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	mappedFilters := make([]query.Filter, 0, len(filters))

	for _, filter := range filters {
		switch filter.Field {
		case constants.NameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "name",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.DeletedField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "deleted",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.CreatorIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "creator_id",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.CreatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "created_at",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.UpdatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "updated_at",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		default:
			// Ignore unknown fields
			continue

		}

	}

	return mappedFilters, nil
}
