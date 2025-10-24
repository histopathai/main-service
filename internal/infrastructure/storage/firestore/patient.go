package firestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
	"google.golang.org/api/iterator"
)

type PatientRepositoryImpl struct {
	client     *firestore.Client
	collection string
}

func NewPatientRepositoryImpl(client *firestore.Client) *PatientRepositoryImpl {
	return &PatientRepositoryImpl{
		client:     client,
		collection: constants.PatientsCollection,
	}
}

func (pr *PatientRepositoryImpl) toFirestoreMap(p *model.Patient) map[string]interface{} {
	m := map[string]interface{}{

		"workspace_id": p.WorkspaceID,
		"anonym_name":  p.AnonymName,
	}
	if p.Age != nil {
		m["age"] = *p.Age
	}
	if p.Gender != nil {
		m["gender"] = *p.Gender
	}
	if p.Race != nil {
		m["race"] = *p.Race
	}
	if p.Disease != nil {
		m["disease"] = *p.Disease
	}
	if p.Subtype != nil {
		m["subtype"] = *p.Subtype
	}
	if p.Grade != nil {
		m["grade"] = *p.Grade
	}
	if p.History != nil {
		m["history"] = *p.History
	}
	m["created_at"] = p.CreatedAt
	m["updated_at"] = p.UpdatedAt
	return m
}

func (pr *PatientRepositoryImpl) fromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Patient, error) {
	p := &model.Patient{}
	data := doc.Data()

	p.ID = doc.Ref.ID
	p.WorkspaceID = data["workspace_id"].(string)
	p.AnonymName = data["anonym_name"].(string)

	if v, ok := data["age"].(int64); ok {
		age := int(v)
		p.Age = &age
	}

	if v, ok := data["gender"].(string); ok {
		p.Gender = &v
	}

	if v, ok := data["race"].(string); ok {
		p.Race = &v
	}

	if v, ok := data["disease"].(string); ok {
		p.Disease = &v
	}

	if v, ok := data["subtype"].(string); ok {
		p.Subtype = &v
	}

	if v, ok := data["grade"].(string); ok {
		p.Grade = &v
	}

	if v, ok := data["history"].(string); ok {
		p.History = &v
	}

	p.CreatedAt, _ = data["created_at"].(time.Time)
	p.UpdatedAt, _ = data["updated_at"].(time.Time)

	return p, nil
}

func (pr *PatientRepositoryImpl) Create(ctx context.Context, entity *model.Patient) (*model.Patient, error) {

	if entity == nil {
		return nil, errors.NewValidationError("patient entity cannot be nil", nil)
	}

	if entity.ID == "" {
		entity.ID = pr.client.Collection(pr.collection).NewDoc().ID
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	data := pr.toFirestoreMap(entity)

	_, err := pr.client.Collection(pr.collection).Doc(entity.ID).Set(ctx, data)
	if err != nil {
		return nil, errors.NewInternalError("failed to create patient", err)
	}

	return entity, nil
}

func (pr *PatientRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Patient, error) {
	docSnap, err := pr.client.Collection(pr.collection).Doc(id).Get(ctx)

	if err != nil {
		return nil, errors.NewNotFoundError("patient not found")
	}

	patient, err := pr.fromFirestoreDoc(docSnap)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse patient data", err)
	}

	return patient, nil
}

func (pr *PatientRepositoryImpl) Update(ctx context.Context, id string, updates map[string]interface{}) error {

	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.PatientWorkspaceIDField:
			firestoreUpdates["workspace_id"] = value
		case constants.PatientAnonymNameField:
			firestoreUpdates["anonym_name"] = value
		case constants.PatientAgeField:
			firestoreUpdates["age"] = value
		case constants.PatientGenderField:
			firestoreUpdates["gender"] = value
		case constants.PatientRaceField:
			firestoreUpdates["race"] = value
		case constants.PatientDiseaseField:
			firestoreUpdates["disease"] = value
		case constants.PatientSubtypeField:
			firestoreUpdates["subtype"] = value
		case constants.PatientGradeField:
			firestoreUpdates["grade"] = value
		case constants.PatientHistoryField:
			firestoreUpdates["history"] = value
		default:
			return errors.NewValidationError("invalid field for update: "+key, nil)
		}
	}

	firestoreUpdates["updated_at"] = time.Now()

	_, err := pr.client.Collection(pr.collection).Doc(id).Set(ctx, firestoreUpdates, firestore.MergeAll)
	if err != nil {
		return errors.NewInternalError("failed to update patient", err)
	}

	return nil
}

func (pr *PatientRepositoryImpl) GetByCriteria(ctx context.Context, filters []sharedQuery.Filter, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Patient], error) {

	query := pr.client.Collection(pr.collection).Query

	for _, f := range filters {
		query = query.Where(f.Field, string(f.Operator), f.Value)
	}

	// Apply pagination
	if paginationOpts == nil {
		paginationOpts = &sharedQuery.Pagination{
			Limit:  10,
			Offset: 0,
		}
	}

	query = query.Offset(paginationOpts.Offset).Limit(paginationOpts.Limit + 1)

	iter := query.Documents(ctx)
	defer iter.Stop()

	patients := []*model.Patient{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.NewInternalError("failed to retrieve patients", err)
		}

		p, err := pr.fromFirestoreDoc(doc)
		if err != nil {
			continue
		}

		patients = append(patients, p)
	}

	hasMore := false
	if len(patients) > paginationOpts.Limit {
		hasMore = true
		patients = patients[:len(patients)-1]
	}

	return &sharedQuery.Result[model.Patient]{
		Data:    patients,
		Total:   0, // Total count can be implemented if needed
		Limit:   paginationOpts.Limit,
		Offset:  paginationOpts.Offset,
		HasMore: hasMore,
	}, nil

}

func (pr *PatientRepositoryImpl) GetByWorkSpaceID(ctx context.Context, workspaceID string, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Patient], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    "workspace_id",
			Operator: "==",
			Value:    workspaceID,
		},
	}

	return pr.GetByCriteria(ctx, filters, paginationOpts)
}
