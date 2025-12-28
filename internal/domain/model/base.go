package model

import "time"

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
}

type BaseEntity struct {
	ID        string
	Name      *string
	CreatorID string
	CreatedAt time.Time
	UpdatedAt time.Time
	Deleted   bool
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
