package usecase

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type recheckImageReader interface {
	Read(ctx context.Context, id string) (*model.Image, error)
	Find(ctx context.Context, spec query.Specification) (*query.Result[*model.Image], error)
}

type recheckPatientReader interface {
	Read(ctx context.Context, id string) (*model.Patient, error)
	Find(ctx context.Context, spec query.Specification) (*query.Result[*model.Patient], error)
}

type RecheckUseCase struct {
	store    port.RecheckStore
	images   recheckImageReader
	patients recheckPatientReader
	now      func() time.Time
}

func NewRecheckUseCase(store port.RecheckStore, images recheckImageReader, patients recheckPatientReader) *RecheckUseCase {
	return &RecheckUseCase{store: store, images: images, patients: patients, now: time.Now}
}

// List returns the requests by workspace, then image name.
func (uc *RecheckUseCase) List(ctx context.Context, status string) ([]port.RecheckRequest, error) {
	if status != "" && status != port.RecheckStatusOpen && status != port.RecheckStatusDone {
		return nil, errors.NewValidationError("status must be open or done", map[string]interface{}{"status": status})
	}
	list, err := uc.store.List(ctx, status)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].WsID != list[j].WsID {
			return list[i].WsID < list[j].WsID
		}
		return list[i].ImageName < list[j].ImageName
	})
	return list, nil
}

func validateRecheckReason(code, note string) error {
	if !port.IsRecheckReason(code) {
		return errors.NewValidationError("unknown recheck reason", map[string]interface{}{"code": code})
	}
	if (code == port.RecheckReasonOther || code == port.RecheckReasonDataset) && note == "" {
		return errors.NewValidationError("this reason needs a note", map[string]interface{}{"code": code})
	}
	if utf8.RuneCountInString(note) > port.RecheckNoteMaxLen {
		return errors.NewValidationError("note is too long", map[string]interface{}{"max": port.RecheckNoteMaxLen})
	}
	return nil
}

// withReason adds the reason to the request (made if nil), replacing the same
// reason (except "other", which adds up), and reopens it.
func withReason(current *port.RecheckRequest, imageID string, reason port.RecheckReason) *port.RecheckRequest {
	next := current
	if next == nil {
		next = &port.RecheckRequest{ImageID: imageID, CreatedAt: reason.RequestedAt}
	}
	replaced := false
	for i := range next.Reasons {
		if next.Reasons[i].Code == reason.Code && reason.Code != port.RecheckReasonOther {
			next.Reasons[i], replaced = reason, true
		}
	}
	if !replaced {
		next.Reasons = append(next.Reasons, reason)
	}
	reopen(next)
	next.UpdatedAt = reason.RequestedAt
	return next
}

// reopen clears what finishing set.
func reopen(r *port.RecheckRequest) {
	r.Status, r.CompletedBy, r.CompletedAt, r.Outcome, r.CompletionNote = port.RecheckStatusOpen, "", nil, "", ""
}

func (uc *RecheckUseCase) Request(ctx context.Context, imageID, userID, code, note string) (*port.RecheckRequest, error) {
	note = strings.TrimSpace(note)
	if err := validateRecheckReason(code, note); err != nil {
		return nil, err
	}
	image, err := uc.images.Read(ctx, imageID)
	if err != nil {
		return nil, err
	}
	if image.Deleted {
		return nil, errors.NewNotFoundError("image not found")
	}
	patientName := ""
	if image.Parent.ID != "" {
		if patient, err := uc.patients.Read(ctx, image.Parent.ID); err == nil && patient != nil {
			patientName = patient.Name
		}
	}

	reason := port.RecheckReason{Code: code, Note: note, RequestedBy: userID, RequestedAt: uc.now()}
	return uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		next := withReason(current, imageID, reason)
		next.ImageName, next.PatientID, next.PatientName, next.WsID = image.Name, image.Parent.ID, patientName, image.WsID
		return next, nil
	})
}

// recheckPage is how many entities one Find reads while walking a workspace.
const recheckPage = 1000

func (uc *RecheckUseCase) workspaceImages(ctx context.Context, wsID string) ([]*model.Image, error) {
	var all []*model.Image
	for offset := 0; ; offset += recheckPage {
		spec := query.NewBuilder().
			Where(fields.ImageWsID.DomainName(), query.OpEqual, wsID).
			Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false).
			Build()
		spec.Pagination = &query.Pagination{Limit: recheckPage, Offset: offset}
		page, err := uc.images.Find(ctx, spec)
		if err != nil {
			return nil, err
		}
		all = append(all, page.Data...)
		if !page.HasMore {
			return all, nil
		}
	}
}

func (uc *RecheckUseCase) patientNames(ctx context.Context, wsID string) (map[string]string, error) {
	names := map[string]string{}
	for offset := 0; ; offset += recheckPage {
		spec := query.NewBuilder().
			Where(fields.EntityParentID.DomainName(), query.OpEqual, wsID).
			Build()
		spec.Pagination = &query.Pagination{Limit: recheckPage, Offset: offset}
		page, err := uc.patients.Find(ctx, spec)
		if err != nil {
			return nil, err
		}
		for _, p := range page.Data {
			names[p.ID] = p.Name
		}
		if !page.HasMore {
			return names, nil
		}
	}
}

func (uc *RecheckUseCase) RequestWorkspace(ctx context.Context, wsID, userID, note string) (int, error) {
	note = strings.TrimSpace(note)
	if err := validateRecheckReason(port.RecheckReasonDataset, note); err != nil {
		return 0, err
	}
	images, err := uc.workspaceImages(ctx, wsID)
	if err != nil {
		return 0, err
	}
	if len(images) == 0 {
		return 0, errors.NewNotFoundError("the workspace has no images")
	}
	patients, err := uc.patientNames(ctx, wsID)
	if err != nil {
		return 0, err
	}
	byID := make(map[string]*model.Image, len(images))
	ids := make([]string, len(images))
	for i, img := range images {
		byID[img.ID], ids[i] = img, img.ID
	}
	reason := port.RecheckReason{Code: port.RecheckReasonDataset, Note: note, RequestedBy: userID, RequestedAt: uc.now()}
	err = uc.store.UpdateMany(ctx, ids, func(id string, current *port.RecheckRequest) (*port.RecheckRequest, error) {
		img := byID[id]
		next := withReason(current, id, reason)
		next.ImageName, next.PatientID, next.PatientName, next.WsID = img.Name, img.Parent.ID, patients[img.Parent.ID], wsID
		return next, nil
	})
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

func (uc *RecheckUseCase) WithdrawWorkspace(ctx context.Context, wsID string) (int, error) {
	all, err := uc.store.List(ctx, "")
	if err != nil {
		return 0, err
	}
	var ids []string
	for _, r := range all {
		if r.WsID != wsID {
			continue
		}
		for _, reason := range r.Reasons {
			if reason.Code == port.RecheckReasonDataset {
				ids = append(ids, r.ImageID)
				break
			}
		}
	}
	now := uc.now()
	err = uc.store.UpdateMany(ctx, ids, func(_ string, current *port.RecheckRequest) (*port.RecheckRequest, error) {
		if current == nil {
			return nil, nil
		}
		kept := current.Reasons[:0]
		for _, reason := range current.Reasons {
			if reason.Code != port.RecheckReasonDataset {
				kept = append(kept, reason)
			}
		}
		if len(kept) == 0 {
			return nil, nil
		}
		current.Reasons, current.UpdatedAt = kept, now
		return current, nil
	})
	if err != nil {
		return 0, err
	}
	return len(ids), nil
}

func (uc *RecheckUseCase) SetDone(ctx context.Context, imageID, userID string, done bool, outcome, note string) (*port.RecheckRequest, error) {
	note = strings.TrimSpace(note)
	if done {
		if !port.IsRecheckOutcome(outcome) {
			return nil, errors.NewValidationError("outcome must be corrected, no_change, undecided or unsuitable",
				map[string]interface{}{"outcome": outcome})
		}
		if outcome != port.RecheckOutcomeCorrected && note == "" {
			return nil, errors.NewValidationError("this outcome needs a note saying why", map[string]interface{}{"outcome": outcome})
		}
		if utf8.RuneCountInString(note) > port.RecheckNoteMaxLen {
			return nil, errors.NewValidationError("note is too long", map[string]interface{}{"max": port.RecheckNoteMaxLen})
		}
	}
	now := uc.now()
	return uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		if current == nil {
			return nil, errors.NewNotFoundError("no recheck request for this image")
		}
		if done {
			current.Status, current.CompletedBy, current.CompletedAt = port.RecheckStatusDone, userID, &now
			current.Outcome, current.CompletionNote = outcome, note
		} else {
			reopen(current)
		}
		current.UpdatedAt = now
		return current, nil
	})
}

func (uc *RecheckUseCase) Cancel(ctx context.Context, imageID string) error {
	return uc.store.Delete(ctx, imageID)
}
