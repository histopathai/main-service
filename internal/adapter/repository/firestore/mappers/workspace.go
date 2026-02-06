// internal/adapter/repository/firestore/mappers/workspace.go
package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
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
	m[fields.WorkspaceOrganType.FirestoreName()] = entity.OrganType.String()
	m[fields.WorkspaceOrganization.FirestoreName()] = entity.Organization
	m[fields.WorkspaceDescription.FirestoreName()] = entity.Description
	m[fields.WorkspaceLicense.FirestoreName()] = entity.License

	if entity.ResourceURL != nil {
		m[fields.WorkspaceResourceURL.FirestoreName()] = *entity.ResourceURL
	}
	if entity.ReleaseYear != nil {
		m[fields.WorkspaceReleaseYear.FirestoreName()] = *entity.ReleaseYear
	}
	if len(entity.AnnotationTypes) > 0 {
		m[fields.WorkspaceAnnotationTypes.FirestoreName()] = entity.AnnotationTypes
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

	if organTypeStr, ok := data[fields.WorkspaceOrganType.FirestoreName()].(string); ok {
		workspace.OrganType, err = vobj.NewOrganTypeFromString(organTypeStr)
		if err != nil {
			return nil, err
		}
	}
	if organization, ok := data[fields.WorkspaceOrganization.FirestoreName()].(string); ok {
		workspace.Organization = organization
	}
	if description, ok := data[fields.WorkspaceDescription.FirestoreName()].(string); ok {
		workspace.Description = description
	}
	if license, ok := data[fields.WorkspaceLicense.FirestoreName()].(string); ok {
		workspace.License = license
	}
	if resourceURL, ok := data[fields.WorkspaceResourceURL.FirestoreName()].(string); ok {
		workspace.ResourceURL = &resourceURL
	}
	if releaseYear, ok := data[fields.WorkspaceReleaseYear.FirestoreName()].(int); ok {
		workspace.ReleaseYear = &releaseYear
	}
	if annotationTypesRaw, ok := data[fields.WorkspaceAnnotationTypes.FirestoreName()].([]interface{}); ok {
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
	mappedUpdates, err := wm.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for k, v := range updates {
		switch k {
		case fields.WorkspaceOrganType.DomainName():
			if organType, ok := v.(vobj.OrganType); ok {
				mappedUpdates[fields.WorkspaceOrganType.FirestoreName()] = organType.String()
			} else if organTypeStr, ok := v.(string); ok {
				mappedUpdates[fields.WorkspaceOrganType.FirestoreName()] = organTypeStr
			} else {
				return nil, errors.NewValidationError("invalid organ_type field", nil)
			}

		case fields.WorkspaceOrganization.DomainName():
			if organization, ok := v.(string); ok {
				mappedUpdates[fields.WorkspaceOrganization.FirestoreName()] = organization
			} else {
				return nil, errors.NewValidationError("invalid organization field", nil)
			}

		case fields.WorkspaceDescription.DomainName():
			if description, ok := v.(string); ok {
				mappedUpdates[fields.WorkspaceDescription.FirestoreName()] = description
			} else {
				return nil, errors.NewValidationError("invalid description field", nil)
			}

		case fields.WorkspaceLicense.DomainName():
			if license, ok := v.(string); ok {
				mappedUpdates[fields.WorkspaceLicense.FirestoreName()] = license
			} else {
				return nil, errors.NewValidationError("invalid license field", nil)
			}

		case fields.WorkspaceResourceURL.DomainName():
			if resourceURL, ok := v.(*string); ok {
				mappedUpdates[fields.WorkspaceResourceURL.FirestoreName()] = *resourceURL
			} else if resourceURLStr, ok := v.(string); ok {
				mappedUpdates[fields.WorkspaceResourceURL.FirestoreName()] = resourceURLStr
			} else {
				return nil, errors.NewValidationError("invalid resource_url field", nil)
			}

		case fields.WorkspaceReleaseYear.DomainName():
			if releaseYear, ok := v.(*int); ok {
				mappedUpdates[fields.WorkspaceReleaseYear.FirestoreName()] = *releaseYear
			} else if releaseYearInt, ok := v.(int); ok {
				mappedUpdates[fields.WorkspaceReleaseYear.FirestoreName()] = releaseYearInt
			} else {
				return nil, errors.NewValidationError("invalid release_year field", nil)
			}

		case fields.WorkspaceAnnotationTypes.DomainName():
			if annotationTypes, ok := v.([]string); ok {
				mappedUpdates[fields.WorkspaceAnnotationTypes.FirestoreName()] = annotationTypes
			} else {
				return nil, errors.NewValidationError("invalid annotation_types field", nil)
			}
		}
	}

	return mappedUpdates, nil
}

func (wm *WorkspaceMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters, err := wm.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	for _, f := range filters {
		firestoreField := fields.MapToFirestore(f.Field)
		if fields.EntityField(f.Field).IsValid() {
			continue
		}
		if fields.WorkspaceField(f.Field).IsValid() {
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    firestoreField,
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return mappedFilters, nil
}
