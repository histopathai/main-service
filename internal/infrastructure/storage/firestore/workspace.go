package firestore

import (
	"fmt"
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"

	"cloud.google.com/go/firestore"
)

type WorkspaceRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Workspace]
	_ repository.WorkspaceRepository // ensure interface compliance
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
	// Önce zorunlu (pointer olmayan) alanları ekle
	m := map[string]interface{}{
		"creator_id":   w.CreatorID,
		"name":         w.Name,
		"organ_type":   w.OrganType,
		"organization": w.Organization,
		"description":  w.Description,
		"license":      w.License,
		"created_at":   w.CreatedAt,
		"updated_at":   w.UpdatedAt,
	}

	if w.AnnotationTypeID != nil {
		m["annotation_type_id"] = *w.AnnotationTypeID
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

	w.ID = doc.Ref.ID
	w.CreatorID = data["creator_id"].(string)

	if v, ok := data["annotation_type_id"].(string); ok {
		w.AnnotationTypeID = &v
	}

	w.Name = data["name"].(string)
	w.OrganType = data["organ_type"].(string)
	w.Organization = data["organization"].(string)
	w.Description = data["description"].(string)
	w.License = data["license"].(string)

	if v, ok := data["resource_url"].(string); ok {
		w.ResourceURL = &v
	}

	if v, ok := data["release_year"].(int64); ok {
		year := int(v)
		w.ReleaseYear = &year
	}

	w.CreatedAt, _ = data["created_at"].(time.Time)
	w.UpdatedAt, _ = data["updated_at"].(time.Time)

	return w, nil
}

func workspaceMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.WorkspaceNameField:
			firestoreUpdates["name"] = value
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
		case constants.WorkspaceAnnotationTypeIDField:
			firestoreUpdates["annotation_type_id"] = value
		default:
			return nil, fmt.Errorf("unknown update field: %s", key)
		}
	}
	return firestoreUpdates, nil
}

func workspaceMapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))
	for _, f := range filters {
		switch f.Field {
		case constants.WorkspaceCreatorIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "creator_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceNameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "name",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WorkspaceAnnotationTypeIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "annotation_type_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
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
		case constants.CreatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "created_at",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.UpdatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "updated_at",
				Operator: f.Operator,
				Value:    f.Value,
			})
		default:
			//ignore unknown fields
			return nil, fmt.Errorf("unknown filter field: %s", f.Field)

		}
	}
	return mappedFilters, nil
}
