package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreatePatientCommand struct {
	Name      string
	CreatorID string
	ParentID  string
	Age       *int
	Gender    *string
	Race      *string
	Disease   *string
	Subtype   *string
	Grade     *int
	History   *string
}

func NewCreatePatientCommand(
	name string,
	creatorID string,
	parentID string,
	age *int,
	gender *string,
	race *string,
	disease *string,
	subtype *string,
	grade *int,
	history *string,
) (*CreatePatientCommand, error) {
	details := make(map[string]any)
	if name == "" {
		details["name required"] = name
	}

	if creatorID == "" {
		details["creator_id required"] = creatorID
	}

	if parentID == "" {
		details["parent_id required"] = parentID
	}

	if age != nil && *age < 0 {
		details["age cannot be negative"] = *age
	}

	if len(details) > 0 {
		return nil, errors.NewValidationError("invalid create patient command", details)
	}
	return &CreatePatientCommand{
		Name:      name,
		CreatorID: creatorID,
		ParentID:  parentID,
		Age:       age,
		Gender:    gender,
		Race:      race,
		Disease:   disease,
		Subtype:   subtype,
		Grade:     grade,
		History:   history,
	}, nil
}

func (c *CreatePatientCommand) ToEntity() (*model.Patient, error) {
	parentRef, err := vobj.NewParentRef(c.ParentID, vobj.ParentTypeWorkspace)
	if err != nil {
		return nil, err
	}

	entity, err := vobj.NewEntity(
		vobj.EntityTypePatient,
		&c.Name,
		c.CreatorID,
		parentRef,
	)
	if err != nil {
		return nil, err
	}

	return &model.Patient{
		Entity:  entity,
		Age:     c.Age,
		Gender:  c.Gender,
		Race:    c.Race,
		Disease: c.Disease,
		Subtype: c.Subtype,
		Grade:   c.Grade,
		History: c.History,
	}, nil
}

type UpdatePatientCommand struct {
	ID        string
	Name      *string
	CreatorID *string
	ParentID  *string
	Age       *int
	Gender    *string
	Race      *string
	Disease   *string
	Subtype   *string
	Grade     *int
	History   *string
}

func (c *UpdatePatientCommand) GetID() string {
	return c.ID
}

func (c *UpdatePatientCommand) ApplyTo(entity *model.Patient) (*model.Patient, error) {
	if c.Name != nil {
		entity.SetName(*c.Name)
	}
	if c.CreatorID != nil {
		entity.SetCreatorID(*c.CreatorID)
	}
	if c.ParentID != nil {
		parentRef, err := vobj.NewParentRef(*c.ParentID, vobj.ParentTypeWorkspace)
		if err != nil {
			return nil, err
		}
		entity.SetParent(parentRef)
	}

	if c.Age != nil {
		if *c.Age < 0 {
			details := map[string]any{"age": *c.Age}
			return nil, errors.NewValidationError("age cannot be negative", details)
		}
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

	return entity, nil
}

func (c *UpdatePatientCommand) GetUpdates() (map[string]any, error) {
	updates := make(map[string]any)

	if c.Name != nil {
		updates[constants.NameField] = *c.Name
	}
	if c.CreatorID != nil {
		updates[constants.CreatorIDField] = *c.CreatorID
	}
	if c.ParentID != nil {
		updates[constants.ParentIDField] = *c.ParentID
		updates[constants.ParentTypeField] = vobj.ParentTypeWorkspace
	}
	if c.Age != nil {
		if *c.Age < 0 {
			details := map[string]any{"age cannot be negative": *c.Age}
			return nil, errors.NewValidationError("age cannot be negative", details)
		}
		updates[constants.AgeField] = *c.Age
	}
	if c.Gender != nil {
		updates[constants.GenderField] = *c.Gender
	}
	if c.Race != nil {
		updates[constants.RaceField] = *c.Race
	}
	if c.Disease != nil {
		updates[constants.DiseaseField] = *c.Disease
	}
	if c.Subtype != nil {
		updates[constants.SubtypeField] = *c.Subtype
	}
	if c.Grade != nil {
		updates[constants.GradeField] = *c.Grade
	}
	if c.History != nil {
		updates[constants.HistoryField] = *c.History
	}

	return updates, nil
}
