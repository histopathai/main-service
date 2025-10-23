package handler

import (
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/models"
)

/*------------------ Workspace Request Handler ------------------*/

type CreateWorkspaceRequest struct {
	Name             string  `json:"name"  validate:"required"`
	OrganType        string  `json:"organ_type"  validate:"required"`
	CreatorID        string  `json:"creator_id"  validate:"required"`
	Description      string  `json:"description" validate:"required"`
	License          string  `json:"license" validate:"required"`
	Organization     string  `json:"organization" validate:"required"`
	AnnotationTypeID string  `json:"annotation_type_id" validate:"required"`
	ResourceURL      *string `json:"resource_url,omitempty" validate:"omitempty,url"`
	ReleaseYear      *int    `json:"release_year,omitempty" validate:"omitempty,min=1900,max=2100"`
	ReleaseVersion   *string `json:"release_version,omitempty" validate:"omitempty"`
}

func (r *CreateWorkspaceRequest) ToModel() *models.Workspace {
	workspace := &models.Workspace{
		Name:             r.Name,
		OrganType:        r.OrganType,
		CreatorID:        r.CreatorID,
		Description:      r.Description,
		License:          r.License,
		Organization:     r.Organization,
		AnnotationTypeID: r.AnnotationTypeID,
	}
	if r.ResourceURL != nil {
		workspace.ResourceURL = *r.ResourceURL
	}
	if r.ReleaseYear != nil {
		workspace.ReleaseYear = *r.ReleaseYear
	}
	if r.ReleaseVersion != nil {
		workspace.ReleaseVersion = *r.ReleaseVersion
	}
	return workspace
}

type UpdateWorkspaceRequest struct {
	Name           *string `json:"name,omitempty" validate:"omitempty"`
	OrganType      *string `json:"organ_type,omitempty" validate:"omitempty"`
	Description    *string `json:"description,omitempty" validate:"omitempty"`
	License        *string `json:"license,omitempty" validate:"omitempty"`
	Organization   *string `json:"organization,omitempty" validate:"omitempty"`
	ResourceURL    *string `json:"resource_url,omitempty" validate:"omitempty,url"`
	ReleaseYear    *int    `json:"release_year,omitempty" validate:"omitempty,min=1900,max=2100"`
	ReleaseVersion *string `json:"release_version,omitempty" validate:"omitempty"`
}

func (r *UpdateWorkspaceRequest) ToUpdateMap() map[string]interface{} {
	updates := make(map[string]interface{})
	if r.Name != nil {
		updates["name"] = *r.Name
	}
	if r.OrganType != nil {
		updates["organ_type"] = *r.OrganType
	}
	if r.Description != nil {
		updates["description"] = *r.Description
	}
	if r.License != nil {
		updates["license"] = *r.License
	}
	if r.Organization != nil {
		updates["organization"] = *r.Organization
	}
	if r.ResourceURL != nil {
		updates["resource_url"] = *r.ResourceURL
	}
	if r.ReleaseYear != nil {
		updates["release_year"] = *r.ReleaseYear
	}
	if r.ReleaseVersion != nil {
		updates["release_version"] = *r.ReleaseVersion
	}
	return updates
}

type ListWorkspacesRequest struct {
	OrganType    string `form:"organ_type"`
	License      string `form:"license"`
	Organization string `form:"organization"`
	CreatorID    string `form:"creator_id"`

	ReleaseYearGT  int `form:"release_year_gt"`  // greater than
	ReleaseYearGTE int `form:"release_year_gte"` // greater than or equal
	ReleaseYearLT  int `form:"release_year_lt"`  // less than
	ReleaseYearLTE int `form:"release_year_lte"` // less than or equal

}

func (r *ListWorkspacesRequest) ToFilterMap() []repository.Filter {
	filters := make([]repository.Filter, 0)
	if r.OrganType != "" {
		filters = append(filters, repository.Filter{
			Field: "organ_type", Op: repository.OpEqual, Value: r.OrganType,
		})
	}
	if r.License != "" {
		filters = append(filters, repository.Filter{
			Field: "license", Op: repository.OpEqual, Value: r.License,
		})
	}
	if r.Organization != "" {
		filters = append(filters, repository.Filter{
			Field: "organization", Op: repository.OpEqual, Value: r.Organization,
		})
	}
	if r.CreatorID != "" {
		filters = append(filters, repository.Filter{
			Field: "creator_id", Op: repository.OpEqual, Value: r.CreatorID,
		})
	}

	if r.ReleaseYearGT != 0 {
		filters = append(filters, repository.Filter{
			Field: "release_year", Op: repository.OpGreaterThan, Value: r.ReleaseYearGT,
		})
	}
	if r.ReleaseYearGTE != 0 {
		filters = append(filters, repository.Filter{
			Field: "release_year", Op: repository.OpGreaterThanOrEq, Value: r.ReleaseYearGTE,
		})
	}
	if r.ReleaseYearLT != 0 {
		filters = append(filters, repository.Filter{
			Field: "release_year", Op: repository.OpLessThan, Value: r.ReleaseYearLT,
		})
	}
	if r.ReleaseYearLTE != 0 {
		filters = append(filters, repository.Filter{
			Field: "release_year", Op: repository.OpLessThanOrEq, Value: r.ReleaseYearLTE,
		})
	}

	return filters
}

/*------------------ Patient Request Handler ------------------*/

type CreatePatientRequest struct {
	AnonymousName *string `json:"anonymous_name,omitempty" validate:"omitempty"`
	Age           *int    `json:"age,omitempty" validate:"omitempty,min=0"`
	Gender        *string `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Race          *string `json:"race,omitempty" validate:"omitempty"`
	Disease       *string `json:"disease,omitempty" validate:"omitempty"`
	SubType       *string `json:"subtype,omitempty" validate:"omitempty"`
	Grade         *string `json:"grade,omitempty" validate:"omitempty"`
	History       *string `json:"history,omitempty" validate:"omitempty"`
	WorkspaceID   string  `json:"workspace_id" validate:"required"`
}

func (r *CreatePatientRequest) ToModel() *models.Patient {
	patient := &models.Patient{
		WorkspaceID: r.WorkspaceID,
	}
	if r.AnonymousName != nil {
		patient.AnonymousName = *r.AnonymousName
	}
	if r.Age != nil {
		patient.Age = r.Age
	}
	if r.Gender != nil {
		patient.Gender = r.Gender
	}
	if r.Race != nil {
		patient.Race = r.Race
	}
	if r.Disease != nil {
		patient.Disease = r.Disease
	}
	if r.SubType != nil {
		patient.SubType = r.SubType
	}
	if r.Grade != nil {
		patient.Grade = r.Grade
	}
	if r.History != nil {
		patient.History = r.History
	}
	return patient
}
