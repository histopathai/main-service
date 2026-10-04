package usecase_test

import (
	"context"
	stderrors "errors"
	"io"
	"strings"
	"testing"
	"time"

	appusecase "github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fakes
// ─────────────────────────────────────────────────────────────────────────────

type blindStore struct {
	sets      map[string]port.BlindTestSet
	keys      map[string]port.BlindTestKey
	responses map[string]*port.BlindTestResponse
}

func newBlindStore() *blindStore {
	return &blindStore{
		sets: map[string]port.BlindTestSet{
			"s1":  {ID: "s1", Name: "Set A", Active: true, ImageIDs: []string{"a", "b", "c", "d"}},
			"off": {ID: "off", Name: "Set B", Active: false, ImageIDs: []string{"x"}},
		},
		keys: map[string]port.BlindTestKey{"s1": {SetID: "s1", RunID: "run-1", Items: map[string]port.BlindTestKeyItem{
			"a": {Label: "real"}, "b": {Label: "real"}, "c": {Label: "synthetic"}, "d": {Label: "synthetic"}}}},
		responses: map[string]*port.BlindTestResponse{},
	}
}

func (s *blindStore) ListSets(context.Context) ([]port.BlindTestSet, error) {
	var out []port.BlindTestSet
	for _, set := range s.sets {
		if set.Active {
			out = append(out, set)
		}
	}
	return out, nil
}

func (s *blindStore) GetSet(_ context.Context, id string) (*port.BlindTestSet, error) {
	if set, ok := s.sets[id]; ok {
		return &set, nil
	}
	return nil, nil
}

func (s *blindStore) GetKey(_ context.Context, id string) (*port.BlindTestKey, error) {
	if key, ok := s.keys[id]; ok {
		return &key, nil
	}
	return nil, nil
}

func (s *blindStore) GetResponse(_ context.Context, setID, userID string) (*port.BlindTestResponse, error) {
	return s.responses[setID+"/"+userID], nil
}

func (s *blindStore) ListResponses(_ context.Context, setID string) ([]port.BlindTestResponse, error) {
	var out []port.BlindTestResponse
	for _, r := range s.responses {
		if r.SetID == setID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (s *blindStore) UpdateResponse(_ context.Context, setID, userID string,
	change func(*port.BlindTestResponse) (*port.BlindTestResponse, error)) (*port.BlindTestResponse, error) {
	var cur *port.BlindTestResponse
	if r := s.responses[setID+"/"+userID]; r != nil {
		copied := *r
		copied.Answers = map[string]port.BlindTestAnswer{}
		for k, v := range r.Answers {
			copied.Answers[k] = v
		}
		cur = &copied
	}
	next, err := change(cur)
	if err != nil {
		return nil, err
	}
	s.responses[setID+"/"+userID] = next
	return next, nil
}

type blindStorage struct {
	port.Storage
	opened []string
}

func (s *blindStorage) Get(_ context.Context, c model.Content) (io.ReadCloser, error) {
	s.opened = append(s.opened, c.Path)
	return io.NopCloser(strings.NewReader("png")), nil
}

func errType(err error) errors.ErrorType {
	var e *errors.Err
	if stderrors.As(err, &e) {
		return e.Type
	}
	return ""
}

func answerAll(t *testing.T, uc *appusecase.BlindTestUseCase, user string, labels map[string]string) {
	for id, label := range labels {
		_, err := uc.Answer(context.Background(), "s1", user, "pathologist", id, label)
		require.NoError(t, err)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestBlindTest_ListHidesImagesAndInactiveSets(t *testing.T) {
	uc := appusecase.NewBlindTestUseCase(newBlindStore(), &blindStorage{})
	answerAll(t, uc, "u1", map[string]string{"a": "real"})

	list, err := uc.List(context.Background(), "u1")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "s1", list[0].Set.ID)
	assert.Empty(t, list[0].Set.ImageIDs)
	assert.Equal(t, 4, list[0].Total)
	assert.Equal(t, 1, list[0].Answered)
	assert.False(t, list[0].Completed)
}

func TestBlindTest_OrderIsPerUserAndStable(t *testing.T) {
	ids := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}
	a1 := appusecase.ParticipantOrder(ids, "s1", "u1")
	assert.Equal(t, a1, appusecase.ParticipantOrder(ids, "s1", "u1"))
	assert.NotEqual(t, a1, appusecase.ParticipantOrder(ids, "s1", "u2"))
	assert.ElementsMatch(t, ids, a1)
	assert.Equal(t, "1", ids[0], "input must not be shuffled in place")
}

func TestBlindTest_AnswerValidation(t *testing.T) {
	uc := appusecase.NewBlindTestUseCase(newBlindStore(), &blindStorage{})
	ctx := context.Background()

	_, err := uc.Answer(ctx, "s1", "u1", "pathologist", "a", "maybe")
	assert.Equal(t, errors.ErrorTypeValidation, errType(err))
	_, err = uc.Answer(ctx, "s1", "u1", "pathologist", "zzz", "real")
	assert.Equal(t, errors.ErrorTypeNotFound, errType(err))
	_, err = uc.Answer(ctx, "off", "u1", "pathologist", "x", "real")
	assert.Equal(t, errors.ErrorTypeNotFound, errType(err), "inactive sets cannot be answered")

	r, err := uc.Answer(ctx, "s1", "u1", "pathologist", "a", "real")
	require.NoError(t, err)
	r, err = uc.Answer(ctx, "s1", "u1", "pathologist", "a", "synthetic")
	require.NoError(t, err)
	assert.Equal(t, "synthetic", r.Answers["a"].Label, "answers can be changed before completing")
	assert.Equal(t, "pathologist", r.UserRole)
}

func TestBlindTest_CompleteNeedsAllAnswersThenLocks(t *testing.T) {
	uc := appusecase.NewBlindTestUseCase(newBlindStore(), &blindStorage{})
	ctx := context.Background()

	_, err := uc.Complete(ctx, "s1", "u1")
	assert.Equal(t, errors.ErrorTypeValidation, errType(err), "nothing answered")
	answerAll(t, uc, "u1", map[string]string{"a": "real", "b": "real", "c": "real"})
	_, err = uc.Complete(ctx, "s1", "u1")
	assert.Equal(t, errors.ErrorTypeValidation, errType(err), "one image left")

	answerAll(t, uc, "u1", map[string]string{"d": "synthetic"})
	r, err := uc.Complete(ctx, "s1", "u1")
	require.NoError(t, err)
	require.NotNil(t, r.CompletedAt)
	first := *r.CompletedAt

	r, err = uc.Complete(ctx, "s1", "u1")
	require.NoError(t, err)
	assert.Equal(t, first, *r.CompletedAt, "completing again changes nothing")
	_, err = uc.Answer(ctx, "s1", "u1", "pathologist", "a", "synthetic")
	assert.Equal(t, errors.ErrorTypeConflict, errType(err), "completed answers are locked")
}

func TestBlindTest_OpenImageOnlyWithinTheSet(t *testing.T) {
	storage := &blindStorage{}
	uc := appusecase.NewBlindTestUseCase(newBlindStore(), storage)
	ctx := context.Background()

	_, err := uc.OpenImage(ctx, "s1", "../../processed/secret")
	assert.Equal(t, errors.ErrorTypeNotFound, errType(err))
	_, err = uc.OpenImage(ctx, "off", "x")
	assert.Equal(t, errors.ErrorTypeNotFound, errType(err))
	rc, err := uc.OpenImage(ctx, "s1", "c")
	require.NoError(t, err)
	_ = rc.Close()
	assert.Equal(t, []string{"blind-tests/s1/c.png"}, storage.opened)
}

func TestBlindTest_ResultsScoreCompletedResponses(t *testing.T) {
	store := newBlindStore()
	uc := appusecase.NewBlindTestUseCase(store, &blindStorage{})
	ctx := context.Background()

	// u1 gets 3 of 4 right and completes; u2 answers one image and stops.
	answerAll(t, uc, "u1", map[string]string{"a": "real", "b": "synthetic", "c": "synthetic", "d": "synthetic"})
	_, err := uc.Complete(ctx, "s1", "u1")
	require.NoError(t, err)
	store.responses["s1/u1"].StartedAt = time.Unix(0, 0)
	answerAll(t, uc, "u2", map[string]string{"c": "real"})

	res, err := uc.Results(ctx, "s1")
	require.NoError(t, err)
	assert.Equal(t, "run-1", res.RunID)
	require.Len(t, res.Users, 2)
	assert.Equal(t, "u1", res.Users[0].Response.UserID)
	assert.Equal(t, 3, res.Users[0].Score.Correct)
	assert.InDelta(t, 0.75, res.Users[0].Score.Accuracy, 1e-9)
	assert.Equal(t, port.BlindTestConfusion{RealAsReal: 1, RealAsSynthetic: 1, SyntheticAsSynthetic: 2},
		res.Users[0].Score.Confusion)
	assert.InDelta(t, 1.0, res.Users[1].Score.SyntheticCalledReal, 1e-9)

	assert.Equal(t, 4, res.Pooled.Answered, "only the completed response is pooled")
	require.Len(t, res.Images, 4)
	assert.Equal(t, port.BlindTestImageResult{ImageID: "c", Label: "synthetic", VotedSynthetic: 1}, res.Images[2],
		"u2's unfinished answer does not vote")

	_, err = uc.Results(ctx, "missing")
	assert.Equal(t, errors.ErrorTypeNotFound, errType(err))
}

func TestBlindTest_BinomialTwoSided(t *testing.T) {
	assert.InDelta(t, 1.0, appusecase.BinomialTwoSided(5, 10), 1e-12)
	assert.InDelta(t, 0.021484375, appusecase.BinomialTwoSided(9, 10), 1e-12) // 2 * (10 + 1) / 1024
	assert.InDelta(t, appusecase.BinomialTwoSided(1, 10), appusecase.BinomialTwoSided(9, 10), 1e-12)
	assert.Less(t, appusecase.BinomialTwoSided(130, 200), 1e-4)
	assert.Equal(t, 1.0, appusecase.BinomialTwoSided(0, 0))
}
