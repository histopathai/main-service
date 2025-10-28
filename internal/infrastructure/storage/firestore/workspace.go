package firestore

import (
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"

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
		),
	}
}

func workspaceToFirestoreMap(w *model.Workspace) map[string]interface{} {
	m := map[string]interface{}{
		"creator_id":         w.CreatorID,
		"annotation_type_id": w.AnnotationTypeID,
		"name":               w.Name,
		"organ_type":         w.OrganType,
		"organization":       w.Organization,
		"description":        w.Description,
		"license":            w.License,
		"resource_url":       w.ResourceURL,
		"release_year":       w.ReleaseYear,
		"created_at":         w.CreatedAt,
		"updated_at":         w.UpdatedAt,
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

func workspaceMapUpdates(updates map[string]interface{}) map[string]interface{} {
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
		}
	}
	return firestoreUpdates
}
