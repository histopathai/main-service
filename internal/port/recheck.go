package port

import (
	"context"
	"time"
)

// Recheck ("Ek Kontrol"): an admin sends an image back to the expert who
// labelled it, with the reason. The image and its annotations stay where they
// are; a request only points at the image, so the Ek Kontrol tab is a filtered
// view of data already uploaded.
const (
	RecheckStatusOpen = "open"
	RecheckStatusDone = "done"
)

// Recheck reasons. The tab shows each as a short sentence; "other" carries
// its sentence in the note.
const (
	RecheckReasonSubtype            = "subtype"              // Alt tip yeniden incelenmeli
	RecheckReasonPolygon            = "polygon"              // Poligon yeniden incelenmeli
	RecheckReasonPolygonMissing     = "polygon_missing"      // Poligon eksik
	RecheckReasonGlobalLabelMissing = "global_label_missing" // Global etiket eksik
	RecheckReasonOther              = "other"
)

func IsRecheckReason(code string) bool {
	switch code {
	case RecheckReasonSubtype, RecheckReasonPolygon, RecheckReasonPolygonMissing,
		RecheckReasonGlobalLabelMissing, RecheckReasonOther:
		return true
	}
	return false
}

// RecheckNoteMaxLen is the longest note accepted, in characters.
const RecheckNoteMaxLen = 500

type RecheckReason struct {
	Code        string
	Note        string
	RequestedBy string
	RequestedAt time.Time
}

// RecheckRequest is the request on one image; its ID is the image ID.
// Workspace, patient and image names are copied in when it is made, so the tab
// can list it without loading every image.
type RecheckRequest struct {
	ImageID     string
	ImageName   string
	PatientID   string
	PatientName string
	WsID        string
	Reasons     []RecheckReason
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedBy string
	CompletedAt *time.Time
}

type RecheckStore interface {
	// List returns the requests with the given status, or all of them when status is empty.
	List(ctx context.Context, status string) ([]RecheckRequest, error)
	// Update reads the request of the image (nil if none), lets change modify
	// it and writes it back, atomically.
	Update(ctx context.Context, imageID string,
		change func(current *RecheckRequest) (*RecheckRequest, error)) (*RecheckRequest, error)
	Delete(ctx context.Context, imageID string) error
}

type RecheckUseCase interface {
	List(ctx context.Context, status string) ([]RecheckRequest, error)
	// Request adds a reason to the image's request, making the request if
	// there is none and reopening it if it was done. The same reason given
	// again replaces its note.
	Request(ctx context.Context, imageID, userID, code, note string) (*RecheckRequest, error)
	// SetDone marks the request done, or open again.
	SetDone(ctx context.Context, imageID, userID string, done bool) (*RecheckRequest, error)
	Cancel(ctx context.Context, imageID string) error
}
