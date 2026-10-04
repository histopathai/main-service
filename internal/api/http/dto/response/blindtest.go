package response

import (
	"time"

	"github.com/histopathai/main-service/internal/port"
)

type BlindTestSummaryResponse struct {
	ID          string `json:"id"          example:"set-a"`
	Name        string `json:"name"        example:"Set A"`
	Description string `json:"description"`
	Total       int    `json:"total"       example:"200"`
	Answered    int    `json:"answered"    example:"37"`
	Completed   bool   `json:"completed"`
}

func NewBlindTestSummaryResponses(list []port.BlindTestSummary) []BlindTestSummaryResponse {
	out := make([]BlindTestSummaryResponse, len(list))
	for i, s := range list {
		out[i] = BlindTestSummaryResponse{ID: s.Set.ID, Name: s.Set.Name, Description: s.Set.Description,
			Total: s.Total, Answered: s.Answered, Completed: s.Completed}
	}
	return out
}

// BlindTestProgressResponse is the participant's own answers: image id -> real | synthetic.
type BlindTestProgressResponse struct {
	Answers     map[string]string `json:"answers"`
	CompletedAt *time.Time        `json:"completed_at"`
}

func NewBlindTestProgressResponse(r *port.BlindTestResponse) BlindTestProgressResponse {
	out := BlindTestProgressResponse{Answers: map[string]string{}}
	if r != nil {
		for id, a := range r.Answers {
			out.Answers[id] = a.Label
		}
		out.CompletedAt = r.CompletedAt
	}
	return out
}

// BlindTestResponse is a set as the participant takes it; image_ids are in their own order.
type BlindTestResponse struct {
	ID          string   `json:"id"          example:"set-a"`
	Name        string   `json:"name"        example:"Set A"`
	Description string   `json:"description"`
	ImageIDs    []string `json:"image_ids"`
	BlindTestProgressResponse
}

func NewBlindTestResponse(v *port.BlindTestView) BlindTestResponse {
	return BlindTestResponse{ID: v.Set.ID, Name: v.Set.Name, Description: v.Set.Description, ImageIDs: v.Set.ImageIDs,
		BlindTestProgressResponse: NewBlindTestProgressResponse(v.Response)}
}

type BlindTestConfusionResponse struct {
	RealAsReal           int `json:"real_as_real"`
	RealAsSynthetic      int `json:"real_as_synthetic"`
	SyntheticAsReal      int `json:"synthetic_as_real"`
	SyntheticAsSynthetic int `json:"synthetic_as_synthetic"`
}

// BlindTestScoreResponse: accuracy 0.5 is chance; p_value is the two-sided exact binomial test against 0.5.
type BlindTestScoreResponse struct {
	Answered            int                        `json:"answered"`
	Correct             int                        `json:"correct"`
	Accuracy            float64                    `json:"accuracy"`
	PValue              float64                    `json:"p_value"`
	SyntheticCalledReal float64                    `json:"synthetic_called_real"`
	Confusion           BlindTestConfusionResponse `json:"confusion"`
}

func newBlindTestScoreResponse(s port.BlindTestScore) BlindTestScoreResponse {
	c := s.Confusion
	return BlindTestScoreResponse{Answered: s.Answered, Correct: s.Correct, Accuracy: s.Accuracy, PValue: s.PValue,
		SyntheticCalledReal: s.SyntheticCalledReal, Confusion: BlindTestConfusionResponse{RealAsReal: c.RealAsReal,
			RealAsSynthetic: c.RealAsSynthetic, SyntheticAsReal: c.SyntheticAsReal, SyntheticAsSynthetic: c.SyntheticAsSynthetic}}
}

type BlindTestUserResultResponse struct {
	UserID      string                 `json:"user_id"`
	UserRole    string                 `json:"user_role"`
	StartedAt   time.Time              `json:"started_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	Score       BlindTestScoreResponse `json:"score"`
}

type BlindTestImageResultResponse struct {
	ImageID        string            `json:"image_id"`
	Label          string            `json:"label" example:"synthetic"`
	Source         map[string]string `json:"source"`
	VotedReal      int               `json:"voted_real"`
	VotedSynthetic int               `json:"voted_synthetic"`
}

// BlindTestResultsResponse is for admins only: it carries the answer key.
type BlindTestResultsResponse struct {
	ID     string                         `json:"id"`
	Name   string                         `json:"name"`
	RunID  string                         `json:"run_id"`
	Active bool                           `json:"active"`
	Pooled BlindTestScoreResponse         `json:"pooled"`
	Users  []BlindTestUserResultResponse  `json:"users"`
	Images []BlindTestImageResultResponse `json:"images"`
}

func NewBlindTestResultsResponse(r *port.BlindTestResults) BlindTestResultsResponse {
	out := BlindTestResultsResponse{ID: r.Set.ID, Name: r.Set.Name, RunID: r.RunID, Active: r.Set.Active,
		Pooled: newBlindTestScoreResponse(r.Pooled), Users: []BlindTestUserResultResponse{},
		Images: make([]BlindTestImageResultResponse, len(r.Images))}
	for _, u := range r.Users {
		out.Users = append(out.Users, BlindTestUserResultResponse{UserID: u.Response.UserID, UserRole: u.Response.UserRole,
			StartedAt: u.Response.StartedAt, UpdatedAt: u.Response.UpdatedAt, CompletedAt: u.Response.CompletedAt,
			Score: newBlindTestScoreResponse(u.Score)})
	}
	for i, img := range r.Images {
		out.Images[i] = BlindTestImageResultResponse{ImageID: img.ImageID, Label: img.Label, Source: img.Source,
			VotedReal: img.VotedReal, VotedSynthetic: img.VotedSynthetic}
	}
	return out
}
