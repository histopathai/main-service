package firestore

import (
	"context"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type WorkspaceRepositoryImpl struct {
	client     *firestore.Client
	collection string
}

func NewWorkspaceRepositoryImpl(client *firestore.Client) *WorkspaceRepositoryImpl {
	return &WorkspaceRepositoryImpl{
		client:     client,
		collection: constants.WorkspaceCollection,
	}
}

func (r *WorkspaceRepositoryImpl) toFirestoreMap(w *model.Workspace) map[string]interface{} {
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

func (r *WorkspaceRepositoryImpl) fromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Workspace, error) {
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

	if v, ok := data["release_year"].(int); ok {
		year := int(v)
		w.ReleaseYear = &year
	}

	w.CreatedAt, _ = data["created_at"].(time.Time)
	w.UpdatedAt, _ = data["updated_at"].(time.Time)

	return w, nil
}

func (r *WorkspaceRepositoryImpl) Create(ctx context.Context, entity *model.Workspace) (*model.Workspace, error) {

	if entity == nil {
		return nil, errors.NewValidationError("workspace entity cannot be nil", nil)
	}

	if entity.ID == "" {
		entity.ID = r.client.Collection(r.collection).NewDoc().ID
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	data := r.toFirestoreMap(entity)

	_, err := r.client.Collection(r.collection).Doc(entity.ID).Set(ctx, data)
	if err != nil {
		return nil, errors.NewInternalError("failed to create workspace", err)
	}

	return entity, nil

}

func (r *WorkspaceRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Workspace, error) {
	docSnap, err := r.client.Collection(r.collection).Doc(id).Get(ctx)

	if err != nil {
		return nil, errors.NewNotFoundError("workspace not found")
	}

	workspace, err := r.fromFirestoreDoc(docSnap)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse workspace data", err)
	}

	return workspace, nil
}

func (r *WorkspaceRepositoryImpl) Update(ctx context.Context, id string, updates map[string]interface{}) error {

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
			return errors.NewValidationError("invalid field for update: "+key, nil)
		}
	}

	if len(firestoreUpdates) == 0 {
		return errors.NewValidationError("no valid fields provided for update", nil)
	}

	firestoreUpdates["updated_at"] = time.Now()

	_, err := r.client.Collection(r.collection).Doc(id).Set(ctx, firestoreUpdates, firestore.MergeAll)
	if err != nil {
		return errors.NewInternalError("failed to update workspace", err)
	}

	return nil
}

func (r *WorkspaceRepositoryImpl) WithTx(ctx context.Context, fn func(ctx context.Context, tx repository.Transaction) error) error {
	err := r.client.RunTransaction(ctx, func(ctx context.Context, fstx *firestore.Transaction) error {

		tx := NewFirestoreTransaction(r.client, fstx)

		return fn(ctx, tx)
	})

	if err != nil {
		return errors.NewInternalError("firestore transaction failed", err)
	}

	return nil
}

func (r *WorkspaceRepositoryImpl) GetByCriteria(ctx context.Context, filters []sharedQuery.Filter, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Workspace], error) {
	query := r.client.Collection(r.collection).Query

	for _, f := range filters {
		query = query.Where(f.Field, string(f.Operator), f.Value)
	}

	if paginationOpts == nil {
		paginationOpts = &sharedQuery.Pagination{
			Limit:  10,
			Offset: 0,
		}
	}
	query = query.Offset(paginationOpts.Offset).Limit(paginationOpts.Limit + 1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	workspaces := []*model.Workspace{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.NewInternalError("failed to retrieve workspaces", err)
		}

		ws, err := r.fromFirestoreDoc(doc)
		if err != nil {
			continue
		}

		workspaces = append(workspaces, ws)
	}

	hasmore := false
	if len(workspaces) > paginationOpts.Limit {
		hasmore = true
		workspaces = workspaces[:len(workspaces)-1]
	}

	return &sharedQuery.Result[model.Workspace]{
		Data:    workspaces,
		Total:   0, // Total count can be implemented if needed
		Limit:   paginationOpts.Limit,
		Offset:  paginationOpts.Offset,
		HasMore: hasmore,
	}, nil
}

func (r *WorkspaceRepositoryImpl) GetByCreatorID(ctx context.Context, creatorID string) (*sharedQuery.Result[model.Workspace], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    "creator_id",
			Operator: sharedQuery.OpEqual,
			Value:    creatorID,
		},
	}

	return r.GetByCriteria(ctx, filters, nil)
}

func (r *WorkspaceRepositoryImpl) GetByeOrganType(ctx context.Context, organType string) (*sharedQuery.Result[model.Workspace], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    "organ_type",
			Operator: sharedQuery.OpEqual,
			Value:    organType,
		},
	}

	return r.GetByCriteria(ctx, filters, nil)
}
