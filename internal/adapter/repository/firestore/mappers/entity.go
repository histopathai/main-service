package mappers

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type EntityMapper[T port.Entity] struct{}

func NewEntityMapper[T port.Entity]() *EntityMapper[T] {
	return &EntityMapper[T]{}
}

func (em *EntityMapper[T]) ToFirestoreMap(entity T) map[string]interface{} {
	m := map[string]interface{}{
		"entity_type": entity.GetEntityType().String(),
		"creator_id":  entity.GetCreatorID(),
		"name":        entity.GetName(),
		"created_at":  entity.GetCreatedAt(),
		"updated_at":  entity.GetUpdatedAt(),
		"is_deleted":  entity.IsDeleted(),
	}

	parent := entity.GetParent()
	if parent != nil && entity.HasParent() {
		m["parent_type"] = parent.Type.String()
		m["parent_id"] = parent.ID
	}

	return m
}

func (em *EntityMapper[T]) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (T, error) {
	var zero T
	return zero, errors.NewInternalError("FromFirestoreDoc must be implemented by concrete mapper", nil)
}

func (em *EntityMapper[T]) ParseEntity(doc *firestore.DocumentSnapshot) (*vobj.Entity, error) {
	data := doc.Data()

	if data == nil {
		return nil, errors.NewInternalError("document data is nil", nil)
	}

	entityTypeStr, ok := data["entity_type"].(string)
	if !ok {
		return nil, errors.NewValidationError("missing entity_type field", nil)
	}
	entityType, err := vobj.NewEntityTypeFromString(entityTypeStr)
	if err != nil {
		return nil, err
	}

	creatorID, ok := data["creator_id"].(string)
	if !ok || creatorID == "" {
		return nil, errors.NewValidationError("missing creator_id field", nil)
	}

	name, ok := data["name"].(string)
	if !ok {
		return nil, errors.NewValidationError("missing name field", nil)
	}

	var parent *vobj.ParentRef
	parentTypeStr, hasParentType := data["parent_type"].(string)
	parentID, hasParentID := data["parent_id"].(string)

	if hasParentType && hasParentID && parentID != "" {
		parentType, err := vobj.NewParentTypeFromString(parentTypeStr)
		if err != nil {
			return nil, err
		}
		parent, err = vobj.NewParentRef(parentID, parentType)
		if err != nil {
			return nil, err
		}
	} else {
		parent, err = vobj.NewParentRef("", vobj.ParentTypeNone)
		if err != nil {
			return nil, err
		}
	}

	createdAt, ok := data["created_at"].(time.Time)
	if !ok {
		return nil, errors.NewValidationError("missing created_at field", nil)
	}

	updatedAt, ok := data["updated_at"].(time.Time)
	if !ok {
		return nil, errors.NewValidationError("missing updated_at field", nil)
	}

	entity, err := vobj.NewEntity(entityType, name, creatorID, parent)
	if err != nil {
		return nil, err
	}

	entity.SetID(doc.Ref.ID)
	entity.SetCreatedAt(createdAt)
	entity.SetUpdatedAt(updatedAt)

	if deleted, ok := data["is_deleted"].(bool); ok {
		entity.SetDeleted(deleted)
	}

	return entity, nil
}

func (em *EntityMapper[T]) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mappedUpdates := make(map[string]interface{})

	for k, v := range updates {
		switch k {
		case constants.NameField:
			if name, ok := v.(string); ok {
				mappedUpdates["name"] = name
			} else {
				return nil, errors.NewValidationError("invalid type for name field", nil)
			}

		case constants.CreatorIDField:
			if creatorID, ok := v.(string); ok {
				mappedUpdates["creator_id"] = creatorID
			} else {
				return nil, errors.NewValidationError("invalid type for creator_id field", nil)
			}

		case constants.ParentIDField:
			if parent, ok := v.(*vobj.ParentRef); ok {
				mappedUpdates["parent_type"] = parent.Type
				mappedUpdates["parent_id"] = parent.ID
			} else {
				return nil, errors.NewValidationError("invalid type for parent field", nil)
			}

		case constants.EntityTypeField:
			if entityType, ok := v.(vobj.EntityType); ok {
				mappedUpdates["entity_type"] = entityType
			} else {
				return nil, errors.NewValidationError("invalid type for entity_type field", nil)
			}

		case constants.DeletedField:
			if deleted, ok := v.(bool); ok {
				mappedUpdates["is_deleted"] = deleted
			} else {
				return nil, errors.NewValidationError("invalid type for deleted field", nil)
			}

		case constants.CreatedAtField:
			// created_at should not be updated
			continue
		case constants.UpdatedAtField:
			// updated_at will be set automatically
		default:
			// Unknown field, will be handled concrete mappers
			continue
		}
	}
	// Always set updated_at if not already set
	mappedUpdates["updated_at"] = time.Now()

	return mappedUpdates, nil
}

func (em *EntityMapper[T]) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))

	for _, filter := range filters {
		switch filter.Field {
		case constants.NameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "name",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.CreatorIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "creator_id",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.EntityTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "entity_type",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.DeletedField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_deleted",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.ParentIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "parent_id",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.ParentTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "parent_type",
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
			continue
		}
	}
	return mappedFilters, nil
}
