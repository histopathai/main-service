package firestore

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"

	"cloud.google.com/go/firestore"
)

type AnnotationTypeRepositoryImpl struct {
	*GenericRepositoryImpl[*model.AnnotationType]

	_ port.AnnotationTypeRepository // ensure interface compliance
}

func NewAnnotationTypeRepositoryImpl(client *firestore.Client, hasUniqueName bool) *AnnotationTypeRepositoryImpl {
	return &AnnotationTypeRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl[*model.AnnotationType](
			client,
			constants.AnnotationTypesCollection,
			hasUniqueName,
			annotationTypeToFirestoreDoc,
			annotatationTypeFirestoreToMap,
			annotationTypeMapUpdates,
			annotationTypeMapFilters,
		),
	}
}

func annotationTypeToFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.AnnotationType, error) {
	atModel := &model.AnnotationType{}

	entity, err := EntityFromFirestore(doc)
	if err != nil {
		return nil, err
	}
	atModel.Entity = *entity

	tag, err := TagFromFirestoreDoc(doc)
	if err != nil {
		return nil, err
	}

	atModel.Tag = *tag
	atModel.Entity = *entity

	return atModel, nil
}

func annotatationTypeFirestoreToMap(at *model.AnnotationType) map[string]interface{} {
	m_entity := EntityToFirestoreMap(&at.Entity)
	m_tag := TagToFirestoreMap(&at.Tag)

	m := make(map[string]interface{})
	for k, v := range m_entity {
		m[k] = v
	}
	for k, v := range m_tag {
		m[k] = v
	}
	return m
}

func annotationTypeMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {

	entityUpdates, err := EntityMapUpdates(updates)
	if err != nil {
		return nil, err
	}

	tagUpdates, err := TagMapUpdates(updates)
	if err != nil {
		return nil, err
	}

	firestoreUpdates := make(map[string]interface{})
	for k, v := range entityUpdates {
		firestoreUpdates[k] = v
	}
	for k, v := range tagUpdates {
		firestoreUpdates[k] = v
	}

	return firestoreUpdates, nil
}

func annotationTypeMapFilters(filters []query.Filter) ([]query.Filter, error) {
	entityMappedFilters, err := EntityMapFilter(filters)
	if err != nil {
		return nil, err
	}

	tagMappedFilters, err := TagMapFilters(filters)
	if err != nil {
		return nil, err
	}

	mappedFilters := make([]query.Filter, 0, len(entityMappedFilters)+len(tagMappedFilters))
	mappedFilters = append(mappedFilters, entityMappedFilters...)
	mappedFilters = append(mappedFilters, tagMappedFilters...)

	return mappedFilters, nil
}

func (atr *AnnotationTypeRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// AnnotationType does not have an owner field; no action needed

	return nil
}
