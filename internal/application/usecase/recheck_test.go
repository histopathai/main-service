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

type recheckImages map[string]*model.Image

func (m recheckImages) Read(_ context.Context, id string) (*model.Image, error) {
	if img, ok := m[id]; ok {
		return img, nil
	}
	return nil, errors.NewNotFoundError("image not found")
}

type recheckPatients map[string]*model.Patient

func (m recheckPatients) Read(_ context.Context, id string) (*model.Patient, error) {
	if p, ok := m[id]; ok {
		return p, nil
	}
	return nil, errors.NewNotFoundError("patient not found")
}

func newRecheck() (*appusecase.RecheckUseCase, *recheckStore) {
	store := &recheckStore{requests: map[string]port.RecheckRequest{}}
	images := recheckImages{
		"img1": {Entity: vobj.Entity{ID: "img1", Name: "24.jpg", Parent: vobj.ParentRef{ID: "p1"}}, WsID: "ws1"},
		"gone": {Entity: vobj.Entity{ID: "gone", Name: "x.jpg", Deleted: true}, WsID: "ws1"},
	}
	patients := recheckPatients{"p1": {Entity: vobj.Entity{ID: "p1", Name: "24"}}}
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

	r, err := uc.SetDone(ctx, "img1", "expert", true)
	require.NoError(t, err)
	assert.Equal(t, port.RecheckStatusDone, r.Status)
	assert.Equal(t, "expert", r.CompletedBy)
	require.NotNil(t, r.CompletedAt)

	open, err := uc.List(ctx, port.RecheckStatusOpen)
	require.NoError(t, err)
	assert.Empty(t, open)

	r, err = uc.SetDone(ctx, "img1", "expert", false)
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
	_, err = uc.SetDone(ctx, "img1", "expert", true)
	require.NoError(t, err)
	r, err := uc.Request(ctx, "img1", "admin", port.RecheckReasonPolygonMissing, "")
	require.NoError(t, err)
	assert.Equal(t, port.RecheckStatusOpen, r.Status)
	assert.Nil(t, r.CompletedAt)
}

func TestRecheckSetDoneWithoutRequest(t *testing.T) {
	uc, _ := newRecheck()
	_, err := uc.SetDone(context.Background(), "img1", "expert", true)
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
