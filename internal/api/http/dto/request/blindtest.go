package request

// BlindTestAnswerRequest is one answer of a blind test.
type BlindTestAnswerRequest struct {
	Label string `json:"label" binding:"required,oneof=real synthetic" example:"synthetic"`
}

// BlindTestNoteRequest is a participant's note on one image; empty removes it.
type BlindTestNoteRequest struct {
	Note string `json:"note" example:"Çekirdek kromatini fazla düzgün, hücre sınırları bulanık."`
}
