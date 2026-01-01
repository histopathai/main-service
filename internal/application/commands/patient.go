package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
)

type CreatePatientCommand struct {
	//Base Entity fields
	Name      string
	CreatorID string
	ParentID  string
	//Patient specific fields
	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *int
	History *string
}

func (c *CreatePatientCommand) ToEntity() model.Patient {
	return model.Patient{
		BaseEntity: model.BaseEntity{
			EntityType: constants.EntityTypePatient,
			Name:       &c.Name,
			CreatorID:  c.CreatorID,
			Parent:     &model.ParentRef{ID: c.ParentID, Type: constants.ParentTypeWorkspace},
		},
		Age:     c.Age,
		Gender:  c.Gender,
		Race:    c.Race,
		Disease: c.Disease,
		Subtype: c.Subtype,
		Grade:   c.Grade,
		History: c.History,
	}

}

type UpdatePatientCommand struct {
	// Base Entity fields
	ID        string
	Name      *string
	CreatorID *string
	ParentID  *string
	// Patient specific fields

	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *int
	History *string
}

func (c *UpdatePatientCommand) GetID() string {
	return c.ID
}

func (c *UpdatePatientCommand) ApplyTo(entity model.Patient) model.Patient {
	if c.Name != nil {
		entity.Name = c.Name
	}
	if c.CreatorID != nil {
		entity.CreatorID = *c.CreatorID
	}
	if c.ParentID != nil {
		entity.Parent = &model.ParentRef{ID: *c.ParentID, Type: constants.ParentTypeWorkspace}
	}
	if c.Age != nil {
		entity.Age = c.Age
	}
	if c.Gender != nil {
		entity.Gender = c.Gender
	}
	if c.Race != nil {
		entity.Race = c.Race
	}
	if c.Disease != nil {
		entity.Disease = c.Disease
	}
	if c.Subtype != nil {
		entity.Subtype = c.Subtype
	}
	if c.Grade != nil {
		entity.Grade = c.Grade
	}
	if c.History != nil {
		entity.History = c.History
	}
	return entity
}
