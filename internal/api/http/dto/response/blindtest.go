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

// BlindTestProgressResponse is the participant's own answers (image id -> real | synthetic) and notes.
type BlindTestProgressResponse struct {
	Answers     map[string]string `json:"answers"`
	Notes       map[string]string `json:"notes"`
	CompletedAt *time.Time        `json:"completed_at"`
}

func NewBlindTestProgressResponse(r *port.BlindTestResponse) BlindTestProgressResponse {
	out := BlindTestProgressResponse{Answers: map[string]string{}, Notes: map[string]string{}}
	if r != nil {
		for id, a := range r.Answers {
			out.Answers[id] = a.Label
		}
		for id, n := range r.Notes {
			out.Notes[id] = n.Text
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
	UserID   string `json:"user_id"`
	UserRole string `json:"user_role"`
	// Guest is set for people who joined through an invitation link (no platform account).
	Guest *BlindTestGuestProfileResponse `json:"guest"`
	// Order is the order this participant was shown the images in; Answers their answer per image.
	Order       []string               `json:"order"`
	Answers     map[string]string      `json:"answers"`
	StartedAt   time.Time              `json:"started_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at"`
	Score       BlindTestScoreResponse `json:"score"`
}

type BlindTestImageNoteResponse struct {
	UserID    string    `json:"user_id"`
	Answer    string    `json:"answer" example:"synthetic"`
	Completed bool      `json:"completed"`
	Note      string    `json:"note"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BlindTestImageResultResponse struct {
	ImageID        string                       `json:"image_id"`
	Label          string                       `json:"label" example:"synthetic"`
	Source         map[string]string            `json:"source"`
	VotedReal      int                          `json:"voted_real"`
	VotedSynthetic int                          `json:"voted_synthetic"`
	Notes          []BlindTestImageNoteResponse `json:"notes"`
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
		var guest *BlindTestGuestProfileResponse
		if u.Guest != nil {
			guest = &BlindTestGuestProfileResponse{Name: u.Guest.Name}
		}
		answers := map[string]string{}
		for id, a := range u.Response.Answers {
			answers[id] = a.Label
		}
		out.Users = append(out.Users, BlindTestUserResultResponse{UserID: u.Response.UserID, UserRole: u.Response.UserRole, Guest: guest,
			Order: u.Order, Answers: answers,
			StartedAt: u.Response.StartedAt, UpdatedAt: u.Response.UpdatedAt, CompletedAt: u.Response.CompletedAt,
			Score: newBlindTestScoreResponse(u.Score)})
	}
	for i, img := range r.Images {
		notes := make([]BlindTestImageNoteResponse, len(img.Notes))
		for j, n := range img.Notes {
			notes[j] = BlindTestImageNoteResponse{UserID: n.UserID, Answer: n.Answer, Completed: n.Completed,
				Note: n.Text, UpdatedAt: n.UpdatedAt}
		}
		out.Images[i] = BlindTestImageResultResponse{ImageID: img.ImageID, Label: img.Label, Source: img.Source,
			VotedReal: img.VotedReal, VotedSynthetic: img.VotedSynthetic, Notes: notes}
	}
	return out
}

// BlindTestInviteResponse is an invitation link as admins see it; the link is
// /kor-test/katil/{token} on the web app.
type BlindTestInviteResponse struct {
	ID              string     `json:"id"`
	Token           string     `json:"token"`
	SetID           string     `json:"set_id"`
	MaxParticipants int        `json:"max_participants" example:"10"`
	Participants    int        `json:"participants" example:"7"`
	ExpiresAt       *time.Time `json:"expires_at"`
	Active          bool       `json:"active"`
	CreatedAt       time.Time  `json:"created_at"`
}

func NewBlindTestInviteResponse(inv *port.BlindTestInvite) BlindTestInviteResponse {
	return BlindTestInviteResponse{ID: inv.ID, Token: inv.Token, SetID: inv.SetID, MaxParticipants: inv.MaxParticipants,
		Participants: inv.Participants, ExpiresAt: inv.ExpiresAt, Active: inv.Active, CreatedAt: inv.CreatedAt}
}

func NewBlindTestInviteResponses(list []port.BlindTestInvite) []BlindTestInviteResponse {
	out := make([]BlindTestInviteResponse, len(list))
	for i := range list {
		out[i] = NewBlindTestInviteResponse(&list[i])
	}
	return out
}

// BlindTestInviteInfoResponse is what anyone holding the link sees before joining.
type BlindTestInviteInfoResponse struct {
	SetName      string `json:"set_name"     example:"Set A"`
	Images       int    `json:"images"       example:"200"`
	Participants int    `json:"participants" example:"7"`
	Max          int    `json:"max"          example:"10"`
	Joinable     bool   `json:"joinable"`
	Closed       bool   `json:"closed"`
	Expired      bool   `json:"expired"`
}

func NewBlindTestInviteInfoResponse(i *port.BlindTestInviteInfo) BlindTestInviteInfoResponse {
	return BlindTestInviteInfoResponse{SetName: i.SetName, Images: i.Images,
		Participants: i.Participants, Max: i.Max, Joinable: i.Joinable, Closed: i.Closed, Expired: i.Expired}
}

// BlindTestGuestSessionResponse: session_token goes in the X-Guest-Session header.
type BlindTestGuestSessionResponse struct {
	SessionToken string `json:"session_token"`
	Name         string `json:"name" example:"Ayşe Yılmaz"`
}

type BlindTestGuestProfileResponse struct {
	Name string `json:"name" example:"AYSE YILMAZ"`
}
