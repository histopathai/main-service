package usecase_test

import (
	"context"
	stderrors "errors"
	"strings"
	"testing"

	appusecase "github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recheckStore struct {
	requests map[string]port.RecheckRequest
}

func (s *recheckStore) List(_ context.Context, status string) ([]port.RecheckRequest, error) {
	var out []port.RecheckRequest
	for _, r := range s.requests {
		if status == "" || r.Status == status {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *recheckStore) Update(_ context.Context, id string,
	change func(*port.RecheckRequest) (*port.RecheckRequest, error)) (*port.RecheckRequest, error) {
	var current *port.RecheckRequest
	if r, ok := s.requests[id]; ok {
		current = &r
	}
	next, err := change(current)
	if err != nil {
		return nil, err
	}
	s.requests[id] = *next
	return next, nil
}

func (s *recheckStore) Delete(_ context.Context, id string) error {
	delete(s.requests, id)
	return nil
}

func (s *recheckStore) UpdateMany(_ context.Context, ids []string,
	change func(string, *port.RecheckRequest) (*port.RecheckRequest, error)) error {
	for _, id := range ids {
		var current *port.RecheckRequest
		if r, ok := s.requests[id]; ok {
			current = &r
		}
		next, err := change(id, current)
		if err != nil {
			return err
		}
		if next == nil {
			delete(s.requests, id)
		} else {
			s.requests[id] = *next
		}
	}
	return nil
}

// filterValue is the value of the filter on field, as the use case sets it.
func filterValue(spec query.Specification, field string) interface{} {
	for _, f := range spec.Filters {
		if f.Field == field {
			return f.Value
		}
	}
	return nil
}

type recheckImages map[string]*model.Image

func (m recheckImages) Read(_ context.Context, id string) (*model.Image, error) {
	if img, ok := m[id]; ok {
		return img, nil
	}
	return nil, errors.NewNotFoundError("image not found")
}

func (m recheckImages) Find(_ context.Context, spec query.Specification) (*query.Result[*model.Image], error) {
	ws := filterValue(spec, "WsID")
	var out []*model.Image
	for _, img := range m {
		if img.WsID == ws && !img.Deleted {
			out = append(out, img)
		}
	}
	return &query.Result[*model.Image]{Data: out}, nil
}

type recheckPatients map[string]*model.Patient

func (m recheckPatients) Read(_ context.Context, id string) (*model.Patient, error) {
	if p, ok := m[id]; ok {
		return p, nil
	}
	return nil, errors.NewNotFoundError("patient not found")
}

func (m recheckPatients) Find(_ context.Context, spec query.Specification) (*query.Result[*model.Patient], error) {
	ws := filterValue(spec, "ParentID")
	var out []*model.Patient
	for _, p := range m {
		if p.Parent.ID == ws {
			out = append(out, p)
		}
	}
	return &query.Result[*model.Patient]{Data: out}, nil
}

func newRecheck() (*appusecase.RecheckUseCase, *recheckStore) {
	store := &recheckStore{requests: map[string]port.RecheckRequest{}}
	images := recheckImages{
		"img1":  {Entity: vobj.Entity{ID: "img1", Name: "24.jpg", Parent: vobj.ParentRef{ID: "p1"}}, WsID: "ws1"},
		"gone":  {Entity: vobj.Entity{ID: "gone", Name: "x.jpg", Deleted: true}, WsID: "ws1"},
		"img2":  {Entity: vobj.Entity{ID: "img2", Name: "60.jpg", Parent: vobj.ParentRef{ID: "p2"}}, WsID: "ws1"},
		"other": {Entity: vobj.Entity{ID: "other", Name: "y.jpg"}, WsID: "ws2"},
	}
	patients := recheckPatients{
		"p1": {Entity: vobj.Entity{ID: "p1", Name: "24", Parent: vobj.ParentRef{ID: "ws1"}}},
		"p2": {Entity: vobj.Entity{ID: "p2", Name: "60", Parent: vobj.ParentRef{ID: "ws1"}}},
	}
	return appusecase.NewRecheckUseCase(store, images, patients), store
}

func isType(err error, t errors.ErrorType) bool {
	var appErr *errors.Err
	return stderrors.As(err, &appErr) && appErr.Type == t
}

func TestRecheckRequestCopiesTheImage(t *testing.T) {
	uc, _ := newRecheck()
	r, err := uc.Request(context.Background(), "img1", "admin1", port.RecheckReasonSubtype, "  ")
	require.NoError(t, err)
	assert.Equal(t, "24.jpg", r.ImageName)
	assert.Equal(t, "p1", r.PatientID)
	assert.Equal(t, "24", r.PatientName)
	assert.Equal(t, "ws1", r.WsID)
	assert.Equal(t, port.RecheckStatusOpen, r.Status)
	require.Len(t, r.Reasons, 1)
	assert.Equal(t, "", r.Reasons[0].Note)
	assert.Equal(t, "admin1", r.Reasons[0].RequestedBy)
}

func TestRecheckSameReasonReplacesItsNote(t *testing.T) {
	uc, _ := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "a", port.RecheckReasonPolygon, "first")
	require.NoError(t, err)
	_, err = uc.Request(ctx, "img1", "a", port.RecheckReasonSubtype, "")
	require.NoError(t, err)
	r, err := uc.Request(ctx, "img1", "a", port.RecheckReasonPolygon, "second")
	require.NoError(t, err)
	require.Len(t, r.Reasons, 2)
	assert.Equal(t, port.RecheckReasonPolygon, r.Reasons[0].Code)
	assert.Equal(t, "second", r.Reasons[0].Note)
}

func TestRecheckSubtypeMissingIsAReason(t *testing.T) {
	uc, _ := newRecheck()
	r, err := uc.Request(context.Background(), "img1", "a", port.RecheckReasonSubtypeMissing, "")
	require.NoError(t, err)
	assert.Equal(t, port.RecheckReasonSubtypeMissing, r.Reasons[0].Code)
}

func TestRecheckOtherReasonsAddUp(t *testing.T) {
	uc, _ := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "a", port.RecheckReasonOther, "one")
	require.NoError(t, err)
	r, err := uc.Request(ctx, "img1", "a", port.RecheckReasonOther, "two")
	require.NoError(t, err)
	assert.Len(t, r.Reasons, 2)
}

func TestRecheckRefusesBadReasons(t *testing.T) {
	uc, store := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "a", "colour", "")
	assert.True(t, isType(err, errors.ErrorTypeValidation))
	_, err = uc.Request(ctx, "img1", "a", port.RecheckReasonOther, " ")
	assert.True(t, isType(err, errors.ErrorTypeValidation), "other needs a note")
	_, err = uc.Request(ctx, "img1", "a", port.RecheckReasonSubtype, strings.Repeat("ç", port.RecheckNoteMaxLen+1))
	assert.True(t, isType(err, errors.ErrorTypeValidation))
	_, err = uc.Request(ctx, "missing", "a", port.RecheckReasonSubtype, "")
	assert.True(t, isType(err, errors.ErrorTypeNotFound))
	_, err = uc.Request(ctx, "gone", "a", port.RecheckReasonSubtype, "")
	assert.True(t, isType(err, errors.ErrorTypeNotFound), "deleted image")
	assert.Empty(t, store.requests)
}

func TestRecheckDoneAndReopen(t *testing.T) {
	uc, _ := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonSubtype, "")
	require.NoError(t, err)

	r, err := uc.SetDone(ctx, "img1", "expert", true, port.RecheckOutcomeCorrected, "")
	require.NoError(t, err)
	assert.Equal(t, port.RecheckStatusDone, r.Status)
	assert.Equal(t, "expert", r.CompletedBy)
	require.NotNil(t, r.CompletedAt)

	open, err := uc.List(ctx, port.RecheckStatusOpen)
	require.NoError(t, err)
	assert.Empty(t, open)

	r, err = uc.SetDone(ctx, "img1", "expert", false, "", "")
	require.NoError(t, err)
	assert.Equal(t, port.RecheckStatusOpen, r.Status)
	assert.Nil(t, r.CompletedAt)
	assert.Equal(t, "", r.CompletedBy)
}

func TestRecheckNewReasonReopensADoneRequest(t *testing.T) {
	uc, _ := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonSubtype, "")
	require.NoError(t, err)
	_, err = uc.SetDone(ctx, "img1", "expert", true, port.RecheckOutcomeCorrected, "")
	require.NoError(t, err)
	r, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonPolygonMissing, "")
	require.NoError(t, err)
	assert.Equal(t, port.RecheckStatusOpen, r.Status)
	assert.Nil(t, r.CompletedAt)
}

func TestRecheckSetDoneWithoutRequest(t *testing.T) {
	uc, _ := newRecheck()
	_, err := uc.SetDone(context.Background(), "img1", "expert", true, port.RecheckOutcomeCorrected, "")
	assert.True(t, isType(err, errors.ErrorTypeNotFound))
}

func TestRecheckListValidatesStatusAndSorts(t *testing.T) {
	uc, store := newRecheck()
	ctx := context.Background()
	_, err := uc.List(ctx, "closed")
	assert.True(t, isType(err, errors.ErrorTypeValidation))

	store.requests["b"] = port.RecheckRequest{ImageID: "b", WsID: "ws2", ImageName: "1.jpg", Status: port.RecheckStatusOpen}
	store.requests["c"] = port.RecheckRequest{ImageID: "c", WsID: "ws1", ImageName: "9.jpg", Status: port.RecheckStatusOpen}
	store.requests["a"] = port.RecheckRequest{ImageID: "a", WsID: "ws1", ImageName: "10.jpg", Status: port.RecheckStatusDone}
	all, err := uc.List(ctx, "")
	require.NoError(t, err)
	ids := []string{all[0].ImageID, all[1].ImageID, all[2].ImageID}
	assert.Equal(t, []string{"a", "c", "b"}, ids)
}

func TestRecheckCancel(t *testing.T) {
	uc, store := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonSubtype, "")
	require.NoError(t, err)
	require.NoError(t, uc.Cancel(ctx, "img1"))
	assert.Empty(t, store.requests)
}

func TestRecheckWorkspaceSendsEveryLiveImage(t *testing.T) {
	uc, store := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonSubtype, "")
	require.NoError(t, err)
	_, err = uc.SetDone(ctx, "img1", "expert", true, port.RecheckOutcomeCorrected, "")
	require.NoError(t, err)

	n, err := uc.RequestWorkspace(ctx, "ws1", "admin", "  Yeni yüklendi, gözden geçirilmeli  ")
	require.NoError(t, err)
	assert.Equal(t, 2, n, "img1 and img2; the deleted image and ws2 are left out")
	require.Contains(t, store.requests, "img2")
	assert.NotContains(t, store.requests, "gone")
	assert.NotContains(t, store.requests, "other")

	img2 := store.requests["img2"]
	assert.Equal(t, "60", img2.PatientName)
	assert.Equal(t, "ws1", img2.WsID)
	require.Len(t, img2.Reasons, 1)
	assert.Equal(t, port.RecheckReasonDataset, img2.Reasons[0].Code)
	assert.Equal(t, "Yeni yüklendi, gözden geçirilmeli", img2.Reasons[0].Note)

	img1 := store.requests["img1"]
	assert.Equal(t, port.RecheckStatusOpen, img1.Status, "a new send reopens")
	assert.Len(t, img1.Reasons, 2, "the image's own reason stays")
}

func TestRecheckWorkspaceNeedsANoteAndImages(t *testing.T) {
	uc, store := newRecheck()
	ctx := context.Background()
	_, err := uc.RequestWorkspace(ctx, "ws1", "admin", " ")
	assert.True(t, isType(err, errors.ErrorTypeValidation))
	_, err = uc.RequestWorkspace(ctx, "empty", "admin", "why")
	assert.True(t, isType(err, errors.ErrorTypeNotFound))
	assert.Empty(t, store.requests)
}

func TestRecheckWorkspaceWithdrawKeepsOtherReasons(t *testing.T) {
	uc, store := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonPolygon, "")
	require.NoError(t, err)
	_, err = uc.RequestWorkspace(ctx, "ws1", "admin", "why")
	require.NoError(t, err)
	_, err = uc.Request(ctx, "other", "admin", port.RecheckReasonSubtype, "")
	require.NoError(t, err)

	n, err := uc.WithdrawWorkspace(ctx, "ws1")
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.NotContains(t, store.requests, "img2", "only the dataset reason: removed")
	require.Contains(t, store.requests, "img1")
	assert.Equal(t, []string{port.RecheckReasonPolygon}, codes(store.requests["img1"]))
	assert.Contains(t, store.requests, "other", "another workspace is not touched")
}

func TestRecheckSingleImageCannotUseTheDatasetReasonWithoutNote(t *testing.T) {
	uc, _ := newRecheck()
	_, err := uc.Request(context.Background(), "img1", "admin", port.RecheckReasonDataset, "")
	assert.True(t, isType(err, errors.ErrorTypeValidation))
}

func codes(r port.RecheckRequest) []string {
	out := make([]string, len(r.Reasons))
	for i, reason := range r.Reasons {
		out[i] = reason.Code
	}
	return out
}

func TestRecheckDoneKeepsTheExpertsAnswer(t *testing.T) {
	uc, _ := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonSubtype, "")
	require.NoError(t, err)

	r, err := uc.SetDone(ctx, "img1", "expert", true, port.RecheckOutcomeNoChange, "  IDC ile uyumlu  ")
	require.NoError(t, err)
	assert.Equal(t, port.RecheckOutcomeNoChange, r.Outcome)
	assert.Equal(t, "IDC ile uyumlu", r.CompletionNote)

	r, err = uc.SetDone(ctx, "img1", "expert", false, "", "")
	require.NoError(t, err)
	assert.Empty(t, r.Outcome, "reopening clears the answer")
	assert.Empty(t, r.CompletionNote)

	_, err = uc.SetDone(ctx, "img1", "expert", true, port.RecheckOutcomeCorrected, "")
	require.NoError(t, err)
	r, err = uc.Request(ctx, "img1", "admin", port.RecheckReasonPolygon, "")
	require.NoError(t, err)
	assert.Empty(t, r.Outcome, "a new reason clears the answer")
}

func TestRecheckDoneNeedsAnOutcomeAndSometimesANote(t *testing.T) {
	uc, _ := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonSubtype, "")
	require.NoError(t, err)
	for _, tc := range []struct{ outcome, note string }{
		{"", ""},
		{"maybe", "x"},
		{port.RecheckOutcomeNoChange, " "},
		{port.RecheckOutcomeUndecided, ""},
		{port.RecheckOutcomeCorrected, strings.Repeat("a", port.RecheckNoteMaxLen+1)},
	} {
		_, err := uc.SetDone(ctx, "img1", "expert", true, tc.outcome, tc.note)
		assert.True(t, isType(err, errors.ErrorTypeValidation), "%q %q", tc.outcome, tc.note)
	}
	r, err := uc.SetDone(ctx, "img1", "expert", true, port.RecheckOutcomeUndecided, "E-cadherin gerekli")
	require.NoError(t, err)
	assert.Equal(t, port.RecheckStatusDone, r.Status)
}

func TestRecheckUnsuitableIsAnOutcome(t *testing.T) {
	uc, _ := newRecheck()
	ctx := context.Background()
	_, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonSubtypeMissing, "")
	require.NoError(t, err)
	r, err := uc.SetDone(ctx, "img1", "expert", true, port.RecheckOutcomeUnsuitable, "")
	require.NoError(t, err, "no note needed")
	assert.Equal(t, port.RecheckOutcomeUnsuitable, r.Outcome)
	assert.Equal(t, port.RecheckStatusDone, r.Status)
}
