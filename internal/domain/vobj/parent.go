package vobj

import "github.com/histopathai/main-service/internal/shared/errors"

type ParentType string

const (
	ParentTypeNone           ParentType = ""
	ParentTypeWorkspace      ParentType = "workspace"
	ParentTypePatient        ParentType = "patient"
	ParentTypeImage          ParentType = "image"
	ParentTypeAnnotationType ParentType = "annotation_type"
	ParentTypeAnnotation     ParentType = "annotation"
)

func (p ParentType) IsValid() bool {
	switch p {
	case ParentTypeWorkspace, ParentTypePatient, ParentTypeImage, ParentTypeAnnotationType, ParentTypeNone:
		return true
	default:
		return false
	}
}

func (p ParentType) String() string {
	return string(p)
}

func NewParentTypeFromString(s string) (ParentType, error) {
	if s == "" {
		return ParentTypeNone, nil
	}
	value := ParentType(s)
	if value.IsValid() {
		return value, nil
	} else {
		details := map[string]any{"value": s}
		return "", errors.NewValidationError("invalid parent type", details)
	}
}

type ParentRef struct {
	ID   string
	Type ParentType
}

func NewParentRef(id string, parentType ParentType) (*ParentRef, error) {
	if id == "" {
		return nil, errors.NewValidationError("parent ID is required", nil)
	}

	if !parentType.IsValid() {
		details := map[string]any{"parent_type": parentType}
		return nil, errors.NewValidationError("invalid parent type", details)
	}

	return &ParentRef{
		ID:   id,
		Type: parentType,
	}, nil
}

func (p *ParentRef) IsValid() bool {
	return p != nil && p.ID != "" && p.Type != ""
}

func (p *ParentRef) Equals(other *ParentRef) bool {
	if p == nil && other == nil {
		return true
	}
	if p == nil || other == nil {
		return false
	}
	return p.ID == other.ID && p.Type == other.Type
}

func (p *ParentRef) Copy() *ParentRef {
	if p == nil {
		return nil
	}
	return &ParentRef{
		ID:   p.ID,
		Type: p.Type,
	}
}

func (p *ParentRef) String() string {
	if p == nil {
		return "<nil>"
	}
	return string(p.Type) + ":" + p.ID
}

func (p *ParentRef) IsEmpty() bool {
	return p == nil || p.ID == "" || p.Type == ""
}

func (p *ParentRef) GetID() string {
	if p == nil {
		return ""
	}
	return p.ID
}

func (p *ParentRef) GetType() ParentType {
	if p == nil {
		return ""
	}
	return p.Type
}
