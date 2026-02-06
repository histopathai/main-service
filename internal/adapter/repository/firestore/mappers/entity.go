package mappers

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type EntityMapper[T port.Entity] struct{}

func NewEntityMapper[T port.Entity]() *EntityMapper[T] {
	return &EntityMapper[T]{}
}

func (em *EntityMapper[T]) ToFirestoreMap(entity T) map[string]interface{} {
	m := map[string]interface{}{
		fields.EntityEntityType.FirestoreName(): entity.GetEntityType().String(),
		fields.EntityCreatorID.FirestoreName():  entity.GetCreatorID(),
		fields.EntityName.FirestoreName():       entity.GetName(),
		fields.EntityCreatedAt.FirestoreName():  entity.GetCreatedAt(),
		fields.EntityUpdatedAt.FirestoreName():  entity.GetUpdatedAt(),
		fields.EntityIsDeleted.FirestoreName():  entity.IsDeleted(),
	}

	parent := entity.GetParent()
	if parent != nil && entity.HasParent() {
		m[fields.EntityParentType.FirestoreName()] = parent.Type.String()
		m[fields.EntityParentID.FirestoreName()] = parent.ID
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

	entityTypeStr, ok := data[fields.EntityEntityType.FirestoreName()].(string)
	if !ok {
		return nil, errors.NewValidationError("missing entity_type field", nil)
	}
	entityType, err := vobj.NewEntityTypeFromString(entityTypeStr)
	if err != nil {
		return nil, err
	}

	creatorID, ok := data[fields.EntityCreatorID.FirestoreName()].(string)
	if !ok || creatorID == "" {
		return nil, errors.NewValidationError("missing creator_id field", nil)
	}

	name, ok := data[fields.EntityName.FirestoreName()].(string)
	if !ok {
		return nil, errors.NewValidationError("missing name field", nil)
	}

	var parent *vobj.ParentRef
	parentTypeStr, hasParentType := data[fields.EntityParentType.FirestoreName()].(string)
	parentID, hasParentID := data[fields.EntityParentID.FirestoreName()].(string)

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

	createdAt, ok := data[fields.EntityCreatedAt.FirestoreName()].(time.Time)
	if !ok {
		return nil, errors.NewValidationError("missing created_at field", nil)
	}

	updatedAt, ok := data[fields.EntityUpdatedAt.FirestoreName()].(time.Time)
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

	if deleted, ok := data[fields.EntityIsDeleted.FirestoreName()].(bool); ok {
		entity.SetDeleted(deleted)
	}

	return entity, nil
}

func (em *EntityMapper[T]) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mappedUpdates := make(map[string]interface{})

	for k, v := range updates {

		switch k {
		case fields.EntityName.DomainName():
			if name, ok := v.(string); ok {
				mappedUpdates[fields.EntityName.FirestoreName()] = name
			} else {
				return nil, errors.NewValidationError("invalid type for name field", nil)
			}

		case fields.EntityCreatorID.DomainName():
			if creatorID, ok := v.(string); ok {
				mappedUpdates[fields.EntityCreatorID.FirestoreName()] = creatorID
			} else {
				return nil, errors.NewValidationError("invalid type for creator_id field", nil)
			}

		case fields.EntityParentID.DomainName():
			// This case handles parent updates which might include both ID and Type in logic, but here k is just ParentID
			// FIXME: In current logic, usually update contains distinct fields.
			// The original code was: case constants.ParentIDField which is "parent_id"
			// But wait, the original code used constants.ParentIDField which maps to "parent_id" string in previous constants?
			// Let's assume input keys are Domain Names (UpperCamel).
			// If input is "ParentID" (DomainName), we map it to "parent_id" (FirestoreName).
			// But wait, the original switch block had:
			// case constants.ParentIDField:
			//     if parent, ok := v.(*vobj.ParentRef); ok { ... }
			// This means the input key was expected to be "parent_id" (or whatever constant was).
			// And the value was expected to be *vobj.ParentRef.
			// If we are standardizing on DomainName input keys:
			// The input key should be fields.EntityParentID.DomainName() ("ParentID").
			// The logic handles *vobj.ParentRef which updates two firestore fields: parent_type and parent_id.

			if parent, ok := v.(*vobj.ParentRef); ok {
				mappedUpdates[fields.EntityParentType.FirestoreName()] = parent.Type
				mappedUpdates[fields.EntityParentID.FirestoreName()] = parent.ID
			} else {
				return nil, errors.NewValidationError("invalid type for parent field", nil)
			}

		case fields.EntityEntityType.DomainName():
			if entityType, ok := v.(vobj.EntityType); ok {
				mappedUpdates[fields.EntityEntityType.FirestoreName()] = entityType.String()
			} else {
				return nil, errors.NewValidationError("invalid type for entity_type field", nil)
			}

		case fields.EntityIsDeleted.DomainName():
			if deleted, ok := v.(bool); ok {
				mappedUpdates[fields.EntityIsDeleted.FirestoreName()] = deleted
			} else {
				return nil, errors.NewValidationError("invalid type for deleted field", nil)
			}

		case fields.EntityCreatedAt.DomainName():
			// created_at should not be updated
			continue
		case fields.EntityUpdatedAt.DomainName():
			// updated_at will be set automatically
		default:
			// Unknown field, will be handled concrete mappers
			continue
		}
	}
	// Always set updated_at if not already set
	mappedUpdates[fields.EntityUpdatedAt.FirestoreName()] = time.Now()

	return mappedUpdates, nil
}

func (em *EntityMapper[T]) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))

	for _, filter := range filters {
		// Use EntityFields logic
		if fields.EntityField(filter.Field).IsValid() {
			firestoreField := fields.MapToFirestore(filter.Field)
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    firestoreField,
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		} else {
			// Check if it's a domain name
			found := false
			for _, ef := range fields.EntityFields {
				if ef.DomainName() == filter.Field {
					mappedFilters = append(mappedFilters, query.Filter{
						Field:    ef.FirestoreName(),
						Operator: filter.Operator,
						Value:    filter.Value,
					})
					found = true
					break
				}
			}
			if found {
				continue
			}

			// Let concrete mappers handle other fields
			continue
		}
	}
	return mappedFilters, nil
}
