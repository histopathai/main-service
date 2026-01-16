package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceMapper struct {
	*EntityMapper[*model.Workspace]
}

func NewWorkspaceMapper() *WorkspaceMapper {
	return &WorkspaceMapper{
		EntityMapper: NewEntityMapper[*model.Workspace](),
	}
}

func (wm *WorkspaceMapper) ToFirestoreMap(entity *model.Workspace) map[string]interface{} {

	m := wm.EntityMapper.ToFirestoreMap(entity)

	// Workspace specific fields
	m["organ_type"] = entity.OrganType.String()
	m["organization"] = entity.Organization
	m["description"] = entity.Description
	m["license"] = entity.License
	if entity.ResourceURL != nil {
		m["resource_url"] = *entity.ResourceURL
	}
	if entity.ReleaseYear != nil {
		m["release_year"] = *entity.ReleaseYear
	}
	if len(entity.AnnotationTypes) > 0 {
		m["annotation_types"] = entity.AnnotationTypes
	}

	return m
}

func (wm *WorkspaceMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Workspace, error) {

	entity, err := wm.EntityMapper.ParseEntity(doc)
	if err != nil {
		return nil, err
	}

	workspace := &model.Workspace{
		Entity: *entity,
	}

	data := doc.Data()

	if organTypeStr, ok := data["organ_type"].(string); ok {
		workspace.OrganType, err = vobj.NewOrganTypeFromString(organTypeStr)
		if err != nil {
			return nil, err
		}
	}
	if organization, ok := data["organization"].(string); ok {
		workspace.Organization = organization
	}
	if description, ok := data["description"].(string); ok {
		workspace.Description = description
	}
	if license, ok := data["license"].(string); ok {
		workspace.License = license
	}
	if resourceURL, ok := data["resource_url"].(string); ok {
		workspace.ResourceURL = &resourceURL
	}
	if releaseYear, ok := data["release_year"].(int); ok {
		workspace.ReleaseYear = &releaseYear
	}
	if annotationTypesRaw, ok := data["annotation_types"].([]interface{}); ok {
		annotationTypes := make([]string, 0, len(annotationTypesRaw))
		for _, at := range annotationTypesRaw {
			if atStr, ok := at.(string); ok {
				annotationTypes = append(annotationTypes, atStr)
			}
		}
		workspace.AnnotationTypes = annotationTypes
	}

	return workspace, nil
}

func (wm *WorkspaceMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	// Base entity updates
	mappedUpdates, err := wm.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	// Workspace specific updates
	for k, v := range updates {
		switch k {
		case constants.WorkspaceOrganTypeField:
			if organType, ok := v.(vobj.OrganType); ok {
				mappedUpdates["organ_type"] = organType.String()
			} else if organTypeStr, ok := v.(string); ok {
				mappedUpdates["organ_type"] = organTypeStr
			} else {
				return nil, errors.NewValidationError("invalid organ_type field", nil)
			}

		case constants.WorkspaceOrganizationField:
			if organization, ok := v.(string); ok {
				mappedUpdates["organization"] = organization
			} else {
				return nil, errors.NewValidationError("invalid organization field", nil)
			}

		case constants.WorkspaceDescField:
			if description, ok := v.(string); ok {
				mappedUpdates["description"] = description
			} else {
				return nil, errors.NewValidationError("invalid description field", nil)
			}

		case constants.WorkspaceLicenseField:
			if license, ok := v.(string); ok {
				mappedUpdates["license"] = license
			} else {
				return nil, errors.NewValidationError("invalid license field", nil)
			}

		case constants.WorkspaceResourceURLField:
			if resourceURL, ok := v.(*string); ok {
				mappedUpdates["resource_url"] = *resourceURL
			} else if resourceURLStr, ok := v.(string); ok {
				mappedUpdates["resource_url"] = resourceURLStr
			} else {
				return nil, errors.NewValidationError("invalid resource_url field", nil)
			}

		case constants.WorkspaceReleaseYearField:
			if releaseYear, ok := v.(*int); ok {
				mappedUpdates["release_year"] = *releaseYear
			} else if releaseYearInt, ok := v.(int); ok {
				mappedUpdates["release_year"] = releaseYearInt
			} else {
				return nil, errors.NewValidationError("invalid release_year field", nil)
			}

		case constants.WorkspaceAnnotationTypes:
			if annotationTypes, ok := v.([]string); ok {
				mappedUpdates["annotation_types"] = annotationTypes
			} else {
				return nil, errors.NewValidationError("invalid annotation_types field", nil)
			}
		}
	}

	return mappedUpdates, nil
}

func (wm *WorkspaceMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	// Base entity filters
	mappedFilters, err := wm.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	// Workspace specific filters
	for _, f := range filters {
		switch f.Field {
		case constants.WorkspaceOrganTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "organ_type",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceOrganizationField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "organization",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceLicenseField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "license",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceReleaseYearField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "release_year",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceAnnotationTypes:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "annotation_types",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.WorkspaceResourceURLField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "resource_url",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceDescField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "description",
				Operator: f.Operator,
				Value:    f.Value,
			})
		}

	}

	return mappedFilters, nil
}
