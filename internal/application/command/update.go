package command

import (
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
)

// ============================================================================
// Update Command Interfaces
// ============================================================================
type UpdateCommand interface {
	Validate() (map[string]interface{}, bool)
	GetID() string
	GetUpdates() map[string]interface{}
}

// ===============================================================================
// Update Entity Command
// ===============================================================================
type UpdateEntityCommand struct {
	ID        string
	CreatorID *string
	Name      *string
}

func (c *UpdateEntityCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})
	if c.ID == "" {
		details["id"] = "ID is required"
	}
	if c.Name != nil && *c.Name == "" {
		details["name"] = "Name cannot be empty"
	}
	if c.CreatorID != nil && *c.CreatorID == "" {
		details["creator_id"] = "CreatorID cannot be empty"
	}
	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *UpdateEntityCommand) GetID() string {
	return c.ID
}

func (c *UpdateEntityCommand) GetUpdates() map[string]interface{} {
	updates := make(map[string]interface{})

	if _, ok := c.Validate(); !ok {
		return nil
	}

	if c.CreatorID != nil {
		updates[fields.EntityCreatorID.DomainName()] = *c.CreatorID
	}
	if c.Name != nil {
		updates[fields.EntityName.DomainName()] = *c.Name
	}
	return updates
}

//===============================================================================
// Update Workspace Command
//===============================================================================

type UpdateWorkspaceCommand struct {
	UpdateEntityCommand

	OrganType       *string
	Organization    *string
	Description     *string
	License         *string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

func (c *UpdateWorkspaceCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.UpdateEntityCommand.Validate()
	if !ok {
		return details, false
	}

	// Workspace-specific validation
	if c.OrganType != nil {
		_, err := vobj.NewOrganTypeFromString(*c.OrganType)
		if err != nil {
			details["organ_type"] = "Invalid OrganType value"
		}
	}

	if len(c.AnnotationTypes) > 0 {
		for _, at := range c.AnnotationTypes {
			if at == "" {
				details["annotation_types"] = "Annotation type IDs cannot be empty"
				break
			}
		}
	}

	if len(details) > 0 {
		return details, false
	}

	return nil, true
}

func (c *UpdateWorkspaceCommand) GetUpdates() map[string]interface{} {
	if _, ok := c.Validate(); !ok {
		return nil
	}

	// Base updates
	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	// Workspace-specific updates
	if c.OrganType != nil {
		organType, _ := vobj.NewOrganTypeFromString(*c.OrganType)
		updates[fields.WorkspaceOrganType.DomainName()] = organType
	}
	if c.Organization != nil {
		updates[fields.WorkspaceOrganization.DomainName()] = *c.Organization
	}
	if c.Description != nil {
		updates[fields.WorkspaceDescription.DomainName()] = *c.Description
	}
	if c.License != nil {
		updates[fields.WorkspaceLicense.DomainName()] = *c.License
	}
	if c.ResourceURL != nil {
		updates[fields.WorkspaceResourceURL.DomainName()] = *c.ResourceURL
	}
	if c.ReleaseYear != nil {
		updates[fields.WorkspaceReleaseYear.DomainName()] = *c.ReleaseYear
	}
	if len(c.AnnotationTypes) != 0 {
		updates[fields.WorkspaceAnnotationTypes.DomainName()] = c.AnnotationTypes
	}

	return updates
}

//===============================================================================
// Update Patient Command
//===============================================================================

type UpdatePatientCommand struct {
	UpdateEntityCommand

	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *string
	History *string
}

func (c *UpdatePatientCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.UpdateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	// Patient-specific validations
	if c.Age != nil && (*c.Age < 0 || *c.Age > 120) {
		details["age"] = "Age must be between 0 and 120"
	}
	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *UpdatePatientCommand) GetUpdates() map[string]interface{} {
	if _, ok := c.Validate(); !ok {
		return nil
	}

	// Base updates
	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	// Patient-specific updates
	if c.Age != nil {
		updates[fields.PatientAge.DomainName()] = *c.Age
	}
	if c.Gender != nil {
		updates[fields.PatientGender.DomainName()] = *c.Gender
	}
	if c.Race != nil {
		updates[fields.PatientRace.DomainName()] = *c.Race
	}
	if c.Disease != nil {
		updates[fields.PatientDisease.DomainName()] = *c.Disease
	}
	if c.Subtype != nil {
		updates[fields.PatientSubtype.DomainName()] = *c.Subtype
	}
	if c.Grade != nil {
		updates[fields.PatientGrade.DomainName()] = *c.Grade
	}
	if c.History != nil {
		updates[fields.PatientHistory.DomainName()] = *c.History
	}

	return updates
}

//===============================================================================
// Update Annotation Command
//===============================================================================

type UpdateAnnotationCommand struct {
	UpdateEntityCommand
	Value    interface{}
	Color    *string
	IsGlobal *bool
	Points   []CommandPoint
}

func (c *UpdateAnnotationCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.UpdateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	if c.Points != nil && len(c.Points) < 3 && (c.IsGlobal == nil || !*c.IsGlobal) {
		details["points"] = "At least 3 points required for non-global annotation"
	}

	if c.Color != nil && *c.Color == "" {
		details["color"] = "Color cannot be empty"
	}
	if len(details) > 0 {
		return details, false
	}

	return nil, true
}

func (c *UpdateAnnotationCommand) GetUpdates() map[string]interface{} {
	if _, ok := c.Validate(); !ok {
		return nil
	}

	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	if c.Value != nil {
		updates[fields.AnnotationTagValue.DomainName()] = c.Value
	}
	if c.Color != nil {
		updates[fields.AnnotationColor.DomainName()] = *c.Color
	}
	if c.IsGlobal != nil {
		updates[fields.AnnotationIsGlobal.DomainName()] = *c.IsGlobal
	}
	if c.Points != nil {
		points := make([]vobj.Point, len(c.Points))
		for i, p := range c.Points {
			points[i] = vobj.Point{X: p.X, Y: p.Y}
		}
		updates[fields.AnnotationPolygon.DomainName()] = points
	}

	return updates
}

//===============================================================================
// Update Annotation Type Command
//===============================================================================

type UpdateAnnotationTypeCommand struct {
	UpdateEntityCommand
	IsGlobal   *bool
	IsRequired *bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}

func (c *UpdateAnnotationTypeCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.UpdateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	if c.Min != nil && c.Max != nil && *c.Min >= *c.Max {
		details["min_max"] = "Min must be less than Max"
	}

	if len(c.Options) > 0 {
		for _, option := range c.Options {
			if option == "" {
				details["options"] = "Options cannot contain empty strings"
				break
			}
		}
	}
	if c.Color != nil && *c.Color == "" {
		details["color"] = "Color cannot be empty"
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *UpdateAnnotationTypeCommand) GetUpdates() map[string]interface{} {
	if _, ok := c.Validate(); !ok {
		return nil
	}

	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	if c.IsGlobal != nil {
		updates[fields.AnnotationTypeIsGlobal.DomainName()] = *c.IsGlobal
	}
	if c.IsRequired != nil {
		updates[fields.AnnotationTypeIsRequired.DomainName()] = *c.IsRequired
	}
	if c.Options != nil {
		updates[fields.AnnotationTypeOptions.DomainName()] = c.Options
	}
	if c.Min != nil {
		updates[fields.AnnotationTypeMin.DomainName()] = *c.Min
	}
	if c.Max != nil {
		updates[fields.AnnotationTypeMax.DomainName()] = *c.Max
	}
	if c.Color != nil {
		updates[fields.AnnotationTypeColor.DomainName()] = *c.Color
	}

	return updates
}

// ================================================================================
// Update Image Command
// ================================================================================

type UpdateImageCommand struct {
	UpdateEntityCommand

	Width         *int
	Height        *int
	Size          *int64
	Magnification *vobj.OpticalMagnification
}

func (c *UpdateImageCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.UpdateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	if c.Width != nil && *c.Width <= 0 {
		details["width"] = "Width must be positive"
	}
	if c.Height != nil && *c.Height <= 0 {
		details["height"] = "Height must be positive"
	}
	if c.Size != nil && *c.Size <= 0 {
		details["size"] = "Size must be positive"
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *UpdateImageCommand) GetUpdates() map[string]interface{} {
	if _, ok := c.Validate(); !ok {
		return nil
	}

	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	if c.Width != nil {
		updates[fields.ImageWidth.DomainName()] = *c.Width
	}
	if c.Height != nil {
		updates[fields.ImageHeight.DomainName()] = *c.Height
	}
	if c.Size != nil {
		updates[fields.ImageSize.DomainName()] = *c.Size
	}
	if c.Magnification != nil {
		updates[fields.ImageMagnification.DomainName()] = c.Magnification.GetMap()
	}

	return updates
}

// ================================================================================
// Update Content Command
// ================================================================================

type UpdateContentCommand struct {
	UpdateEntityCommand

	Provider *string
	Path     *string
}

func (c *UpdateContentCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.UpdateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	if c.Provider != nil {
		_, err := vobj.NewContentProviderFromString(*c.Provider)
		if err != nil {
			details["provider"] = "Invalid ContentProvider value"
		}
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *UpdateContentCommand) GetUpdates() map[string]interface{} {
	if _, ok := c.Validate(); !ok {
		return nil
	}

	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	if c.Provider != nil {
		provider, _ := vobj.NewContentProviderFromString(*c.Provider)
		updates[fields.ContentProvider.DomainName()] = provider
	}
	if c.Path != nil {
		updates[fields.ContentPath.DomainName()] = *c.Path
	}

	return updates
}
