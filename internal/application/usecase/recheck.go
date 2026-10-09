package usecase

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type recheckImageReader interface {
	Read(ctx context.Context, id string) (*model.Image, error)
	Find(ctx context.Context, spec query.Specification) (*query.Result[*model.Image], error)
	// Update sets "Çalışmaya uygun değil" on the image when the expert decides so here.
	Update(ctx context.Context, id string, updates map[string]interface{}) error
}

type recheckPatientReader interface {
	Read(ctx context.Context, id string) (*model.Patient, error)
	Find(ctx context.Context, spec query.Specification) (*query.Result[*model.Patient], error)
}

type recheckAnnotationFinder interface {
	Find(ctx context.Context, spec query.Specification) (*query.Result[*model.Annotation], error)
}

type RecheckUseCase struct {
	store       port.RecheckStore
	images      recheckImageReader
	patients    recheckPatientReader
	annotations recheckAnnotationFinder
	now         func() time.Time
}

func NewRecheckUseCase(store port.RecheckStore, images recheckImageReader, patients recheckPatientReader,
	annotations recheckAnnotationFinder) *RecheckUseCase {
	return &RecheckUseCase{store: store, images: images, patients: patients, annotations: annotations, now: time.Now}
}

// RecheckAutoNote is the completion note of a request finished by Reconcile.
const RecheckAutoNote = "Eksik etiket girildi (otomatik)"

// imageLabels reports whether the image has a global label and a polygon now.
func (uc *RecheckUseCase) imageLabels(ctx context.Context, imageID string) (hasGlobal, hasPolygon bool, err error) {
	for offset := 0; ; offset += recheckPage {
		spec := query.NewBuilder().
			Where(fields.EntityParentID.DomainName(), query.OpEqual, imageID).
			Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false).
			Build()
		spec.Pagination = &query.Pagination{Limit: recheckPage, Offset: offset}
		page, err := uc.annotations.Find(ctx, spec)
		if err != nil {
			return false, false, err
		}
		for _, a := range page.Data {
			if a.IsGlobal {
				hasGlobal = true
			} else if a.Polygon != nil && len(*a.Polygon) > 0 {
				hasPolygon = true
			}
		}
		if !page.HasMore {
			return hasGlobal, hasPolygon, nil
		}
	}
}

func (uc *RecheckUseCase) Reconcile(ctx context.Context, imageID, actorID string) error {
	hasGlobal, hasPolygon, err := uc.imageLabels(ctx, imageID)
	if err != nil {
		return err
	}
	now := uc.now()
	_, err = uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		if current == nil {
			return nil, nil
		}
		changed, allSettled := false, len(current.Reasons) > 0
		for i := range current.Reasons {
			r := &current.Reasons[i]
			if !port.IsMissingLabelReason(r.Code) {
				allSettled = false
				continue
			}
			missing := !hasGlobal
			if r.Code == port.RecheckReasonPolygonMissing {
				missing = !hasPolygon
			}
			switch {
			case !missing && r.ResolvedAt == nil:
				r.ResolvedAt, changed = &now, true
			case missing && r.ResolvedAt != nil:
				r.ResolvedAt, changed = nil, true
			}
			if missing {
				allSettled = false
			}
		}
		switch {
		case current.Status == port.RecheckStatusOpen && allSettled:
			current.Status, current.CompletedBy, current.CompletedAt = port.RecheckStatusDone, actorID, &now
			current.Outcome, current.CompletionNote, current.AutoCompleted =
				port.RecheckOutcomeCorrected, RecheckAutoNote, true
			changed = true
		case current.Status == port.RecheckStatusDone && current.AutoCompleted && !allSettled:
			reopen(current)
			changed = true
		}
		if !changed {
			return nil, nil
		}
		current.UpdatedAt = now
		return current, nil
	})
	return err
}

// List returns the requests by workspace, then image name.
func (uc *RecheckUseCase) List(ctx context.Context, viewer port.RecheckViewer, status string) ([]port.RecheckRequest, error) {
	if status != "" && status != port.RecheckStatusOpen && status != port.RecheckStatusDone {
		return nil, errors.NewValidationError("status must be open or done", map[string]interface{}{"status": status})
	}
	all, err := uc.store.List(ctx, status)
	if err != nil {
		return nil, err
	}
	list := all[:0]
	for _, r := range all {
		if viewer.Sees(r) {
			list = append(list, r)
		}
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

func validateAssignee(assigneeID string) error {
	if strings.TrimSpace(assigneeID) == "" {
		return errors.NewValidationError("choose who the request goes to (assignee_id)", nil)
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
	r.AutoCompleted = false
}

func (uc *RecheckUseCase) Request(ctx context.Context, imageID, userID, code, note, assigneeID string) (*port.RecheckRequest, error) {
	note, assigneeID = strings.TrimSpace(note), strings.TrimSpace(assigneeID)
	if err := validateRecheckReason(code, note); err != nil {
		return nil, err
	}
	if err := validateAssignee(assigneeID); err != nil {
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
	r, err := uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		next := withReason(current, imageID, reason)
		next.ImageName, next.PatientID, next.PatientName, next.WsID = image.Name, image.Parent.ID, patientName, image.WsID
		next.AssigneeID = assigneeID
		return next, nil
	})
	if err != nil || !port.IsMissingLabelReason(code) {
		return r, err
	}
	// The label may be there already: settle the reason now, not at the next edit.
	if err := uc.Reconcile(ctx, imageID, userID); err != nil {
		return r, nil
	}
	if fresh, err := uc.get(ctx, imageID); err == nil && fresh != nil {
		return fresh, nil
	}
	return r, nil
}

// get reads the image's request without writing it.
func (uc *RecheckUseCase) get(ctx context.Context, imageID string) (*port.RecheckRequest, error) {
	var got *port.RecheckRequest
	_, err := uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		got = current
		return nil, nil
	})
	return got, err
}

func (uc *RecheckUseCase) Assign(ctx context.Context, imageID, assigneeID string) (*port.RecheckRequest, error) {
	assigneeID = strings.TrimSpace(assigneeID)
	if err := validateAssignee(assigneeID); err != nil {
		return nil, err
	}
	return uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		if current == nil {
			return nil, errors.NewNotFoundError("no recheck request for this image")
		}
		current.AssigneeID, current.UpdatedAt = assigneeID, uc.now()
		return current, nil
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

func (uc *RecheckUseCase) RequestWorkspace(ctx context.Context, wsID, userID, note, assigneeID string) (int, error) {
	note, assigneeID = strings.TrimSpace(note), strings.TrimSpace(assigneeID)
	if err := validateRecheckReason(port.RecheckReasonDataset, note); err != nil {
		return 0, err
	}
	if err := validateAssignee(assigneeID); err != nil {
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
		next.AssigneeID = assigneeID
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

func (uc *RecheckUseCase) SetDone(ctx context.Context, viewer port.RecheckViewer, imageID string, done bool, outcome, note string) (*port.RecheckRequest, error) {
	userID := viewer.ID
	note = strings.TrimSpace(note)
	if done {
		if !port.IsRecheckOutcome(outcome) {
			return nil, errors.NewValidationError("outcome must be corrected, no_change, undecided or unsuitable",
				map[string]interface{}{"outcome": outcome})
		}
		if (outcome == port.RecheckOutcomeNoChange || outcome == port.RecheckOutcomeUndecided) && note == "" {
			return nil, errors.NewValidationError("this outcome needs a note saying why", map[string]interface{}{"outcome": outcome})
		}
		if utf8.RuneCountInString(note) > port.RecheckNoteMaxLen {
			return nil, errors.NewValidationError("note is too long", map[string]interface{}{"max": port.RecheckNoteMaxLen})
		}
	}
	now := uc.now()
	wasUnsuitable := false
	r, err := uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		if current == nil || !viewer.Sees(*current) {
			return nil, errors.NewNotFoundError("no recheck request for this image")
		}
		wasUnsuitable = current.Status == port.RecheckStatusDone && current.Outcome == port.RecheckOutcomeUnsuitable
		if done {
			current.Status, current.CompletedBy, current.CompletedAt = port.RecheckStatusDone, userID, &now
			current.Outcome, current.CompletionNote = outcome, note
		} else {
			reopen(current)
		}
		current.UpdatedAt = now
		return current, nil
	})
	if err != nil {
		return nil, err
	}
	// "Çalışmaya uygun değil" belongs to the image, so Veri Etiketleyici shows it
	// too: set it when the expert decides so here, clear it when that is undone.
	unsuitable := done && outcome == port.RecheckOutcomeUnsuitable
	if unsuitable || wasUnsuitable {
		at := func() time.Time { return now }
		if err := uc.images.Update(ctx, imageID, command.UnsuitableUpdates(unsuitable, userID, note, at)); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (uc *RecheckUseCase) Cancel(ctx context.Context, imageID string) error {
	return uc.store.Delete(ctx, imageID)
}
