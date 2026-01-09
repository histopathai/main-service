package vobj

import (
	"time"

	"github.com/histopathai/main-service/internal/shared/errors"
)

type EntityType string

const (
	EntityTypeImage          EntityType = "image"
	EntityTypeAnnotation     EntityType = "annotation"
	EntityTypePatient        EntityType = "patient"
	EntityTypeWorkspace      EntityType = "workspace"
	EntityTypeAnnotationType EntityType = "annotation_type"
)

func (e EntityType) IsValid() bool {
	switch e {
	case EntityTypeImage, EntityTypeAnnotation, EntityTypePatient, EntityTypeWorkspace, EntityTypeAnnotationType:
		return true
	default:
		return false
	}
}

func (e EntityType) String() string {
	return string(e)
}

func NewEntityTypeFromString(s string) (EntityType, error) {
	switch s {
	case "image":
		return EntityTypeImage, nil
	case "annotation":
		return EntityTypeAnnotation, nil
	case "patient":
		return EntityTypePatient, nil
	case "workspace":
		return EntityTypeWorkspace, nil
	case "annotation_type":
		return EntityTypeAnnotationType, nil
	default:
		details := map[string]any{"value": s}
		return "", errors.NewValidationError("invalid entity type", details)
	}
}

type Entity struct {
	ID         string
	EntityType EntityType
	Name       *string
	CreatorID  string
	Parent     *ParentRef
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func NewEntity(entityType EntityType, name *string, creatorID string, parent *ParentRef) (*Entity, error) {
	if !entityType.IsValid() {
		details := map[string]any{"entity_type": entityType}
		return nil, errors.NewValidationError("invalid entity type", details)
	}

	if creatorID == "" {
		return nil, errors.NewValidationError("creator ID is required", nil)
	}

	if name != nil && *name == "" {
		name = nil
	}

	if parent != nil && parent.IsEmpty() {
		parent = nil
	}

	return &Entity{
		EntityType: entityType,
		Name:       name,
		CreatorID:  creatorID,
		Parent:     parent,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// Getter metodları - value receiver (embedded struct'lar için çalışır)
func (e Entity) GetID() string {
	return e.ID
}

func (e Entity) GetCreatorID() string {
	return e.CreatorID
}

func (e Entity) GetName() string {
	if e.Name == nil {
		return ""
	}
	return *e.Name
}

func (e Entity) GetEntityType() EntityType {
	return e.EntityType
}

func (e Entity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e Entity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

func (e Entity) GetParent() *ParentRef {
	return e.Parent
}

func (e Entity) HasParent() bool {
	return e.Parent != nil && !e.Parent.IsEmpty()
}

// Setter metodları - pointer receiver (değişiklik yapabilmek için)
func (e *Entity) SetID(id string) {
	e.ID = id
}

func (e *Entity) SetCreatorID(creatorID string) {
	e.CreatorID = creatorID
}

func (e *Entity) SetName(name string) {
	e.Name = &name
}

func (e *Entity) SetEntityType(entityType EntityType) {
	e.EntityType = entityType
}

func (e *Entity) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

func (e *Entity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

func (e *Entity) SetParent(parent *ParentRef) {
	e.Parent = parent
}
