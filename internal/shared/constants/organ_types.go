package constants

import "strings"

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

// AllOrganTypes provides a list of all defined organ types (useful for validation or enumeration).
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

func IsValidOrganType(s string) bool {

	lowers := strings.ToLower(s)
	for _, organ := range AllOrganTypes {
		if string(organ) == lowers {
			return true
		}
	}
	return false
}
