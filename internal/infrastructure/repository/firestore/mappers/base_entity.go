package firestoreMappers

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type EntityMapper struct{}

func (em *EntityMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*vobj.Entity, error) {
	data := doc.Data()

	entityTypeStr := data["entity_type"].(string)
	entityType, err := vobj.NewEntityTypeFromString(entityTypeStr)
	if err != nil {
		return nil, err
	}

	var name *string
	if data["name"] != nil {
		nameVal := data["name"].(string)
		name = &nameVal
	}

	var parent *vobj.ParentRef
	parentRefType := data["parent_type"]
	parentRefID := data["parent_id"]
	if parentRefType != nil && parentRefID != nil {
		parentType, err := vobj.NewEntityTypeFromString(parentRefType.(string))
		if err != nil {
			return nil, err
		}
		parent = &vobj.ParentRef{
			ID:   parentRefID.(string),
			Type: vobj.ParentType(parentType),
		}
	}

	entity := &vobj.Entity{
		ID:         doc.Ref.ID,
		EntityType: entityType,
		Name:       name,
		CreatorID:  data["creator_id"].(string),
		Parent:     parent,
		CreatedAt:  data["created_at"].(time.Time),
		UpdatedAt:  data["updated_at"].(time.Time),
		Deleted:    data["deleted"].(bool),
	}

	if hasChildren, ok := data["has_children"].(bool); ok {
		entity.HasChildren = hasChildren
	}
	if childCount, ok := data["child_count"].(int64); ok {
		entity.ChildCount = &childCount
	}

	return entity, nil
}

func (em *EntityMapper) ToFirestoreMap(entity *vobj.Entity) map[string]interface{} {
	m := map[string]interface{}{
		"entity_type": entity.EntityType.String(),
		"creator_id":  entity.CreatorID,
		"deleted":     entity.Deleted,
		"created_at":  entity.CreatedAt,
		"updated_at":  entity.UpdatedAt,
	}

	if entity.Name != nil {
		m["name"] = *entity.Name
	}

	if entity.Parent != nil && !entity.Parent.IsEmpty() {
		m["parent_id"] = entity.Parent.ID
		m["parent_type"] = entity.Parent.Type.String()
	}

	if entity.HasChildren {
		m["has_children"] = entity.HasChildren
	}
	if entity.ChildCount != nil {
		m["child_count"] = *entity.ChildCount
	}

	return m
}

func (em *EntityMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
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
		case constants.EntityTypeField:
			firestoreUpdates["entity_type"] = v
		case constants.ParentIDField:
			parentType, ok := updates[constants.ParentTypeField]
			if !ok {
				return nil, nil
			}
			if v == nil || v == "" || parentType == nil || parentType == "" {
				firestoreUpdates["parent_id"] = nil
				firestoreUpdates["parent_type"] = nil
			} else {
				firestoreUpdates["parent_id"] = v
				firestoreUpdates["parent_type"] = parentType
			}
		case constants.ParentTypeField:
			// ParentType Field processed in ParentID case
		case constants.HasChildrenField:
			firestoreUpdates["has_children"] = v
		case constants.ChildCountField:
			firestoreUpdates["child_count"] = v
		}
	}

	delete(updates, constants.NameField)
	delete(updates, constants.CreatorIDField)
	delete(updates, constants.DeletedField)
	delete(updates, constants.UpdatedAtField)
	delete(updates, constants.EntityTypeField)
	delete(updates, constants.ParentIDField)
	delete(updates, constants.ParentTypeField)
	delete(updates, constants.HasChildrenField)
	delete(updates, constants.ChildCountField)

	return firestoreUpdates, nil
}

func (em *EntityMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	mappedFilters := make([]query.Filter, 0)
	unprocessedIdx := 0

	for i, filter := range filters {
		processed := false

		switch filter.Field {
		case constants.NameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "name",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.DeletedField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "deleted",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.CreatorIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "creator_id",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.CreatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "created_at",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.UpdatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "updated_at",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.EntityTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "entity_type",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.ParentIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "parent_id",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.ParentTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "parent_type",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.HasChildrenField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "has_children",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		case constants.ChildCountField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "child_count",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		}

		if !processed {
			filters[unprocessedIdx] = filters[i]
			unprocessedIdx++
		}
	}

	for i := unprocessedIdx; i < len(filters); i++ {
		filters[i] = query.Filter{}
	}

	return mappedFilters, nil
}
