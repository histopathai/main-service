package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/histopathai/main-service/internal/application/command"
	appusecase "github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fakes
// ─────────────────────────────────────────────────────────────────────────────

type tissueImageRepo struct {
	port.ImageRepository
	images map[string]*model.Image
}

func (r *tissueImageRepo) Read(_ context.Context, id string) (*model.Image, error) {
	if img, ok := r.images[id]; ok {
		return img, nil
	}
	return nil, errors.NewNotFoundError("document not found")
}

type tissueMaskRepo struct {
	port.TissueMaskRepository
	masks  map[string]*model.TissueMask
	writes int
}

func (r *tissueMaskRepo) Read(_ context.Context, id string) (*model.TissueMask, error) {
	if m, ok := r.masks[id]; ok {
		copy := *m
		return &copy, nil
	}
	return nil, errors.NewNotFoundError("document not found")
}

func (r *tissueMaskRepo) Create(_ context.Context, m *model.TissueMask) (*model.TissueMask, error) {
	r.writes++
	m.CreatedAt, m.UpdatedAt = time.Unix(1, 0), time.Unix(1, 0)
	copy := *m
	r.masks[m.ID] = &copy
	return m, nil
}

func (r *tissueMaskRepo) Update(_ context.Context, id string, updates map[string]interface{}) error {
	r.writes++
	stored := r.masks[id]
	for k, v := range updates {
		switch k {
		case "Mask":
			next := *v.(*model.TissueMask)
			next.CreatedAt = stored.CreatedAt
			*stored = next
		case fields.TissueMaskStatus.DomainName():
			stored.Status = v.(vobj.TissueMaskStatus)
		case fields.TissueMaskApprovedBy.DomainName():
			stored.ApprovedBy = v.(*string)
		case fields.TissueMaskApprovedAt.DomainName():
			stored.ApprovedAt = v.(*time.Time)
		case fields.EntityIsDeleted.DomainName():
			stored.Deleted = v.(bool)
		}
	}
	return nil
}

type tissueUoW struct {
	images *tissueImageRepo
	masks  *tissueMaskRepo
}

func (u *tissueUoW) WithTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }
func (u *tissueUoW) GetWorkspaceRepo() port.WorkspaceRepository                       { return nil }
func (u *tissueUoW) GetPatientRepo() port.PatientRepository                           { return nil }
func (u *tissueUoW) GetImageRepo() port.ImageRepository                               { return u.images }
func (u *tissueUoW) GetAnnotationRepo() port.AnnotationRepository                     { return nil }
func (u *tissueUoW) GetAnnotationReviewRepo() port.AnnotationReviewRepository         { return nil }
func (u *tissueUoW) GetAnnotationTypeRepo() port.AnnotationTypeRepository             { return nil }
func (u *tissueUoW) GetContentRepo() port.ContentRepository                           { return nil }
func (u *tissueUoW) GetTissueMaskRepo() port.TissueMaskRepository                     { return u.masks }

func review(user string, expected *int, reason *string) command.ReviewTissueMaskCommand {
	return command.ReviewTissueMaskCommand{ImageID: "img-1", UserID: user, ExpectedRevision: expected, Reason: reason}
}

func ptr[T any](v T) *T { return &v }

func newTissueFixture() (*appusecase.TissueMaskUseCase, *tissueUoW) {
	uow := &tissueUoW{
		images: &tissueImageRepo{images: map[string]*model.Image{
			"img-1": {Entity: vobj.Entity{ID: "img-1", Name: "slide.svs", CreatorID: "owner"}, WsID: "ws-1"},
		}},
		masks: &tissueMaskRepo{masks: map[string]*model.TissueMask{}},
	}
	return appusecase.NewTissueMaskUseCase(uow), uow
}

func validTissueData(ratio float64) command.TissueMaskData {
	return command.TissueMaskData{
		AlgorithmVersion: "tissue-v1",
		Params: vobj.TissueParams{
			Method: "saturation-gray", SaturationThreshold: 0.05, GrayThreshold: 0.92,
			ClosingRadius: 3, MinObjectArea: 500, MinHoleArea: 500, SimplifyTolerance: 1,
		},
		Polygons: []vobj.TissuePolygon{{
			Exterior: []vobj.Point{{X: 10, Y: 10}, {X: 900, Y: 10}, {X: 900, Y: 700}},
			Holes:    [][]vobj.Point{{{X: 100, Y: 100}, {X: 200, Y: 100}, {X: 200, Y: 200}}},
		}},
		PreviewWidth: 2048, PreviewHeight: 1536,
		Level0Width: 40960, Level0Height: 30720,
		DownsampleX: 20, DownsampleY: 20,
		TissueAreaRatio: ratio,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// State machine
// ─────────────────────────────────────────────────────────────────────────────

func TestTissueMask_WorkerCreatesAutoMask(t *testing.T) {
	uc, uow := newTissueFixture()

	applied, err := uc.ApplyWorkerResult(context.Background(), command.ApplyWorkerTissueMaskCommand{
		ImageID: "img-1", TissueMaskData: validTissueData(0.3),
	})
	require.NoError(t, err)
	assert.True(t, applied)

	m := uow.masks.masks["img-1"]
	require.NotNil(t, m)
	assert.Equal(t, vobj.TissueMaskStatusAuto, m.Status)
	assert.Equal(t, "ws-1", m.WsID)
	assert.Equal(t, "owner", m.CreatorID)
	assert.Equal(t, vobj.ParentRef{ID: "img-1", Type: vobj.ParentTypeImage}, m.Parent)
	assert.Equal(t, vobj.EntityTypeTissueMask, m.EntityType)
}

func TestTissueMask_WorkerOverwritesOnlyAutoMasks(t *testing.T) {
	uc, uow := newTissueFixture()
	ctx := context.Background()
	apply := func(ratio float64) bool {
		applied, err := uc.ApplyWorkerResult(ctx, command.ApplyWorkerTissueMaskCommand{ImageID: "img-1", TissueMaskData: validTissueData(ratio)})
		require.NoError(t, err)
		return applied
	}

	require.True(t, apply(0.1))
	require.True(t, apply(0.2), "auto masks are replaced by a newer worker result")
	assert.Equal(t, 0.2, uow.masks.masks["img-1"].TissueAreaRatio)
	assert.Equal(t, time.Unix(1, 0), uow.masks.masks["img-1"].CreatedAt, "created_at survives replacement")

	_, err := uc.Save(ctx, command.SaveTissueMaskCommand{ImageID: "img-1", UserID: "ml-user", TissueMaskData: validTissueData(0.5)})
	require.NoError(t, err)
	assert.False(t, apply(0.9), "edited masks are never overwritten")
	assert.Equal(t, 0.5, uow.masks.masks["img-1"].TissueAreaRatio)

	_, err = uc.Approve(ctx, review("reviewer", nil, nil))
	require.NoError(t, err)
	assert.False(t, apply(0.9), "approved masks are never overwritten")
	assert.Equal(t, vobj.TissueMaskStatusApproved, uow.masks.masks["img-1"].Status)
}

func TestTissueMask_SaveAfterApprovalClearsApproval(t *testing.T) {
	uc, uow := newTissueFixture()
	ctx := context.Background()

	saved, err := uc.Save(ctx, command.SaveTissueMaskCommand{ImageID: "img-1", UserID: "ml-user", TissueMaskData: validTissueData(0.4)})
	require.NoError(t, err)
	assert.Equal(t, vobj.TissueMaskStatusEdited, saved.Status)
	require.NotNil(t, saved.EditedBy)
	assert.Equal(t, "ml-user", *saved.EditedBy)
	assert.Equal(t, "ml-user", uow.masks.masks["img-1"].CreatorID, "a mask first created by a person belongs to them")

	approved, err := uc.Approve(ctx, review("reviewer", nil, nil))
	require.NoError(t, err)
	assert.Equal(t, vobj.TissueMaskStatusApproved, approved.Status)
	require.NotNil(t, uow.masks.masks["img-1"].ApprovedBy)
	assert.Equal(t, "reviewer", *uow.masks.masks["img-1"].ApprovedBy)

	_, err = uc.Save(ctx, command.SaveTissueMaskCommand{ImageID: "img-1", UserID: "ml-user", TissueMaskData: validTissueData(0.45)})
	require.NoError(t, err)
	m := uow.masks.masks["img-1"]
	assert.Equal(t, vobj.TissueMaskStatusEdited, m.Status)
	assert.Nil(t, m.ApprovedBy)
	assert.Nil(t, m.ApprovedAt)
}

func TestTissueMask_ApproveIsIdempotentAndNeedsAMask(t *testing.T) {
	uc, uow := newTissueFixture()
	ctx := context.Background()

	_, err := uc.Approve(ctx, review("reviewer", nil, nil))
	requireErrorType(t, err, errors.ErrorTypeNotFound)

	_, err = uc.ApplyWorkerResult(ctx, command.ApplyWorkerTissueMaskCommand{ImageID: "img-1", TissueMaskData: validTissueData(0.3)})
	require.NoError(t, err)
	_, err = uc.Approve(ctx, review("reviewer", nil, nil))
	require.NoError(t, err, "an auto mask can be approved as is")
	writes := uow.masks.writes
	_, err = uc.Approve(ctx, review("someone-else", nil, nil))
	require.NoError(t, err)
	assert.Equal(t, writes, uow.masks.writes)
	assert.Equal(t, "reviewer", *uow.masks.masks["img-1"].ApprovedBy)
}

func TestTissueMask_SaveRejectsUnknownImage(t *testing.T) {
	uc, _ := newTissueFixture()
	_, err := uc.Save(context.Background(), command.SaveTissueMaskCommand{ImageID: "missing", UserID: "u", TissueMaskData: validTissueData(0.1)})
	requireErrorType(t, err, errors.ErrorTypeNotFound)
}

// ─────────────────────────────────────────────────────────────────────────────
// Validation
// ─────────────────────────────────────────────────────────────────────────────

func TestTissueMask_Validation(t *testing.T) {
	cases := map[string]struct {
		mutate func(*command.TissueMaskData)
		field  string
	}{
		"unknown method":  {func(d *command.TissueMaskData) { d.Params.Method = "magic" }, "params.method"},
		"threshold":       {func(d *command.TissueMaskData) { d.Params.GrayThreshold = 1.5 }, "params.gray_threshold"},
		"closing radius":  {func(d *command.TissueMaskData) { d.Params.ClosingRadius = 65 }, "params.closing_radius"},
		"short ring":      {func(d *command.TissueMaskData) { d.Polygons[0].Exterior = d.Polygons[0].Exterior[:2] }, "polygons[0].exterior"},
		"short hole":      {func(d *command.TissueMaskData) { d.Polygons[0].Holes[0] = nil }, "polygons[0].holes[0]"},
		"out of bounds":   {func(d *command.TissueMaskData) { d.Polygons[0].Exterior[1].X = 50000 }, "polygons[0].exterior"},
		"no preview size": {func(d *command.TissueMaskData) { d.PreviewWidth = 0 }, "preview_size"},
		"bad ratio":       {func(d *command.TissueMaskData) { d.TissueAreaRatio = 2 }, "tissue_area_ratio"},
		"no version":      {func(d *command.TissueMaskData) { d.AlgorithmVersion = "" }, "algorithm_version"},
		"zero downsample": {func(d *command.TissueMaskData) { d.DownsampleY = 0 }, "downsample"},
		"too many points": {func(d *command.TissueMaskData) {
			ring := make([]vobj.Point, vobj.TissueMaxPoints+1)
			for i := range ring {
				ring[i] = vobj.Point{X: float64(i % 100), Y: float64(i / 100)}
			}
			d.Polygons = []vobj.TissuePolygon{{Exterior: ring}}
		}, "polygons"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			uc, uow := newTissueFixture()
			data := validTissueData(0.2)
			data.Polygons = []vobj.TissuePolygon{{
				Exterior: append([]vobj.Point(nil), data.Polygons[0].Exterior...),
				Holes:    [][]vobj.Point{append([]vobj.Point(nil), data.Polygons[0].Holes[0]...)},
			}}
			tc.mutate(&data)

			_, err := uc.Save(context.Background(), command.SaveTissueMaskCommand{ImageID: "img-1", UserID: "u", TissueMaskData: data})
			appErr := requireErrorType(t, err, errors.ErrorTypeValidation)
			assert.Contains(t, appErr.Details, tc.field)
			assert.Zero(t, uow.masks.writes)
		})
	}
}

func requireErrorType(t *testing.T, err error, want errors.ErrorType) *errors.Err {
	t.Helper()
	require.Error(t, err)
	appErr, ok := err.(*errors.Err)
	require.True(t, ok, "expected *errors.Err, got %T: %v", err, err)
	require.Equal(t, want, appErr.Type)
	return appErr
}

// ─────────────────────────────────────────────────────────────────────────────
// Rejection and revisions
// ─────────────────────────────────────────────────────────────────────────────

func TestTissueMask_RejectFlow(t *testing.T) {
	uc, uow := newTissueFixture()
	ctx := context.Background()
	_, err := uc.ApplyWorkerResult(ctx, command.ApplyWorkerTissueMaskCommand{ImageID: "img-1", TissueMaskData: validTissueData(0.3)})
	require.NoError(t, err)
	_, err = uc.Approve(ctx, review("reviewer", nil, nil))
	require.NoError(t, err)

	rejected, err := uc.Reject(ctx, review("reviewer", nil, ptr("  IHC, doku ayrılamıyor ")))
	require.NoError(t, err)
	assert.Equal(t, vobj.TissueMaskStatusRejected, rejected.Status)
	m := uow.masks.masks["img-1"]
	require.NotNil(t, m.RejectReason)
	assert.Equal(t, "IHC, doku ayrılamıyor", *m.RejectReason, "reason is trimmed")
	assert.Nil(t, m.ApprovedBy, "rejecting clears the approval")

	writes := uow.masks.writes
	_, err = uc.Reject(ctx, review("other", nil, ptr("IHC, doku ayrılamıyor")))
	require.NoError(t, err)
	assert.Equal(t, writes, uow.masks.writes, "same rejection is a no-op")

	applied, err := uc.ApplyWorkerResult(ctx, command.ApplyWorkerTissueMaskCommand{ImageID: "img-1", TissueMaskData: validTissueData(0.9)})
	require.NoError(t, err)
	assert.False(t, applied, "the worker never overwrites a rejected mask")

	_, err = uc.Approve(ctx, review("reviewer", nil, nil))
	require.NoError(t, err)
	m = uow.masks.masks["img-1"]
	assert.Equal(t, vobj.TissueMaskStatusApproved, m.Status)
	assert.Nil(t, m.RejectedBy, "approving clears the rejection")
	assert.Nil(t, m.RejectReason)

	_, err = uc.Reject(ctx, review("reviewer", nil, nil))
	require.NoError(t, err)
	_, err = uc.Save(ctx, command.SaveTissueMaskCommand{ImageID: "img-1", UserID: "ml-user", TissueMaskData: validTissueData(0.4)})
	require.NoError(t, err)
	m = uow.masks.masks["img-1"]
	assert.Equal(t, vobj.TissueMaskStatusEdited, m.Status)
	assert.Nil(t, m.RejectedBy, "saving clears the rejection")
}

func TestTissueMask_RejectNeedsAMaskAndShortReason(t *testing.T) {
	uc, _ := newTissueFixture()
	ctx := context.Background()
	_, err := uc.Reject(ctx, review("reviewer", nil, nil))
	requireErrorType(t, err, errors.ErrorTypeNotFound)

	long := make([]byte, vobj.TissueMaxRejectReasonLength+1)
	for i := range long {
		long[i] = 'a'
	}
	_, err = uc.Reject(ctx, review("reviewer", nil, ptr(string(long))))
	requireErrorType(t, err, errors.ErrorTypeValidation)
}

func TestTissueMask_RevisionsRefuseStaleWrites(t *testing.T) {
	uc, uow := newTissueFixture()
	ctx := context.Background()
	save := func(user string, expected *int) error {
		_, err := uc.Save(ctx, command.SaveTissueMaskCommand{ImageID: "img-1", UserID: user, TissueMaskData: validTissueData(0.2), ExpectedRevision: expected})
		return err
	}

	// No mask yet: the client expects revision 0.
	require.NoError(t, save("alice", ptr(0)))
	assert.Equal(t, 1, uow.masks.masks["img-1"].Revision)

	// Alice and Bob both loaded revision 1; Bob saves first.
	require.NoError(t, save("bob", ptr(1)))
	assert.Equal(t, 2, uow.masks.masks["img-1"].Revision)
	err := save("alice", ptr(1))
	conflict := requireErrorType(t, err, errors.ErrorTypeConflict)
	assert.Equal(t, 2, conflict.Details["current_revision"])
	assert.Equal(t, "bob", *uow.masks.masks["img-1"].EditedBy, "the stale write changed nothing")

	// Approve and reject check the revision too.
	_, err = uc.Approve(ctx, review("alice", ptr(1), nil))
	requireErrorType(t, err, errors.ErrorTypeConflict)
	approved, err := uc.Approve(ctx, review("alice", ptr(2), nil))
	require.NoError(t, err)
	assert.Equal(t, 3, approved.Revision)
	_, err = uc.Reject(ctx, review("bob", ptr(2), nil))
	requireErrorType(t, err, errors.ErrorTypeConflict)

	// The worker ignores revisions but still advances them.
	uow.masks.masks["img-1"].Status = vobj.TissueMaskStatusAuto
	applied, err := uc.ApplyWorkerResult(ctx, command.ApplyWorkerTissueMaskCommand{ImageID: "img-1", TissueMaskData: validTissueData(0.1)})
	require.NoError(t, err)
	assert.True(t, applied)
	assert.Equal(t, 4, uow.masks.masks["img-1"].Revision)

	// Omitting the expectation skips the check.
	require.NoError(t, save("carol", nil))
}
