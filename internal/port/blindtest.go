package port

import (
	"context"
	"io"
	"time"
)

// Blind test: pathologists are shown a set of image patches, half of them real
// and half produced by a generative model, and mark each one "real" or
// "synthetic". The answer key never leaves the server except to admins.
const (
	BlindTestLabelReal      = "real"
	BlindTestLabelSynthetic = "synthetic"
)

// BlindTestSet is what a participant may know about a set: its images, but not
// which of them are real nor which model produced the others. Name is neutral
// ("Set A"); the run is recorded only in the answer key.
type BlindTestSet struct {
	ID          string
	Name        string
	Description string
	ImageIDs    []string
	Active      bool
	CreatedAt   time.Time
}

// BlindTestKeyItem is the truth about one image of a set.
type BlindTestKeyItem struct {
	Label string
	// Source describes where the image came from (dataset, slide, patch, seed).
	Source map[string]string
}

// BlindTestKey is the answer key of a set. Admins only.
type BlindTestKey struct {
	SetID string
	RunID string
	Items map[string]BlindTestKeyItem
}

type BlindTestAnswer struct {
	Label      string
	AnsweredAt time.Time
}

// BlindTestResponse is one user's answers to one set.
type BlindTestResponse struct {
	SetID       string
	UserID      string
	UserRole    string
	Answers     map[string]BlindTestAnswer
	StartedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// BlindTestStore keeps sets, their answer keys and the users' responses.
type BlindTestStore interface {
	ListSets(ctx context.Context) ([]BlindTestSet, error)
	// GetSet and GetKey return nil, nil when the set does not exist.
	GetSet(ctx context.Context, setID string) (*BlindTestSet, error)
	GetKey(ctx context.Context, setID string) (*BlindTestKey, error)
	// GetResponse returns nil, nil when the user has not answered yet.
	GetResponse(ctx context.Context, setID, userID string) (*BlindTestResponse, error)
	ListResponses(ctx context.Context, setID string) ([]BlindTestResponse, error)
	// UpdateResponse reads the user's response (nil if none), lets change
	// modify it and writes it back, atomically.
	UpdateResponse(ctx context.Context, setID, userID string,
		change func(current *BlindTestResponse) (*BlindTestResponse, error)) (*BlindTestResponse, error)
}

// BlindTestSummary is a set as listed to a participant, with their progress.
type BlindTestSummary struct {
	Set       BlindTestSet
	Total     int
	Answered  int
	Completed bool
}

// BlindTestAnswerPair is one answer next to the truth.
type BlindTestAnswerPair struct {
	Truth  string
	Answer string
}

// BlindTestView is a set as a participant takes it: the images in their own
// order and their own answers so far.
type BlindTestView struct {
	Set      BlindTestSet
	Response *BlindTestResponse
}

// BlindTestConfusion counts answers by truth (rows) and answer (columns).
type BlindTestConfusion struct {
	RealAsReal           int
	RealAsSynthetic      int
	SyntheticAsReal      int
	SyntheticAsSynthetic int
}

// BlindTestScore is how well answers told real from synthetic. Accuracy 0.5 is
// chance: the images could not be told apart. PValue is the two-sided exact
// binomial test of Correct out of Answered against 0.5.
type BlindTestScore struct {
	Answered  int
	Correct   int
	Accuracy  float64
	PValue    float64
	Confusion BlindTestConfusion
	// SyntheticCalledReal is the share of synthetic images taken for real.
	SyntheticCalledReal float64
}

type BlindTestUserResult struct {
	Response BlindTestResponse
	Score    BlindTestScore
}

type BlindTestImageResult struct {
	ImageID        string
	Label          string
	Source         map[string]string
	VotedReal      int
	VotedSynthetic int
}

// BlindTestResults is the admin's view: the key, every response and the scores.
// Pooled is computed over completed responses only.
type BlindTestResults struct {
	Set    BlindTestSet
	RunID  string
	Users  []BlindTestUserResult
	Pooled BlindTestScore
	Images []BlindTestImageResult
}

type BlindTestUseCase interface {
	List(ctx context.Context, userID string) ([]BlindTestSummary, error)
	Get(ctx context.Context, setID, userID string) (*BlindTestView, error)
	Answer(ctx context.Context, setID, userID, userRole, imageID, label string) (*BlindTestResponse, error)
	Complete(ctx context.Context, setID, userID string) (*BlindTestResponse, error)
	OpenImage(ctx context.Context, setID, imageID string) (io.ReadCloser, error)
	Results(ctx context.Context, setID string) (*BlindTestResults, error)
}
