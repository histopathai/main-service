package constants

import (
	"regexp"
	"strings"
)

// OrganType represents the standardized organ or tissue source for histopathology samples.
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

// AllOrganTypes holds all defined organ names for iteration or validation purposes.
var AllOrganTypes = []OrganType{
	OrganUnknown,
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
	OrganTongue,
}

// Precompiled regex (auto-generated or manually updated)
var organTypeRegex = regexp.MustCompile(`^(unknown|brain|lung|liver|kidney|heart|stomach|small_intestine|large_intestine|pancreas|spleen|bladder|prostate|testis|ovary|uterus|skin|bone|bone_marrow|breast|thyroid|lymph_node|esophagus|gallbladder|salivary_gland|adrenal_gland|placenta|eye|tongue)$`)

// normalizeOrganString normalizes a string to snake_case and lowercase
func normalizeOrganString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	// CamelCase → snake_case dönüşümü
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	s = re.ReplaceAllString(s, `${1}_${2}`)

	return strings.ToLower(s)
}

// IsValidOrganType returns true if the input matches any known organ name
func IsValidOrganType(s string) bool {
	return organTypeRegex.MatchString(normalizeOrganString(s))
}
