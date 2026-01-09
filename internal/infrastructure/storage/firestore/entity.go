package firestore

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

func EntityFromFirestore(doc *firestore.DocumentSnapshot) (*vobj.Entity, error) {
	data := doc.Data()

	entityType, ok := data["entity_type"].(vobj.EntityType)
	if !ok || !entityType.IsValid() {
		details := map[string]any{"entity_type": data["entity_type"]}
		return nil, errors.NewValidationError("invalid entity type", details)
	}

	namePtr, _ := data["name"].(string)
	var name *string
	if namePtr != "" {
		name = &namePtr
	}

	creatorID, ok := data["creator_id"].(string)
	if !ok || creatorID == "" {
		return nil, errors.NewValidationError("creator ID is required", nil)
	}

	var parent *vobj.ParentRef
	if parentType, ok := data["parent_type"].(vobj.ParentType); ok {
		if parentID, ok := data["parent_id"].(string); ok && parentID != "" {
			var err error
			parent, err = vobj.NewParentRef(parentID, parentType)
			if err != nil {
				return nil, err
			}
		}
	}

	createdAt, ok := data["created_at"].(time.Time)
	if !ok {
		return nil, errors.NewValidationError("created_at is required", nil)
	}

	updatedAt, ok := data["updated_at"].(time.Time)
	if !ok {
		return nil, errors.NewValidationError("updated_at is required", nil)
	}

	return &vobj.Entity{
		ID:         doc.Ref.ID,
		EntityType: entityType,
		Name:       name,
		CreatorID:  creatorID,
		Parent:     parent,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}, nil
}

func EntityToFirestoreMap(entity *vobj.Entity) map[string]interface{} {
	m := map[string]interface{}{
		"entity_type": entity.EntityType,
		"creator_id":  entity.CreatorID,
		"created_at":  entity.CreatedAt,
		"updated_at":  entity.UpdatedAt,
		"id":          entity.ID,
	}

	if entity.Name != nil {
		m["name"] = *entity.Name
	} else {
		m["name"] = ""
	}

	if entity.Parent != nil {
		m["parent_id"] = entity.Parent.ID
		m["parent_type"] = entity.Parent.Type
	} else {
		m["parent_id"] = ""
		m["parent_type"] = vobj.ParentTypeNone
	}

	return m
}

func EntityMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.NameField:
			firestoreUpdates["name"] = value
			delete(updates, key) // Processed field removed from original map
		case constants.CreatorIDField:
			firestoreUpdates["creator_id"] = value
			delete(updates, key)
		case constants.ParentIDField:
			firestoreUpdates["parent_id"] = value
			delete(updates, key)
		case constants.ParentTypeField:
			firestoreUpdates["parent_type"] = value
			delete(updates, key)
		case constants.UpdatedAtField:
			firestoreUpdates["updated_at"] = value
			delete(updates, key)
		case constants.CreatedAtField:
			firestoreUpdates["created_at"] = value
			delete(updates, key)
		case constants.EntityTypeField:
			firestoreUpdates["entity_type"] = value
			delete(updates, key)
		default:
			continue
			// Ignore unknown fields
			// The upper layer will handle or throw an error for these
		}
	}
	return firestoreUpdates, nil
}

func EntityMapFilter(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))
	processedIndices := make(map[int]bool)

	for i, f := range filters {
		switch f.Field {
		case constants.NameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    constants.NameField,
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.CreatorIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "creator_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.ParentIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "parent_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.ParentTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "parent_type",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.EntityTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "entity_type",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.CreatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "created_at",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.UpdatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "updated_at",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		default:
			continue
			// İgnore unknown fields
		}
	}

	// Remove processed filters from the original slice
	for i := len(filters) - 1; i >= 0; i-- {
		if processedIndices[i] {
			filters = append(filters[:i], filters[i+1:]...)
		}
	}

	return mappedFilters, nil
}
