package command

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

// ============================================================================
// Create Command Interfaces
// ============================================================================
type CreateCommand interface {
	Validate() (map[string]interface{}, bool)
	ToEntity() (any, error)
}

// =============================================================================
// Create Entity Command
// =============================================================================

type CreateEntityCommand struct {
	Name       string
	EntityType string
	CreatorID  string
	ParentID   string
	ParentType string
}

func (c *CreateEntityCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if c.Name == "" {
		details["name"] = "Name is required"
	}
	if c.CreatorID == "" {
		details["creator_id"] = "CreatorID is required"
	}
	if c.ParentID == "" {
		details["parent_id"] = "ParentID is required"
	}
	if c.ParentType == "" {
		details["parent_type"] = "ParentType is required"
	}

	if c.EntityType != "" {
		_, err := vobj.NewEntityTypeFromString(c.EntityType)
		if err != nil {
			details["entity_type"] = "Invalid EntityType"
		}
	}

	if c.ParentType != "" {
		_, err := vobj.NewParentTypeFromString(c.ParentType)
		if err != nil {
			details["parent_type"] = "Invalid ParentType"
		}
	}
	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *CreateEntityCommand) ToEntity() (*vobj.Entity, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	entity_type, _ := vobj.NewEntityTypeFromString(c.EntityType)

	parent_type, _ := vobj.NewParentTypeFromString(c.ParentType)
	parent, _ := vobj.NewParentRef(c.ParentID, parent_type)

	entity, err := vobj.NewEntity(
		entity_type,
		c.Name,
		c.CreatorID,
		parent,
	)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

// =============================================================================
// Create Workspace Command
// =============================================================================

type CreateWorkspaceCommand struct {
	CreateEntityCommand
	OrganType       string
	Organization    string
	Description     string
	License         string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

func (c *CreateWorkspaceCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if c.Name == "" {
		details["name"] = "Name is required"
	}
	if c.CreatorID == "" {
		details["creator_id"] = "CreatorID is required"
	}

	if c.OrganType == "" {
		details["organ_type"] = "OrganType is required"
	} else {
		_, err := vobj.NewOrganTypeFromString(c.OrganType)
		if err != nil {
			details["organ_type"] = "Invalid OrganType"
		}
	}

	if c.Organization == "" {
		details["organization"] = "Organization is required"
	}

	if c.Description == "" {
		details["description"] = "Description is required"
	}

	if c.License == "" {
		details["license"] = "License is required"
	}

	if c.ResourceURL != nil && *c.ResourceURL == "" {
		details["resource_url"] = "ResourceURL cannot be empty if provided"
	}

	if c.ReleaseYear != nil && *c.ReleaseYear < 0 {
		details["release_year"] = "ReleaseYear cannot be negative"
	}
	if len(c.AnnotationTypes) > 0 {
		//Check all annotation types are non-empty strings
		for i, at := range c.AnnotationTypes {
			if at == "" {
				details["annotation_types"] = "AnnotationTypes cannot contain empty strings"
				break
			}
			//Check for duplicates
			for j := i + 1; j < len(c.AnnotationTypes); j++ {
				if at == c.AnnotationTypes[j] {
					details["annotation_types"] = "AnnotationTypes cannot contain duplicate values"
					break
				}
			}
		}
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *CreateWorkspaceCommand) ToEntity() (*model.Workspace, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	organType, err := vobj.NewOrganTypeFromString(c.OrganType)
	if err != nil {
		return nil, err
	}

	workspaceEntity := model.Workspace{
		Entity: vobj.Entity{
			EntityType: vobj.EntityTypeWorkspace,
			Name:       c.Name,
			CreatorID:  c.CreatorID,
			Parent: vobj.ParentRef{
				ID:   "None",
				Type: vobj.ParentTypeNone,
			},
		},
		OrganType:    organType,
		Organization: c.Organization,
		Description:  c.Description,
		License:      c.License,
	}

	if c.ResourceURL != nil {
		workspaceEntity.ResourceURL = c.ResourceURL
	}

	if c.ReleaseYear != nil {
		workspaceEntity.ReleaseYear = c.ReleaseYear
	}

	if c.AnnotationTypes != nil {
		workspaceEntity.AnnotationTypes = c.AnnotationTypes
	}

	return &workspaceEntity, nil
}

// =============================================================================
// Create Patient Command
// =============================================================================

type CreatePatientCommand struct {
	CreateEntityCommand
	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *int
	History *string
}

func (c *CreatePatientCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.CreateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	if c.ParentType != vobj.ParentTypeWorkspace.String() {
		details["parent_type"] = "ParentType must be WORKSPACE"
	}

	if c.Age != nil && (*c.Age < 0 || *c.Age > 120) {
		details["age"] = "Age must be between 0 and 120"
	}

	if len(details) > 0 {
		return details, false
	}

	return nil, true
}

func (c *CreatePatientCommand) ToEntity() (*model.Patient, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	patientEntity := model.Patient{
		Entity:  *baseEntity,
		Age:     c.Age,
		Gender:  c.Gender,
		Race:    c.Race,
		Disease: c.Disease,
		Subtype: c.Subtype,
		Grade:   c.Grade,
		History: c.History,
	}

	return &patientEntity, nil
}

// =============================================================================
// Create Annotation Command
// =============================================================================

type CommandPoint struct {
	X float64
	Y float64
}

type CreateAnnotationCommand struct {
	CreateEntityCommand
	WsID     *string
	TagType  string
	Value    any
	Color    *string
	IsGlobal bool
	Points   []CommandPoint
}

func (c *CreateAnnotationCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.CreateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	if c.TagType == "" {
		details["tag_type"] = "TagType is required"
	} else {
		_, err := vobj.NewTagTypeFromString(c.TagType)
		if err != nil {
			details["tag_type"] = "Invalid TagType"
		}
	}

	if c.Value == nil {
		details["value"] = "Value is required"
	}

	if !c.IsGlobal && len(c.Points) < 3 {
		details["points"] = "At least 3 points are required for non-global annotations"
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *CreateAnnotationCommand) ToEntity() (*model.Annotation, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}
	tagType, err := vobj.NewTagTypeFromString(c.TagType)
	if err != nil {
		return nil, err
	}

	var points []vobj.Point
	if !c.IsGlobal {
		points = make([]vobj.Point, len(c.Points))
		for i, p := range c.Points {
			points[i] = vobj.Point{X: p.X, Y: p.Y}
		}
	} else {
		points = nil
	}
	wsID := ""
	if c.WsID != nil {
		wsID = *c.WsID
	}
	annotationEntity := model.Annotation{
		Entity:   *baseEntity,
		WsID:     wsID,
		TagType:  tagType,
		Value:    c.Value,
		Color:    c.Color,
		IsGlobal: c.IsGlobal,
		Polygon:  &points,
	}

	return &annotationEntity, nil
}

// =============================================================================
// Create AnnotationType Command
// =============================================================================

type CreateAnnotationTypeCommand struct {
	CreateEntityCommand
	TagType    string
	IsGlobal   bool
	IsRequired bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}

func (c *CreateAnnotationTypeCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if c.Name == "" {
		details["name"] = "Name is required"
	}
	if c.CreatorID == "" {
		details["creator_id"] = "CreatorID is required"
	}

	if c.TagType == "" {
		details["tag_type"] = "TagType is required"
	} else {
		tagType, err := vobj.NewTagTypeFromString(c.TagType)
		if err != nil {
			details["tag_type"] = "Invalid TagType"
		} else {
			switch tagType {
			case vobj.NumberTag:
				if c.Min == nil || c.Max == nil {
					details["min_max"] = "Min and Max are required for Number TagType"
				} else if *c.Min >= *c.Max {
					details["min_max"] = "Min must be less than Max for Number TagType"
				}
			case vobj.BooleanTag, vobj.TextTag:
				if len(c.Options) > 0 || c.Min != nil || c.Max != nil {
					details["options_min_max"] = "Options, Min, and Max are not applicable for Boolean or Text TagType"
				}
			case vobj.MultiSelectTag, vobj.SelectTag:
				if len(c.Options) < 1 {
					details["options"] = "At least one option must be provided for MultiSelect or SingleSelect TagType"
				}
			}
		}
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *CreateAnnotationTypeCommand) ToEntity() (*model.AnnotationType, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	tagType, err := vobj.NewTagTypeFromString(c.TagType)
	if err != nil {
		return nil, err
	}

	annotationTypeEntity := model.AnnotationType{
		Entity: vobj.Entity{
			EntityType: vobj.EntityTypeAnnotationType,
			Name:       c.Name,
			CreatorID:  c.CreatorID,
			Parent: vobj.ParentRef{
				ID:   "None",
				Type: vobj.ParentTypeNone,
			},
		},
		TagType:    tagType,
		IsGlobal:   c.IsGlobal,
		IsRequired: c.IsRequired,
		Options:    c.Options,
		Min:        c.Min,
		Max:        c.Max,
		Color:      c.Color,
	}

	return &annotationTypeEntity, nil
}
