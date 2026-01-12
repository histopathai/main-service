package firestore

import (
	"fmt"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"

	"cloud.google.com/go/firestore"
)

type WorkspaceRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Workspace]
	_ port.WorkspaceRepository // ensure interface compliance
}

func NewWorkspaceRepositoryImpl(client *firestore.Client, hasUniqueName bool) *WorkspaceRepositoryImpl {
	return &WorkspaceRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl[*model.Workspace](
			client,
			constants.WorkspaceCollection,
			hasUniqueName,
			workspaceFromFirestoreDoc,
			workspaceToFirestoreMap,
			workspaceMapUpdates,
			workspaceMapFilters,
		),
	}
}

func workspaceToFirestoreMap(w *model.Workspace) map[string]interface{} {
	m := EntityToFirestoreMap(&w.Entity)
	m["organ_type"] = w.OrganType
	m["organization"] = w.Organization
	m["description"] = w.Description
	m["license"] = w.License

	if len(w.AnnotationTypes) > 0 {
		m["annotation_types"] = w.AnnotationTypes
	}
	if w.ResourceURL != nil {
		m["resource_url"] = *w.ResourceURL
	}
	if w.ReleaseYear != nil {
		m["release_year"] = *w.ReleaseYear
	}

	return m
}

func workspaceFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Workspace, error) {
	w := &model.Workspace{}
	data := doc.Data()

	entity, err := EntityFromFirestore(doc)
	if err != nil {
		return nil, err
	}
	w.Entity = *entity

	// Use safe type assertions to avoid panics on missing fields
	if v, ok := data["organ_type"].(string); ok {
		w.OrganType = v
	}
	if v, ok := data["organization"].(string); ok {
		w.Organization = v
	}
	if v, ok := data["description"].(string); ok {
		w.Description = v
	}
	if v, ok := data["license"].(string); ok {
		w.License = v
	}

	if at, ok := data["annotation_types"].([]interface{}); ok {
		w.AnnotationTypes = make([]string, len(at))
		for i, v := range at {
			if strVal, ok := v.(string); ok {
				w.AnnotationTypes[i] = strVal
			}
		}
	}
	if v, ok := data["resource_url"].(string); ok {
		w.ResourceURL = &v
	}

	if v, ok := data["release_year"].(int64); ok {
		year := int(v)
		w.ReleaseYear = &year
	}

	return w, nil
}

func workspaceMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	firestoreUpdates, err := EntityMapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for key, value := range updates {
		if EntityFields[key] {
			continue
		}
		switch key {
		case constants.WorkspaceOrganTypeField:
			firestoreUpdates["organ_type"] = value
		case constants.WorkspaceOrganizationField:
			firestoreUpdates["organization"] = value
		case constants.WorkspaceDescField:
			firestoreUpdates["description"] = value
		case constants.WorkspaceLicenseField:
			firestoreUpdates["license"] = value
		case constants.WorkspaceResourceURLField:
			firestoreUpdates["resource_url"] = value
		case constants.WorkspaceReleaseYearField:
			firestoreUpdates["release_year"] = value
		case constants.WorkspaceAnnotationTypes:
			firestoreUpdates["annotation_types"] = value
		default:
			return nil, fmt.Errorf("unknown update field: %s", key)
		}
	}
	return firestoreUpdates, nil
}

func workspaceMapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return []query.Filter{}, nil
	}
	mappedFilters, err := EntityMapFilter(filters)
	if err != nil {
		return nil, err
	}

	for _, f := range filters {
		if EntityFields[f.Field] {
			continue
		}
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
		case constants.WorkspaceDescField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "description",
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
		case constants.WorkspaceAnnotationTypes:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "annotation_types",
				Operator: f.Operator,
				Value:    f.Value,
			})
		default:
			return nil, fmt.Errorf("unknown filter field: %s", f.Field)
		}
	}

	return mappedFilters, nil
}
