package request

// BlindTestAnswerRequest is one answer of a blind test.
type BlindTestAnswerRequest struct {
	Label string `json:"label" binding:"required,oneof=real synthetic" example:"synthetic"`
}
