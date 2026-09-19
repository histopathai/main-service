package queries

import (
	"context"
	"errors"
	"testing"

	"github.com/histopathai/main-service/internal/port"
	"github.com/stretchr/testify/assert"
)

type fakeLabelReader struct {
	rows []port.AnnotationLabelRow
	err  error
}

func (f fakeLabelReader) LabelRows(context.Context, string) ([]port.AnnotationLabelRow, error) {
	return f.rows, f.err
}

func row(creator, annotationType, resource, name, image string) port.AnnotationLabelRow {
	return port.AnnotationLabelRow{CreatorID: creator, AnnotationTypeID: annotationType, Resource: resource, Name: name, ImageID: image}
}

func TestLabelSetsByWsID(t *testing.T) {
	const imported = "111111111111111111111"
	q := NewAnnotationQuery(nil, fakeLabelReader{rows: []port.AnnotationLabelRow{
		// An imported dataset and a person answered the same question, on different images.
		row(imported, "t-gleason", "imported", "Gleason Pattern", "img-b"),
		row(imported, "t-gleason", "imported", "Gleason Pattern", "img-a"),
		row(imported, "t-gleason", "imported", "Gleason Pattern", "img-a"),
		row("selva", "t-gleason", "manual", "Gleason Pattern", "img-c"),
		// A type whose annotations carry file names: no agreed name.
		row(imported, "t-score", "imported", "slide_1.json", "img-a"),
		row(imported, "t-score", "imported", "slide_2.json", "img-b"),
		// One person with hand-drawn and model-made polygons of one type.
		row("selva", "t-tumor", "manual", "Tümör Bölgesi", "img-c"),
		row("selva", "t-tumor", "model", "Tümör Bölgesi", "img-c"),
		// Drawn before the resource field existed: counts as manual, not as a kind of its own.
		row("selva", "t-gleason", "", "Gleason Pattern", "img-d"),
	}})

	sets, err := q.LabelSetsByWsID(context.Background(), "ws")
	assert.NoError(t, err)
	assert.Equal(t, []port.LabelSet{
		{CreatorID: imported, AnnotationTypeID: "t-gleason", Resources: []string{"imported"}, Name: "Gleason Pattern",
			Polygons: 3, ImageIDs: []string{"img-a", "img-b"}},
		{CreatorID: imported, AnnotationTypeID: "t-score", Resources: []string{"imported"}, Name: "",
			Polygons: 2, ImageIDs: []string{"img-a", "img-b"}},
		{CreatorID: "selva", AnnotationTypeID: "t-gleason", Resources: []string{"manual"}, Name: "Gleason Pattern",
			Polygons: 2, ImageIDs: []string{"img-c", "img-d"}},
		{CreatorID: "selva", AnnotationTypeID: "t-tumor", Resources: []string{"manual", "model"}, Name: "Tümör Bölgesi",
			Polygons: 2, ImageIDs: []string{"img-c"}},
	}, sets)
}

func TestLabelSetsByWsIDEmptyAndError(t *testing.T) {
	sets, err := NewAnnotationQuery(nil, fakeLabelReader{}).LabelSetsByWsID(context.Background(), "ws")
	assert.NoError(t, err)
	assert.Empty(t, sets)

	_, err = NewAnnotationQuery(nil, fakeLabelReader{err: errors.New("boom")}).LabelSetsByWsID(context.Background(), "ws")
	assert.Error(t, err)
}
