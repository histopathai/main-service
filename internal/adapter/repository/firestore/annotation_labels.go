package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/port"
	"google.golang.org/api/iterator"
)

// AnnotationLabelReaderImpl reads, for every region annotation of a workspace,
// the few fields a label-set summary is made of. The projection leaves the
// polygons on the server: a workspace with tens of thousands of annotations is
// a few hundred kilobytes this way instead of tens of megabytes.
type AnnotationLabelReaderImpl struct {
	client     *firestore.Client
	collection string
}

func NewAnnotationLabelReader(client *firestore.Client) *AnnotationLabelReaderImpl {
	return &AnnotationLabelReaderImpl{client: client, collection: "annotations"}
}

func (r *AnnotationLabelReaderImpl) LabelRows(ctx context.Context, wsID string) ([]port.AnnotationLabelRow, error) {
	creatorID := fields.EntityCreatorID.FirestoreName()
	parentID := fields.EntityParentID.FirestoreName()
	name := fields.EntityName.FirestoreName()
	typeID := fields.AnnotationTypeID.FirestoreName()
	resource := fields.AnnotationResource.FirestoreName()

	// Equality filters only: Firestore answers these from its single-field
	// indexes, no composite index is needed.
	iter := r.client.Collection(r.collection).
		Where(fields.AnnotationWsID.FirestoreName(), "==", wsID).
		Where(fields.EntityIsDeleted.FirestoreName(), "==", false).
		Where(fields.AnnotationIsGlobal.FirestoreName(), "==", false).
		Select(creatorID, parentID, name, typeID, resource).
		Documents(ctx)
	defer iter.Stop()

	rows := []port.AnnotationLabelRow{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			return rows, nil
		}
		if err != nil {
			if isCollectionNotFoundError(err) {
				return rows, nil
			}
			return nil, mapFirestoreError(err)
		}
		data := doc.Data()
		text := func(key string) string {
			value, _ := data[key].(string)
			return value
		}
		rows = append(rows, port.AnnotationLabelRow{
			CreatorID:        text(creatorID),
			AnnotationTypeID: text(typeID),
			Resource:         text(resource),
			Name:             text(name),
			ImageID:          text(parentID),
		})
	}
}
