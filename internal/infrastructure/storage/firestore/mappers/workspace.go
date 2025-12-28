package firestoreMappers

import (
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceMapper struct{}

func (wm *WorkspaceMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Workspace, error) {
	fw := &model.Workspace{}

	data := doc.Data()

	if data == nil {
		return nil, fmt.Errorf("firestore document data is nil")
	}

	beMapper := &BaseEntityMapper{}
	baseEntity, _ := beMapper.FromFirestoreDoc(doc)

	if baseEntity == nil {
		return nil, fmt.Errorf("failed to map base entity from firestore document")
	}

	fw.BaseEntity = *baseEntity
	fw.OrganType = data["organ_type"].(string)
	fw.Organization = data["organization"].(string)
	fw.License = data["license"].(string)

	if data["annotation_type_id"] != nil {
		annotationTypeID := data["annotation_type_id"].(string)
		fw.AnnotationTypeID = &annotationTypeID
	}
	if data["resource_url"] != nil {
		resourceURL := data["resource_url"].(string)
		fw.ResourceURL = &resourceURL
	}
	if data["release_year"] != nil {
		releaseYear := int(data["release_year"].(int64))
		fw.ReleaseYear = &releaseYear
	}

	if data["description"] != nil {
		fw.Description = data["description"].(string)
	}

	return fw, nil
}

func (wm *WorkspaceMapper) ToFirestoreMap(w *model.Workspace) map[string]interface{} {
	beMapper := &BaseEntityMapper{}
	m := beMapper.ToFirestoreMap(&w.BaseEntity)

	m["organ_type"] = w.OrganType
	m["organization"] = w.Organization
	m["license"] = w.License

	if w.AnnotationTypeID != nil {
		m["annotation_type_id"] = *w.AnnotationTypeID
	}
	if w.ResourceURL != nil {
		m["resource_url"] = *w.ResourceURL
	}
	if w.ReleaseYear != nil {
		m["release_year"] = *w.ReleaseYear
	}
	if w.Description != "" {
		m["description"] = w.Description
	}

	return m
}

func (wm *WorkspaceMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	beMapper := &BaseEntityMapper{}
	firestoreUpdates, _ := beMapper.MapUpdates(updates)

	// Clear BaseEntity related fields from updates to avoid duplication
	delete(updates, constants.NameField)
	delete(updates, constants.DeletedField)
	delete(updates, constants.UpdatedAtField)
	delete(updates, constants.CreatorIDField)

	for key, value := range updates {
		switch key {
		case constants.WorkspaceOrganTypeField:
			firestoreUpdates["organ_type"] = value
		case constants.WorkspaceOrganizationField:
			firestoreUpdates["organization"] = value
		case constants.WorkspaceLicenseField:
			firestoreUpdates["license"] = value
		case constants.WorkspaceAnnotationTypeIDField:
			firestoreUpdates["annotation_type_id"] = value
		case constants.WorkspaceResourceURLField:
			firestoreUpdates["resource_url"] = value
		case constants.WorkspaceReleaseYearField:
			firestoreUpdates["release_year"] = value
		case constants.WorkspaceDescField:
			firestoreUpdates["description"] = value
		default:
			return nil, fmt.Errorf("unknown field in workspace updates: %s", key)
		}
	}

	return firestoreUpdates, nil
}

func (wm *WorkspaceMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	beMapper := &BaseEntityMapper{}
	mappedFilters, _ := beMapper.MapFilters(filters)

	// Clear BaseEntity related fields from filters to avoid duplication
	filteredFilters := make([]query.Filter, 0)
	for _, filter := range filters {
		if filter.Field == constants.NameField ||
			filter.Field == constants.DeletedField ||
			filter.Field == constants.CreatorIDField {
			continue
		}
		filteredFilters = append(filteredFilters, filter)
	}

	for _, f := range filteredFilters {
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
		case constants.WorkspaceAnnotationTypeIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "annotation_type_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceLicenseField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "license",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceResourceURLField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "resource_url",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceReleaseYearField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "release_year",
				Operator: f.Operator,
				Value:    f.Value,
			})
		default:
			return nil, fmt.Errorf("unknown filter field for workspace: %s", f.Field)
		}
	}

	return mappedFilters, nil
}
