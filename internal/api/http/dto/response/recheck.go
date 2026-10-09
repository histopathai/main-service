package response

import (
	"time"

	"github.com/histopathai/main-service/internal/port"
)

type RecheckReasonResponse struct {
	Code        string    `json:"code"         example:"subtype"`
	Note        string    `json:"note"`
	RequestedBy string    `json:"requested_by"`
	RequestedAt time.Time `json:"requested_at"`
}

type RecheckResponse struct {
	ImageID     string                  `json:"image_id"`
	ImageName   string                  `json:"image_name"   example:"24.jpg"`
	PatientID   string                  `json:"patient_id"`
	PatientName string                  `json:"patient_name" example:"24"`
	WsID        string                  `json:"ws_id"`
	Status      string                  `json:"status"       example:"open"`
	Reasons     []RecheckReasonResponse `json:"reasons"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	CompletedBy string                  `json:"completed_by,omitempty"`
	CompletedAt *time.Time              `json:"completed_at"`
}

func NewRecheckResponse(r *port.RecheckRequest) RecheckResponse {
	reasons := make([]RecheckReasonResponse, len(r.Reasons))
	for i, reason := range r.Reasons {
		reasons[i] = RecheckReasonResponse{Code: reason.Code, Note: reason.Note, RequestedBy: reason.RequestedBy,
			RequestedAt: reason.RequestedAt}
	}
	return RecheckResponse{ImageID: r.ImageID, ImageName: r.ImageName, PatientID: r.PatientID,
		PatientName: r.PatientName, WsID: r.WsID, Status: r.Status, Reasons: reasons, CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt, CompletedBy: r.CompletedBy, CompletedAt: r.CompletedAt}
}

func NewRecheckResponses(list []port.RecheckRequest) []RecheckResponse {
	out := make([]RecheckResponse, len(list))
	for i := range list {
		out[i] = NewRecheckResponse(&list[i])
	}
	return out
}
