package command

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
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
		updates["creator_id"] = *c.CreatorID
	}
	if c.Name != nil {
		updates["name"] = *c.Name
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
		updates[constants.WorkspaceOrganTypeField] = organType
	}
	if c.Organization != nil {
		updates[constants.WorkspaceOrganizationField] = *c.Organization
	}
	if c.Description != nil {
		updates[constants.WorkspaceDescField] = *c.Description
	}
	if c.License != nil {
		updates[constants.WorkspaceLicenseField] = *c.License
	}
	if c.ResourceURL != nil {
		updates[constants.WorkspaceResourceURLField] = *c.ResourceURL
	}
	if c.ReleaseYear != nil {
		updates[constants.WorkspaceReleaseYearField] = *c.ReleaseYear
	}
	if len(c.AnnotationTypes) != 0 {
		updates[constants.WorkspaceAnnotationTypes] = c.AnnotationTypes
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
		updates[constants.PatientAgeField] = *c.Age
	}
	if c.Gender != nil {
		updates[constants.PatientGenderField] = *c.Gender
	}
	if c.Race != nil {
		updates[constants.PatientRaceField] = *c.Race
	}
	if c.Disease != nil {
		updates[constants.PatientDiseaseField] = *c.Disease
	}
	if c.Subtype != nil {
		updates[constants.PatientSubtypeField] = *c.Subtype
	}
	if c.Grade != nil {
		updates[constants.PatientGradeField] = *c.Grade
	}
	if c.History != nil {
		updates[constants.PatientHistoryField] = *c.History
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
		updates[constants.TagValueField] = c.Value
	}
	if c.Color != nil {
		updates[constants.TagColorField] = *c.Color
	}
	if c.IsGlobal != nil {
		updates[constants.TagGlobalField] = *c.IsGlobal
	}
	if c.Points != nil {
		points := make([]vobj.Point, len(c.Points))
		for i, p := range c.Points {
			points[i] = vobj.Point{X: p.X, Y: p.Y}
		}
		updates[constants.PolygonField] = points
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
		updates[constants.TagGlobalField] = *c.IsGlobal
	}
	if c.IsRequired != nil {
		updates[constants.TagRequiredField] = *c.IsRequired
	}
	if c.Options != nil {
		updates[constants.TagOptionsField] = c.Options
	}
	if c.Min != nil {
		updates[constants.TagMinField] = *c.Min
	}
	if c.Max != nil {
		updates[constants.TagMaxField] = *c.Max
	}
	if c.Color != nil {
		updates[constants.TagColorsField] = *c.Color
	}

	return updates
}

//===============================================================================
// Update Image Command
//===============================================================================

type UpdateImageCommand struct {
	UpdateEntityCommand

	// Basic image fields
	Width  *int
	Height *int
	Size   *int64
	Format *string

	// Magnification
	Magnification *struct {
		Objective         *float64
		NativeLevel       *int
		ScanMagnification *float64
	}

	// Origin content (usually not updated, but possible)
	OriginContent *ContentData

	// Processed content updates
	ProcessedContent *struct {
		DZI       *ContentData
		Tiles     *ContentData
		Thumbnail *ContentData
		IndexMap  *ContentData
	}

	// Processing info updates
	Processing *struct {
		Status          *string
		Version         *string
		FailureReason   *string
		RetryCount      *int
		LastProcessedAt *string // ISO 8601 format
	}
}

func (c *UpdateImageCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.UpdateEntityCommand.Validate()
	if !ok {
		return details, false
	}

	details = make(map[string]interface{})

	// Basic field validations
	if c.Size != nil && *c.Size < 0 {
		details["size"] = "Size cannot be negative"
	}
	if c.Width != nil && *c.Width < 0 {
		details["width"] = "Width cannot be negative"
	}
	if c.Height != nil && *c.Height < 0 {
		details["height"] = "Height cannot be negative"
	}
	if c.Format != nil && *c.Format == "" {
		details["format"] = "Format cannot be empty"
	}

	// Origin content validation
	if c.OriginContent != nil {
		if contentDetails, valid := c.validateContentData(c.OriginContent); !valid {
			details["origin_content"] = contentDetails
		}
	}

	// Processed content validation
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

	// Processing validations
	if c.Processing != nil {
		if c.Processing.Status != nil {
			status, err := vobj.NewImageStatusFromString(*c.Processing.Status)
			if err != nil {
				details["processing.status"] = "Invalid processing status value"
			} else if status == vobj.StatusDeleting {
				details["processing.status"] = "Cannot set status to DELETING via update"
			} else if status == vobj.StatusProcessed {
				// If setting to PROCESSED, ensure processed content exists or is being set
				if c.ProcessedContent == nil {
					details["processing.status"] = "Cannot set status to PROCESSED without processed content"
				}
			}
		}

		if c.Processing.Version != nil {
			version := vobj.ProcessingVersion(*c.Processing.Version)
			if !version.IsValid() {
				details["processing.version"] = "Invalid processing version"
			}
		}

		if c.Processing.FailureReason != nil && *c.Processing.FailureReason == "" {
			details["processing.failure_reason"] = "Failure reason cannot be empty"
		}

		if c.Processing.RetryCount != nil && *c.Processing.RetryCount < 0 {
			details["processing.retry_count"] = "Retry count cannot be negative"
		}
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *UpdateImageCommand) validateContentData(data *ContentData) (map[string]interface{}, bool) {
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

func (c *UpdateImageCommand) GetUpdates() map[string]interface{} {
	if _, ok := c.Validate(); !ok {
		return nil
	}

	// Base updates
	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	// Basic image field updates
	if c.Width != nil {
		updates[constants.ImageWidthField] = *c.Width
	}
	if c.Height != nil {
		updates[constants.ImageHeightField] = *c.Height
	}
	if c.Size != nil {
		updates[constants.ImageSizeField] = *c.Size
	}
	if c.Format != nil {
		updates[constants.ImageFormatField] = *c.Format
	}

	// Magnification update
	if c.Magnification != nil {
		mag := &vobj.OpticalMagnification{
			Objective:         c.Magnification.Objective,
			NativeLevel:       c.Magnification.NativeLevel,
			ScanMagnification: c.Magnification.ScanMagnification,
		}
		updates[constants.ImageMagnificationField] = mag
	}

	// Origin content update
	if c.OriginContent != nil {
		updates[constants.ImageOriginContentField] = c.contentDataToVobj(c.OriginContent)
	}

	// Processed content update
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

		updates[constants.ImageProcessedContentField] = procContent
	}

	// Processing updates (individual fields)
	if c.Processing != nil {
		if c.Processing.Status != nil {
			updates[constants.ImageProcessingStatusField] = *c.Processing.Status
		}
		if c.Processing.Version != nil {
			updates[constants.ImageProcessingVersionField] = *c.Processing.Version
		}
		if c.Processing.FailureReason != nil {
			updates[constants.ImageProcessingFailureReasonField] = *c.Processing.FailureReason
		}
		if c.Processing.RetryCount != nil {
			updates[constants.ImageProcessingRetryCountField] = *c.Processing.RetryCount
		}
		if c.Processing.LastProcessedAt != nil {
			// Parse ISO 8601 timestamp if needed
			// For now we'll skip it, or you can parse and add
			// updates[constants.ImageProcessingLastProcessedAtField] = parsedTime
		}
	}

	return updates
}

func (c *UpdateImageCommand) contentDataToVobj(data *ContentData) *vobj.Content {
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
