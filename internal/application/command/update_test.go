package command_test

import (
	"strings"
	"testing"
	"time"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateAnnotationReviewCommand_GetUpdates(t *testing.T) {
	status := fields.ReviewStatusApproved
	comments := "Looks good"
	cmd := command.UpdateAnnotationReviewCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{ID: "review-1"},
		Status:              &status,
		Comments:            &comments,
		ModifiedValue:       float64(5.0),
		ModifiedPolygon: &[]command.CommandPoint{
			{X: 1, Y: 2},
			{X: 3, Y: 4},
			{X: 5, Y: 6},
		},
	}

	updates := cmd.GetUpdates()
	require.NotNil(t, updates)

	assert.Equal(t, status, updates[fields.AnnotationReviewStatus.DomainName()])
	assert.Equal(t, comments, updates[fields.AnnotationReviewComments.DomainName()])
	assert.Equal(t, float64(5.0), updates[fields.AnnotationReviewModifiedValue.DomainName()])

	poly, ok := updates[fields.AnnotationReviewModifiedPolygon.DomainName()]
	require.True(t, ok)
	assert.Len(t, poly, 3)
}

func TestUpdateImageCommandUnsuitable(t *testing.T) {
	at := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	yes, no, creator := true, false, "u"
	cmd := command.UpdateImageCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{ID: "i1", CreatorID: &creator},
		Unsuitable:          &yes, UnsuitableBy: "p1", UnsuitableNote: "Kesit bozuk",
		Now: func() time.Time { return at },
	}
	u := cmd.GetUpdates()
	assert.Equal(t, true, u["Unsuitable"])
	assert.Equal(t, "p1", u["UnsuitableBy"])
	assert.Equal(t, "Kesit bozuk", u["UnsuitableNote"])
	assert.Equal(t, at, u["UnsuitableAt"])

	cmd.Unsuitable = &no
	u = cmd.GetUpdates()
	assert.Equal(t, false, u["Unsuitable"])
	assert.Equal(t, "", u["UnsuitableBy"])
	assert.Equal(t, "", u["UnsuitableNote"])
	assert.Nil(t, u["UnsuitableAt"])

	cmd.Unsuitable = nil
	_, set := cmd.GetUpdates()["Unsuitable"]
	assert.False(t, set, "left alone when not given")

	cmd.UnsuitableNote = strings.Repeat("ç", command.UnsuitableNoteMaxLen+1)
	_, ok := cmd.Validate()
	require.False(t, ok, "long note refused")
}
