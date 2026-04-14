package command_test

import (
	"testing"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Validate()
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateAnnotationReviewCommand_Validate_RequiredFields(t *testing.T) {
	tests := []struct {
		name        string
		cmd         command.CreateAnnotationReviewCommand
		expectValid bool
		expectKeys  []string
	}{
		{
			name:        "empty command",
			cmd:         command.CreateAnnotationReviewCommand{},
			expectValid: false,
			expectKeys:  []string{"annotation_id", "reviewer_id", "status"},
		},
		{
			name: "missing annotation_id",
			cmd: command.CreateAnnotationReviewCommand{
				ReviewerID: "user-1",
				Status:     "approved",
			},
			expectValid: false,
			expectKeys:  []string{"annotation_id"},
		},
		{
			name: "missing reviewer_id",
			cmd: command.CreateAnnotationReviewCommand{
				AnnotationID: "anno-1",
				Status:       "approved",
			},
			expectValid: false,
			expectKeys:  []string{"reviewer_id"},
		},
		{
			name: "invalid status",
			cmd: command.CreateAnnotationReviewCommand{
				AnnotationID: "anno-1",
				ReviewerID:   "user-1",
				Status:       "unknown_status",
			},
			expectValid: false,
			expectKeys:  []string{"status"},
		},
		{
			name: "modified status without modified fields",
			cmd: command.CreateAnnotationReviewCommand{
				AnnotationID: "anno-1",
				ReviewerID:   "user-1",
				Status:       "modified",
			},
			expectValid: false,
			expectKeys:  []string{"modified"},
		},
		{
			name: "approved - valid",
			cmd: command.CreateAnnotationReviewCommand{
				AnnotationID: "anno-1",
				ReviewerID:   "user-1",
				Status:       "approved",
			},
			expectValid: true,
		},
		{
			name: "rejected - valid",
			cmd: command.CreateAnnotationReviewCommand{
				AnnotationID: "anno-1",
				ReviewerID:   "user-1",
				Status:       "rejected",
			},
			expectValid: true,
		},
		{
			name: "modified with ModifiedValue - valid",
			cmd: command.CreateAnnotationReviewCommand{
				AnnotationID:  "anno-1",
				ReviewerID:    "user-1",
				Status:        "modified",
				ModifiedValue: float64(5.0),
			},
			expectValid: true,
		},
		{
			name: "modified with ModifiedPolygon - valid",
			cmd: command.CreateAnnotationReviewCommand{
				AnnotationID: "anno-1",
				ReviewerID:   "user-1",
				Status:       "modified",
				ModifiedPolygon: &[]command.CommandPoint{
					{X: 0, Y: 0},
					{X: 1, Y: 0},
					{X: 1, Y: 1},
				},
			},
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			details, ok := tt.cmd.Validate()
			assert.Equal(t, tt.expectValid, ok)
			if !tt.expectValid {
				require.NotNil(t, details)
				for _, key := range tt.expectKeys {
					assert.Contains(t, details, key, "expected error key %q in details", key)
				}
			} else {
				assert.Nil(t, details)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ToEntity()
// ─────────────────────────────────────────────────────────────────────────────

// Note: ToEntity auto-populates CreateEntityCommand fields (EntityType, ParentID, ParentType, CreatorID, Name)
// before calling the base ToEntity, so we only need AnnotationID, ReviewerID, Status in the top-level cmd.
// However, the base Validate() runs inside ToEntity after we set those fields.
// We verify the returned entity has the expected values.
func TestCreateAnnotationReviewCommand_ToEntity_SetsCorrectFields(t *testing.T) {
	comment := "Looks good"
	cmd := command.CreateAnnotationReviewCommand{
		AnnotationID:  "anno-abc",
		ReviewerID:    "reviewer-xyz",
		Status:        "approved",
		Comments:      &comment,
		ModifiedValue: nil,
	}

	entity, err := cmd.ToEntity()
	require.NoError(t, err, "ToEntity should succeed for valid approved review")
	require.NotNil(t, entity)

	assert.Equal(t, "anno-abc", entity.Parent.ID, "Parent.ID (AnnotationID) should be set")
	assert.Equal(t, "reviewer-xyz", entity.ReviewerID)
	assert.Equal(t, "reviewer-xyz", entity.CreatorID, "CreatorID should equal ReviewerID")
	assert.Equal(t, "approved", string(entity.Status))
	assert.Equal(t, &comment, entity.Comments)
	assert.Nil(t, entity.ModifiedPolygon)
	assert.Nil(t, entity.ModifiedValue)
}

func TestCreateAnnotationReviewCommand_ToEntity_MapsPolygon(t *testing.T) {
	cmd := command.CreateAnnotationReviewCommand{
		AnnotationID:  "anno-abc",
		ReviewerID:    "reviewer-xyz",
		Status:        "modified",
		ModifiedValue: float64(3.14),
		ModifiedPolygon: &[]command.CommandPoint{
			{X: 10, Y: 20},
			{X: 30, Y: 40},
			{X: 50, Y: 60},
		},
	}

	entity, err := cmd.ToEntity()
	require.NoError(t, err)
	require.NotNil(t, entity.ModifiedPolygon)

	pts := *entity.ModifiedPolygon
	assert.Len(t, pts, 3)
	assert.Equal(t, 10.0, pts[0].X)
	assert.Equal(t, 20.0, pts[0].Y)
	assert.Equal(t, 50.0, pts[2].X)
}

func TestCreateAnnotationReviewCommand_ToEntity_InvalidCommand_ReturnsError(t *testing.T) {
	cmd := command.CreateAnnotationReviewCommand{
		// missing required fields
	}

	entity, err := cmd.ToEntity()
	assert.Error(t, err)
	assert.Nil(t, entity)
}
