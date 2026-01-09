package vobj

import (
	"regexp"
	"strings"

	"github.com/histopathai/main-service/internal/shared/errors"
)

type OrganType string

const (
	OrganUnknown        OrganType = "unknown"
	OrganBrain          OrganType = "brain"
	OrganLung           OrganType = "lung"
	OrganLiver          OrganType = "liver"
	OrganKidney         OrganType = "kidney"
	OrganHeart          OrganType = "heart"
	OrganStomach        OrganType = "stomach"
	OrganIntestineSmall OrganType = "small_intestine"
	OrganIntestineLarge OrganType = "large_intestine"
	OrganPancreas       OrganType = "pancreas"
	OrganSpleen         OrganType = "spleen"
	OrganBladder        OrganType = "bladder"
	OrganProstate       OrganType = "prostate"
	OrganTestis         OrganType = "testis"
	OrganOvary          OrganType = "ovary"
	OrganUterus         OrganType = "uterus"
	OrganSkin           OrganType = "skin"
	OrganBone           OrganType = "bone"
	OrganBoneMarrow     OrganType = "bone_marrow"
	OrganBreast         OrganType = "breast"
	OrganThyroid        OrganType = "thyroid"
	OrganLymphNode      OrganType = "lymph_node"
	OrganEsophagus      OrganType = "esophagus"
	OrganGallbladder    OrganType = "gallbladder"
	OrganSalivaryGland  OrganType = "salivary_gland"
	OrganAdrenalGland   OrganType = "adrenal_gland"
	OrganPlacenta       OrganType = "placenta"
	OrganEye            OrganType = "eye"
	OrganTongue         OrganType = "tongue"
)

func normalizeOrganString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, `${1}_${2}`)

	return strings.ToLower(s)
}

func (o OrganType) IsValid() bool {
	switch o {
	case OrganUnknown,
		OrganBrain,
		OrganLung,
		OrganLiver,
		OrganKidney,
		OrganHeart,
		OrganStomach,
		OrganIntestineSmall,
		OrganIntestineLarge,
		OrganPancreas,
		OrganSpleen,
		OrganBladder,
		OrganProstate,
		OrganTestis,
		OrganOvary,
		OrganUterus,
		OrganSkin,
		OrganBone,
		OrganBoneMarrow,
		OrganBreast,
		OrganThyroid,
		OrganLymphNode,
		OrganEsophagus,
		OrganGallbladder,
		OrganSalivaryGland,
		OrganAdrenalGland,
		OrganPlacenta,
		OrganEye,
		OrganTongue:
		return true
	default:
		return false
	}
}

func NewOrganTypeFromString(s string) (OrganType, error) {
	if s == "" {
		return OrganUnknown, nil
	}

	normalized := OrganType(normalizeOrganString(s))

	if !normalized.IsValid() {
		details := map[string]any{"value": s, "normalized": normalized}
		return "", errors.NewValidationError("invalid organ type", details)
	}

	return normalized, nil
}

func (o OrganType) String() string {
	return string(o)
}
