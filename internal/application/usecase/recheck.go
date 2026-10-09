package usecase

import (
	"context"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type recheckImageReader interface {
	Read(ctx context.Context, id string) (*model.Image, error)
}

type recheckPatientReader interface {
	Read(ctx context.Context, id string) (*model.Patient, error)
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

func (uc *RecheckUseCase) Request(ctx context.Context, imageID, userID, code, note string) (*port.RecheckRequest, error) {
	note = strings.TrimSpace(note)
	if !port.IsRecheckReason(code) {
		return nil, errors.NewValidationError("unknown recheck reason", map[string]interface{}{"code": code})
	}
	if code == port.RecheckReasonOther && note == "" {
		return nil, errors.NewValidationError("the reason \"other\" needs a note", nil)
	}
	if utf8.RuneCountInString(note) > port.RecheckNoteMaxLen {
		return nil, errors.NewValidationError("note is too long", map[string]interface{}{"max": port.RecheckNoteMaxLen})
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

	now := uc.now()
	return uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		next := current
		if next == nil {
			next = &port.RecheckRequest{ImageID: imageID, CreatedAt: now}
		}
		next.ImageName, next.PatientID, next.PatientName, next.WsID = image.Name, image.Parent.ID, patientName, image.WsID
		reason := port.RecheckReason{Code: code, Note: note, RequestedBy: userID, RequestedAt: now}
		replaced := false
		for i := range next.Reasons {
			if next.Reasons[i].Code == code && code != port.RecheckReasonOther {
				next.Reasons[i], replaced = reason, true
			}
		}
		if !replaced {
			next.Reasons = append(next.Reasons, reason)
		}
		next.Status, next.CompletedBy, next.CompletedAt = port.RecheckStatusOpen, "", nil
		next.UpdatedAt = now
		return next, nil
	})
}

func (uc *RecheckUseCase) SetDone(ctx context.Context, imageID, userID string, done bool) (*port.RecheckRequest, error) {
	now := uc.now()
	return uc.store.Update(ctx, imageID, func(current *port.RecheckRequest) (*port.RecheckRequest, error) {
		if current == nil {
			return nil, errors.NewNotFoundError("no recheck request for this image")
		}
		if done {
			current.Status, current.CompletedBy, current.CompletedAt = port.RecheckStatusDone, userID, &now
		} else {
			current.Status, current.CompletedBy, current.CompletedAt = port.RecheckStatusOpen, "", nil
		}
		current.UpdatedAt = now
		return current, nil
	})
}

func (uc *RecheckUseCase) Cancel(ctx context.Context, imageID string) error {
	return uc.store.Delete(ctx, imageID)
}
