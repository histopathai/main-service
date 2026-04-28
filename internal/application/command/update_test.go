package command_test

import (
	"testing"

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
		Status:          &status,
		Comments:        &comments,
		ModifiedValue:   float64(5.0),
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
