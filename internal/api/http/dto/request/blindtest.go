package request

import "time"

// BlindTestAnswerRequest is one answer of a blind test.
type BlindTestAnswerRequest struct {
	Label string `json:"label" binding:"required,oneof=real synthetic" example:"synthetic"`
}

// BlindTestNoteRequest is a participant's note on one image; empty removes it.
type BlindTestNoteRequest struct {
	Note string `json:"note" example:"Çekirdek kromatini fazla düzgün, hücre sınırları bulanık."`
}

// CreateBlindTestInviteRequest makes a shared invitation link to a blind test.
type CreateBlindTestInviteRequest struct {
	MaxParticipants int `json:"max_participants" binding:"required,min=1,max=500" example:"10"`
	// ExpiresAt stops new people joining after it (optional).
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

// UpdateBlindTestInviteRequest opens / closes a link or changes its cap; omitted fields stay.
type UpdateBlindTestInviteRequest struct {
	Active          *bool `json:"active,omitempty"`
	MaxParticipants *int  `json:"max_participants,omitempty" example:"12"`
}

// BlindTestJoinRequest: a person joining through an invitation link.
// The name is kept in capitals with Turkish letters in ASCII ("AYSE YILMAZ").
type BlindTestJoinRequest struct {
	Name    string `json:"name" example:"Ayşe Yılmaz"`
	PIN     string `json:"pin" example:"4821"`
	Consent bool   `json:"consent"`
}

// BlindTestResumeRequest: coming back with the name and PIN given on joining.
type BlindTestResumeRequest struct {
	Name string `json:"name" example:"Ayşe Yılmaz"`
	PIN  string `json:"pin" example:"4821"`
}
