package model

import "time"

type ParentType string

const (
	ParentTypeNone           ParentType = ""
	ParentTypeWorkspace      ParentType = "workspace"
	ParentTypePatient        ParentType = "patient"
	ParentTypeImage          ParentType = "image"
	ParentTypeAnnotationType ParentType = "annotation_type"
)

type ParentRef struct {
	ID   string
	Type ParentType
}

type BaseEntity struct {
	ID        string
	Parent    *ParentRef
	Name      *string
	CreatorID string
	CreatedAt time.Time
	UpdatedAt time.Time
	Deleted   bool
}

type Entity interface {
	GetID() string
	SetID(id string)
	SetCreatedAt(t time.Time)
	SetUpdatedAt(t time.Time)
	IsDeleted() bool
	SetDeleted(deleted bool)
	SetCreatorID(creatorID string)
	GetCreatorID() string
	GetName() string
	SetName(name string)

	GetParent() *ParentRef
	SetParent(parentID string, parentType ParentType)
	HasParent() bool
	GetParentID() string
	GetParentType() ParentType
}

func (e *BaseEntity) GetParent() *ParentRef {
	return e.Parent
}

func (e *BaseEntity) SetParent(parentID string, parentType ParentType) {
	if parentID == "" || parentType == ParentTypeNone {
		e.Parent = nil
		return
	}
	e.Parent = &ParentRef{
		ID:   parentID,
		Type: parentType,
	}
}

func (e *BaseEntity) HasParent() bool {
	return e.Parent != nil
}

func (e *BaseEntity) GetParentID() string {
	if e.Parent == nil {
		return ""
	}
	return e.Parent.ID
}

func (e *BaseEntity) GetParentType() ParentType {
	if e.Parent == nil {
		return ParentTypeNone
	}
	return e.Parent.Type
}

func (e *BaseEntity) GetID() string {
	return e.ID
}

func (e *BaseEntity) SetID(id string) {
	e.ID = id
}

func (e *BaseEntity) SetCreatedAt(t time.Time) {
	e.CreatedAt = t
}

func (e *BaseEntity) SetUpdatedAt(t time.Time) {
	e.UpdatedAt = t
}

func (e *BaseEntity) GetCreatedAt() time.Time {
	return e.CreatedAt
}

func (e *BaseEntity) GetUpdatedAt() time.Time {
	return e.UpdatedAt
}

func (e *BaseEntity) IsDeleted() bool {
	return e.Deleted
}

func (e *BaseEntity) SetDeleted(deleted bool) {
	e.Deleted = deleted
}

func (e *BaseEntity) SetCreatorID(creatorID string) {
	e.CreatorID = creatorID
}

func (e *BaseEntity) GetCreatorID() string {
	return e.CreatorID
}

func (e *BaseEntity) GetName() string {
	if e.Name == nil {
		return ""
	}
	return *e.Name
}

func (e *BaseEntity) SetName(name string) {
	e.Name = &name
}
