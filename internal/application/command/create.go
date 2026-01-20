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
	details, ok := c.CreateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
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

	entityInterface, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	organType, err := vobj.NewOrganTypeFromString(c.OrganType)
	if err != nil {
		return nil, err
	}

	workspaceEntity := model.Workspace{
		Entity:       *entityInterface,
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
	details, ok := c.CreateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
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

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}
	tagType, err := vobj.NewTagTypeFromString(c.TagType)
	if err != nil {
		return nil, err
	}

	annotationTypeEntity := model.AnnotationType{
		Entity:     *baseEntity,
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

// =============================================================================
// Create Image Command
// =============================================================================
type CreateImageCommand struct {
	CreateEntityCommand

	// Required fields
	WsID   string
	Format string

	// Origin content (required)
	OriginContent struct {
		Provider    string
		Path        string
		ContentType string
		Size        int64
		Metadata    map[string]string
	}

	// Optional basic fields
	Width  *int
	Height *int

	// Optional WSI fields
	Magnification *struct {
		Objective         *float64
		NativeLevel       *int
		ScanMagnification *float64
	}

	// Optional processed content (usually not provided on create)
	ProcessedContent *struct {
		DZI       *ContentData
		Tiles     *ContentData
		Thumbnail *ContentData
		IndexMap  *ContentData
	}

	// Optional processing info (usually not provided on web uploads)
	Processing *struct {
		Status          *string
		Version         *string
		FailureReason   *string
		RetryCount      *int
		LastProcessedAt *string // ISO 8601 format
	}
}

type ContentData struct {
	Provider    string
	Path        string
	ContentType string
	Size        int64
	Metadata    map[string]string
}

func (c *CreateImageCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.CreateEntityCommand.Validate()
	if !ok {
		// Base entity validation failed, use those details
		return details, false
	}

	// Initialize details for this level
	details = make(map[string]interface{})

	// Required fields
	if c.WsID == "" {
		details["ws_id"] = "WsID is required"
	}
	if c.Format == "" {
		details["format"] = "Format is required"
	}

	// Origin content validation
	if c.OriginContent.Provider == "" {
		details["origin_content.provider"] = "Origin content provider is required"
	} else {
		provider := vobj.ContentProvider(c.OriginContent.Provider)
		if !provider.IsValid() {
			details["origin_content.provider"] = "Invalid origin content provider"
		}
	}

	if c.OriginContent.Path == "" {
		details["origin_content.path"] = "Origin content path is required"
	}

	if c.OriginContent.ContentType == "" {
		details["origin_content.content_type"] = "Origin content type is required"
	} else {
		contentType := vobj.ContentType(c.OriginContent.ContentType)
		if !contentType.IsValid() {
			details["origin_content.content_type"] = "Invalid origin content type"
		} else if contentType.GetCategory() != "image" {
			// Ensure origin is an image type
			details["origin_content.content_type"] = "Origin content must be an image type"
		}
	}

	if c.OriginContent.Size <= 0 {
		details["origin_content.size"] = "Origin content size must be positive"
	}

	// Optional fields validation
	if c.Width != nil && *c.Width <= 0 {
		details["width"] = "Width must be positive if provided"
	}

	if c.Height != nil && *c.Height <= 0 {
		details["height"] = "Height must be positive if provided"
	}

	// Processing validation (if provided)
	if c.Processing != nil {
		if c.Processing.Status != nil {
			if _, err := vobj.NewImageStatusFromString(*c.Processing.Status); err != nil {
				details["processing.status"] = "Invalid processing status"
			}
		}

		if c.Processing.Version != nil {
			version := vobj.ProcessingVersion(*c.Processing.Version)
			if !version.IsValid() {
				details["processing.version"] = "Invalid processing version"
			}
		}

		if c.Processing.RetryCount != nil && *c.Processing.RetryCount < 0 {
			details["processing.retry_count"] = "Retry count cannot be negative"
		}
	}

	// Processed content validation (if provided)
	if c.ProcessedContent != nil {
		if c.ProcessedContent.DZI != nil {
			if contentDetails, valid := c.validateContentData(c.ProcessedContent.DZI); !valid {
				details["processed_content.dzi"] = contentDetails
			}
		}
		if c.ProcessedContent.Tiles != nil {
			if contentDetails, valid := c.validateContentData(c.ProcessedContent.Tiles); !valid {
				details["processed_content.tiles"] = contentDetails
			}
		}
		if c.ProcessedContent.Thumbnail != nil {
			if contentDetails, valid := c.validateContentData(c.ProcessedContent.Thumbnail); !valid {
				details["processed_content.thumbnail"] = contentDetails
			}
		}
		if c.ProcessedContent.IndexMap != nil {
			if contentDetails, valid := c.validateContentData(c.ProcessedContent.IndexMap); !valid {
				details["processed_content.index_map"] = contentDetails
			}
		}
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *CreateImageCommand) validateContentData(data *ContentData) (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if data == nil {
		return nil, true
	}

	if data.Provider == "" {
		details["provider"] = "Provider is required"
	} else {
		provider := vobj.ContentProvider(data.Provider)
		if !provider.IsValid() {
			details["provider"] = "Invalid provider"
		}
	}

	if data.Path == "" {
		details["path"] = "Path is required"
	}

	if data.ContentType == "" {
		details["content_type"] = "Content type is required"
	} else {
		contentType := vobj.ContentType(data.ContentType)
		if !contentType.IsValid() {
			details["content_type"] = "Invalid content type"
		}
	}

	if data.Size <= 0 {
		details["size"] = "Size must be positive"
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *CreateImageCommand) ToEntity() (*model.Image, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	// Build origin content
	originContent := &vobj.Content{
		Provider:    vobj.ContentProvider(c.OriginContent.Provider),
		Path:        c.OriginContent.Path,
		ContentType: vobj.ContentType(c.OriginContent.ContentType),
		Size:        c.OriginContent.Size,
		Metadata:    c.OriginContent.Metadata,
	}

	// Build image entity
	image := &model.Image{
		Entity:        *baseEntity,
		WsID:          c.WsID,
		Format:        c.Format,
		Width:         c.Width,
		Height:        c.Height,
		OriginContent: originContent,
	}

	// Size is from origin content
	image.Size = &c.OriginContent.Size

	// Optional magnification
	if c.Magnification != nil {
		image.Magnification = &vobj.OpticalMagnification{
			Objective:         c.Magnification.Objective,
			NativeLevel:       c.Magnification.NativeLevel,
			ScanMagnification: c.Magnification.ScanMagnification,
		}
	}

	// Optional processed content
	if c.ProcessedContent != nil {
		procContent := &model.ProcessedContent{}

		if c.ProcessedContent.DZI != nil {
			procContent.DZI = c.contentDataToVobj(c.ProcessedContent.DZI)
		}
		if c.ProcessedContent.Tiles != nil {
			procContent.Tiles = c.contentDataToVobj(c.ProcessedContent.Tiles)
		}
		if c.ProcessedContent.Thumbnail != nil {
			procContent.Thumbnail = c.contentDataToVobj(c.ProcessedContent.Thumbnail)
		}
		if c.ProcessedContent.IndexMap != nil {
			procContent.IndexMap = c.contentDataToVobj(c.ProcessedContent.IndexMap)
		}

		image.ProcessedContent = procContent
	}

	// Processing info - defaults for web uploads
	if c.Processing != nil {
		if c.Processing.Status != nil {
			image.Processing.Status, err = vobj.NewImageStatusFromString(*c.Processing.Status)
			if err != nil {
				return nil, err
			}
		} else {
			image.Processing.Status = vobj.StatusPending // Default for web uploads
		}

		if c.Processing.Version != nil {
			image.Processing.Version = vobj.ProcessingVersion(*c.Processing.Version)
		}

		if c.Processing.FailureReason != nil {
			image.Processing.FailureReason = c.Processing.FailureReason
		}

		if c.Processing.RetryCount != nil {
			image.Processing.RetryCount = *c.Processing.RetryCount
		}

		if c.Processing.LastProcessedAt != nil {
			// Parse ISO 8601 timestamp if needed
			// For now just skip it, will be set when processing starts
		}
	} else {
		// Default processing info for web uploads
		image.Processing.Status = vobj.StatusPending
		image.Processing.RetryCount = 0
	}

	return image, nil
}

func (c *CreateImageCommand) contentDataToVobj(data *ContentData) *vobj.Content {
	if data == nil {
		return nil
	}

	return &vobj.Content{
		Provider:    vobj.ContentProvider(data.Provider),
		Path:        data.Path,
		ContentType: vobj.ContentType(data.ContentType),
		Size:        data.Size,
		Metadata:    data.Metadata,
	}
}
