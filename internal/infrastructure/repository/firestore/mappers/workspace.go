package firestoreMappers

import (
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceMapper struct {
	entityMapper *EntityMapper
}

func NewWorkspaceMapper() *WorkspaceMapper {
	return &WorkspaceMapper{
		entityMapper: &EntityMapper{},
	}
}

func (wm *WorkspaceMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Workspace, error) {
	data := doc.Data()

	if data == nil {
		return nil, fmt.Errorf("firestore document data is nil")
	}

	entity, err := wm.entityMapper.FromFirestoreDoc(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity from firestore document: %w", err)
	}

	workspace := &model.Workspace{
		Entity:       entity,
		OrganType:    data["organ_type"].(string),
		Organization: data["organization"].(string),
		License:      data["license"].(string),
	}

	if data["resource_url"] != nil {
		resourceURL := data["resource_url"].(string)
		workspace.ResourceURL = &resourceURL
	}

	if data["release_year"] != nil {
		releaseYear := int(data["release_year"].(int64))
		workspace.ReleaseYear = &releaseYear
	}

	if data["description"] != nil {
		workspace.Description = data["description"].(string)
	}

	return workspace, nil
}

func (wm *WorkspaceMapper) ToFirestoreMap(workspace *model.Workspace) map[string]interface{} {

	m := wm.entityMapper.ToFirestoreMap(workspace.Entity)

	m["organ_type"] = workspace.OrganType
	m["organization"] = workspace.Organization
	m["license"] = workspace.License

	if workspace.ResourceURL != nil {
		m["resource_url"] = *workspace.ResourceURL
	}

	if workspace.ReleaseYear != nil {
		m["release_year"] = *workspace.ReleaseYear
	}

	if workspace.Description != "" {
		m["description"] = workspace.Description
	}

	return m
}

func (wm *WorkspaceMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	firestoreUpdates, err := wm.entityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for key, value := range updates {
		switch key {
		case constants.OrganTypeField:
			firestoreUpdates["organ_type"] = value
			delete(updates, constants.OrganTypeField)

		case constants.OrganizationField:
			firestoreUpdates["organization"] = value
			delete(updates, constants.OrganizationField)

		case constants.LicenseField:
			firestoreUpdates["license"] = value
			delete(updates, constants.LicenseField)

		case constants.ResourceURLField:
			firestoreUpdates["resource_url"] = value
			delete(updates, constants.ResourceURLField)

		case constants.ReleaseYearField:
			firestoreUpdates["release_year"] = value
			delete(updates, constants.ReleaseYearField)

		case constants.DescField:
			firestoreUpdates["description"] = value
			delete(updates, constants.DescField)
		}
	}

	return firestoreUpdates, nil
}

func (wm *WorkspaceMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	mappedFilters, err := wm.entityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	unprocessedIdx := 0
	for i, filter := range filters {
		processed := false

		switch filter.Field {
		case constants.OrganTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "organ_type",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.OrganizationField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "organization",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.LicenseField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "license",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.ResourceURLField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "resource_url",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.ReleaseYearField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "release_year",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.DescField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "description",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		}

		if !processed {
			filters[unprocessedIdx] = filters[i]
			unprocessedIdx++
		}
	}

	for i := unprocessedIdx; i < len(filters); i++ {
		filters[i] = query.Filter{}
	}

	return mappedFilters, nil
}
