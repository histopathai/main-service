package usecase

import (
	"context"
	"hash/fnv"
	"io"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

// BlindTestImagePrefix is where the images of a set live in the processed bucket:
// blind-tests/{set_id}/{image_id}.png
const BlindTestImagePrefix = "blind-tests"

type BlindTestUseCase struct {
	store   port.BlindTestStore
	storage port.Storage
	guests  port.BlindTestGuestLister
	now     func() time.Time
}

func NewBlindTestUseCase(store port.BlindTestStore, storage port.Storage) *BlindTestUseCase {
	return &BlindTestUseCase{store: store, storage: storage, now: time.Now}
}

func (uc *BlindTestUseCase) List(ctx context.Context, userID string) ([]port.BlindTestSummary, error) {
	sets, err := uc.store.ListSets(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(sets, func(i, j int) bool { return sets[i].Name < sets[j].Name })
	out := make([]port.BlindTestSummary, 0, len(sets))
	for _, set := range sets {
		resp, err := uc.store.GetResponse(ctx, set.ID, userID)
		if err != nil {
			return nil, err
		}
		summary := port.BlindTestSummary{Set: withoutImages(set), Total: len(set.ImageIDs)}
		if resp != nil {
			summary.Answered = answeredIn(set, resp)
			summary.Completed = resp.CompletedAt != nil
		}
		out = append(out, summary)
	}
	return out, nil
}

// Get returns the set with its images in the user's own order: every
// participant sees a different order, and the same one each time they return.
func (uc *BlindTestUseCase) Get(ctx context.Context, setID, userID string) (*port.BlindTestView, error) {
	set, err := uc.activeSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	resp, err := uc.store.GetResponse(ctx, setID, userID)
	if err != nil {
		return nil, err
	}
	set.ImageIDs = ParticipantOrder(set.ImageIDs, setID, userID)
	return &port.BlindTestView{Set: *set, Response: resp}, nil
}

// Answer records or changes the user's answer for one image. A completed test
// is locked.
func (uc *BlindTestUseCase) Answer(ctx context.Context, setID, userID, userRole, imageID, label string) (*port.BlindTestResponse, error) {
	if label != port.BlindTestLabelReal && label != port.BlindTestLabelSynthetic {
		return nil, errors.NewValidationError("label must be real or synthetic", map[string]interface{}{"label": label})
	}
	set, err := uc.activeSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if !contains(set.ImageIDs, imageID) {
		return nil, errors.NewNotFoundError("image is not part of this blind test")
	}
	return uc.store.UpdateResponse(ctx, setID, userID, func(cur *port.BlindTestResponse) (*port.BlindTestResponse, error) {
		now := uc.now()
		cur = startResponse(cur, setID, userID, userRole, now)
		if cur.CompletedAt != nil {
			return nil, errors.NewConflictError("this blind test is already completed", nil)
		}
		cur.Answers[imageID] = port.BlindTestAnswer{Label: label, AnsweredAt: now}
		cur.UpdatedAt = now
		return cur, nil
	})
}

// Note writes the user's note on one image of the set; empty text removes it.
// Allowed after completion too: only the answers are locked.
func (uc *BlindTestUseCase) Note(ctx context.Context, setID, userID, userRole, imageID, text string) (*port.BlindTestResponse, error) {
	text = strings.TrimSpace(text)
	if utf8.RuneCountInString(text) > port.BlindTestNoteMaxLen {
		return nil, errors.NewValidationError("note is too long",
			map[string]interface{}{"max_length": port.BlindTestNoteMaxLen})
	}
	set, err := uc.activeSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if !contains(set.ImageIDs, imageID) {
		return nil, errors.NewNotFoundError("image is not part of this blind test")
	}
	return uc.store.UpdateResponse(ctx, setID, userID, func(cur *port.BlindTestResponse) (*port.BlindTestResponse, error) {
		now := uc.now()
		cur = startResponse(cur, setID, userID, userRole, now)
		if text == "" {
			delete(cur.Notes, imageID)
		} else {
			cur.Notes[imageID] = port.BlindTestNote{Text: text, UpdatedAt: now}
		}
		cur.UpdatedAt = now
		return cur, nil
	})
}

// startResponse returns the user's response, a new one if there is none, with its maps ready.
func startResponse(cur *port.BlindTestResponse, setID, userID, userRole string, now time.Time) *port.BlindTestResponse {
	if cur == nil {
		cur = &port.BlindTestResponse{SetID: setID, UserID: userID, UserRole: userRole, StartedAt: now}
	}
	if cur.Answers == nil {
		cur.Answers = map[string]port.BlindTestAnswer{}
	}
	if cur.Notes == nil {
		cur.Notes = map[string]port.BlindTestNote{}
	}
	return cur
}

// Complete locks the user's answers. Every image must be answered first.
// Completing again changes nothing.
func (uc *BlindTestUseCase) Complete(ctx context.Context, setID, userID string) (*port.BlindTestResponse, error) {
	set, err := uc.activeSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	return uc.store.UpdateResponse(ctx, setID, userID, func(cur *port.BlindTestResponse) (*port.BlindTestResponse, error) {
		answered := 0
		if cur != nil {
			if cur.CompletedAt != nil {
				return cur, nil
			}
			answered = answeredIn(*set, cur)
		}
		if answered < len(set.ImageIDs) {
			return nil, errors.NewValidationError("every image must be answered before completing",
				map[string]interface{}{"answered": answered, "total": len(set.ImageIDs)})
		}
		now := uc.now()
		cur.CompletedAt, cur.UpdatedAt = &now, now
		return cur, nil
	})
}

// OpenImage serves an image of an active set. The id must belong to the set,
// so no other object of the bucket can be reached through this route.
func (uc *BlindTestUseCase) OpenImage(ctx context.Context, setID, imageID string) (io.ReadCloser, error) {
	set, err := uc.activeSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if !contains(set.ImageIDs, imageID) {
		return nil, errors.NewNotFoundError("image is not part of this blind test")
	}
	return uc.storage.Get(ctx, model.Content{Path: BlindTestImagePrefix + "/" + setID + "/" + imageID + ".png"})
}

// WithGuests lets the results name the people who joined through invitation links.
func (uc *BlindTestUseCase) WithGuests(guests port.BlindTestGuestLister) *BlindTestUseCase {
	uc.guests = guests
	return uc
}

// Results is for admins: the answer key, every response and the scores.
func (uc *BlindTestUseCase) Results(ctx context.Context, setID string) (*port.BlindTestResults, error) {
	set, err := uc.store.GetSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, errors.NewNotFoundError("blind test not found")
	}
	key, err := uc.store.GetKey(ctx, setID)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, errors.NewNotFoundError("answer key of the blind test not found")
	}
	responses, err := uc.store.ListResponses(ctx, setID)
	if err != nil {
		return nil, err
	}
	res := BuildBlindTestResults(*set, *key, responses)
	if uc.guests != nil {
		guests, err := uc.guests.ListGuests(ctx, setID)
		if err != nil {
			return nil, err
		}
		byUser := map[string]port.BlindTestGuest{}
		for _, g := range guests {
			byUser[BlindTestGuestUserID(g)] = g
		}
		for i, u := range res.Users {
			if g, ok := byUser[u.Response.UserID]; ok {
				res.Users[i].Guest = &port.BlindTestGuestProfile{Name: g.Name, Institution: g.Institution,
					ExperienceYears: g.ExperienceYears}
			}
		}
	}
	return res, nil
}

func (uc *BlindTestUseCase) activeSet(ctx context.Context, setID string) (*port.BlindTestSet, error) {
	set, err := uc.store.GetSet(ctx, setID)
	if err != nil {
		return nil, err
	}
	if set == nil || !set.Active {
		return nil, errors.NewNotFoundError("blind test not found")
	}
	return set, nil
}

// ParticipantOrder shuffles ids with a seed made of the set and the user.
func ParticipantOrder(ids []string, setID, userID string) []string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(setID + "\x00" + userID))
	out := append([]string(nil), ids...)
	rand.New(rand.NewSource(int64(h.Sum64()))).Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// BuildBlindTestResults scores every response against the key. Answers to
// images not in the key are ignored. The pooled score uses completed responses.
func BuildBlindTestResults(set port.BlindTestSet, key port.BlindTestKey, responses []port.BlindTestResponse) *port.BlindTestResults {
	res := &port.BlindTestResults{Set: set, RunID: key.RunID}
	votes := map[string][2]int{}
	notes := map[string][]port.BlindTestImageNote{}
	var pooled []port.BlindTestAnswerPair
	sort.Slice(responses, func(i, j int) bool { return responses[i].StartedAt.Before(responses[j].StartedAt) })
	for _, r := range responses {
		var pairs []port.BlindTestAnswerPair
		for id, a := range r.Answers {
			item, ok := key.Items[id]
			if !ok {
				continue
			}
			pairs = append(pairs, port.BlindTestAnswerPair{Truth: item.Label, Answer: a.Label})
			if r.CompletedAt != nil {
				v := votes[id]
				if a.Label == port.BlindTestLabelReal {
					v[0]++
				} else {
					v[1]++
				}
				votes[id] = v
			}
		}
		for id, n := range r.Notes {
			notes[id] = append(notes[id], port.BlindTestImageNote{UserID: r.UserID, Answer: r.Answers[id].Label,
				Completed: r.CompletedAt != nil, Text: n.Text, UpdatedAt: n.UpdatedAt})
		}
		res.Users = append(res.Users, port.BlindTestUserResult{Response: r, Score: Score(pairs)})
		if r.CompletedAt != nil {
			pooled = append(pooled, pairs...)
		}
	}
	res.Pooled = Score(pooled)
	for _, id := range set.ImageIDs {
		item := key.Items[id]
		v := votes[id]
		res.Images = append(res.Images, port.BlindTestImageResult{ImageID: id, Label: item.Label, Source: item.Source,
			VotedReal: v[0], VotedSynthetic: v[1], Notes: notes[id]})
	}
	return res
}

// Score counts a list of (truth, answer) pairs.
func Score(pairs []port.BlindTestAnswerPair) port.BlindTestScore {
	var s port.BlindTestScore
	for _, p := range pairs {
		realTruth, realAnswer := p.Truth == port.BlindTestLabelReal, p.Answer == port.BlindTestLabelReal
		switch {
		case realTruth && realAnswer:
			s.Confusion.RealAsReal++
		case realTruth:
			s.Confusion.RealAsSynthetic++
		case realAnswer:
			s.Confusion.SyntheticAsReal++
		default:
			s.Confusion.SyntheticAsSynthetic++
		}
	}
	c := s.Confusion
	s.Answered = c.RealAsReal + c.RealAsSynthetic + c.SyntheticAsReal + c.SyntheticAsSynthetic
	s.Correct = c.RealAsReal + c.SyntheticAsSynthetic
	s.PValue = 1
	if s.Answered > 0 {
		s.Accuracy = float64(s.Correct) / float64(s.Answered)
		s.PValue = BinomialTwoSided(s.Correct, s.Answered)
	}
	if synthetic := c.SyntheticAsReal + c.SyntheticAsSynthetic; synthetic > 0 {
		s.SyntheticCalledReal = float64(c.SyntheticAsReal) / float64(synthetic)
	}
	return s
}

// BinomialTwoSided is the exact two-sided p-value of k successes in n trials
// against p = 0.5: the probability of an outcome at least as far from n/2.
func BinomialTwoSided(k, n int) float64 {
	if n == 0 {
		return 1
	}
	dist := math.Abs(float64(k) - float64(n)/2)
	logHalf := float64(n) * math.Log(0.5)
	lgN, _ := math.Lgamma(float64(n + 1))
	p := 0.0
	for i := 0; i <= n; i++ {
		if math.Abs(float64(i)-float64(n)/2) >= dist-1e-9 {
			lgI, _ := math.Lgamma(float64(i + 1))
			lgR, _ := math.Lgamma(float64(n - i + 1))
			p += math.Exp(lgN - lgI - lgR + logHalf)
		}
	}
	return math.Min(1, p)
}

func answeredIn(set port.BlindTestSet, r *port.BlindTestResponse) int {
	n := 0
	for _, id := range set.ImageIDs {
		if _, ok := r.Answers[id]; ok {
			n++
		}
	}
	return n
}

func withoutImages(set port.BlindTestSet) port.BlindTestSet {
	set.ImageIDs = nil
	return set
}

func contains(ids []string, id string) bool {
	for _, x := range ids {
		if x == id {
			return true
		}
	}
	return false
}
