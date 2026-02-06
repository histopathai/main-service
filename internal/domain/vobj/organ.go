package vobj

import (
	"regexp"
	"strings"
)

type OrganType string

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

func (o OrganType) String() string {
	return string(o)
}

// === Helpers ===
func normalizeOrganString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, `${1}_${2}`)

	return strings.ToLower(s)
}
