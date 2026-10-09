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
	RecheckReasonSubtypeMissing     = "subtype_missing"      // Alt tip eksik
	RecheckReasonOther              = "other"
	// RecheckReasonDataset is set on every image of a workspace sent as a
	// whole; its note says why and is required.
	RecheckReasonDataset = "dataset"
)

func IsRecheckReason(code string) bool {
	switch code {
	case RecheckReasonSubtype, RecheckReasonPolygon, RecheckReasonPolygonMissing,
		RecheckReasonGlobalLabelMissing, RecheckReasonSubtypeMissing, RecheckReasonOther, RecheckReasonDataset:
		return true
	}
	return false
}

// What the expert concluded when finishing a request. "no_change" and
// "undecided" need a note saying why; for the others it is optional.
const (
	RecheckOutcomeCorrected = "corrected" // Etiketler düzeltildi
	RecheckOutcomeNoChange  = "no_change" // Değişiklik gerekmedi, mevcut etiket doğru
	RecheckOutcomeUndecided = "undecided" // Karar verilemedi
	// RecheckOutcomeUnsuitable: the image is not fit for the study (too little
	// tumour, poor section, another lesion); its labels are left as they are.
	// Saying so is enough: the note is optional.
	RecheckOutcomeUnsuitable = "unsuitable" // Çalışmaya uygun değil
)

func IsRecheckOutcome(code string) bool {
	switch code {
	case RecheckOutcomeCorrected, RecheckOutcomeNoChange, RecheckOutcomeUndecided, RecheckOutcomeUnsuitable:
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
	// Outcome and CompletionNote are the expert's answer, set when done.
	Outcome        string
	CompletionNote string
}

type RecheckStore interface {
	// List returns the requests with the given status, or all of them when status is empty.
	List(ctx context.Context, status string) ([]RecheckRequest, error)
	// Update reads the request of the image (nil if none), lets change modify
	// it and writes it back, atomically.
	Update(ctx context.Context, imageID string,
		change func(current *RecheckRequest) (*RecheckRequest, error)) (*RecheckRequest, error)
	Delete(ctx context.Context, imageID string) error
	// UpdateMany does what Update does for many images at once, in batches
	// (not one transaction); change returning nil deletes the request.
	UpdateMany(ctx context.Context, imageIDs []string,
		change func(imageID string, current *RecheckRequest) (*RecheckRequest, error)) error
}

type RecheckUseCase interface {
	List(ctx context.Context, status string) ([]RecheckRequest, error)
	// Request adds a reason to the image's request, making the request if
	// there is none and reopening it if it was done. The same reason given
	// again replaces its note.
	Request(ctx context.Context, imageID, userID, code, note string) (*RecheckRequest, error)
	// SetDone marks the request done with the expert's outcome and note, or
	// open again (outcome and note are cleared).
	SetDone(ctx context.Context, imageID, userID string, done bool, outcome, note string) (*RecheckRequest, error)
	Cancel(ctx context.Context, imageID string) error
	// RequestWorkspace sends every image of the workspace with the reason
	// "dataset" and the note; returns how many images it reached.
	RequestWorkspace(ctx context.Context, wsID, userID, note string) (int, error)
	// WithdrawWorkspace takes the reason "dataset" off the workspace's images,
	// removing requests left without a reason; other reasons stay.
	WithdrawWorkspace(ctx context.Context, wsID string) (int, error)
}
