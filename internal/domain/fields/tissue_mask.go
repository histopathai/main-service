package fields

type TissueMaskField string

const (
	TissueMaskWsID             TissueMaskField = "ws_id"
	TissueMaskStatus           TissueMaskField = "status"
	TissueMaskAlgorithmVersion TissueMaskField = "algorithm_version"
	TissueMaskParams           TissueMaskField = "params"
	TissueMaskPolygons         TissueMaskField = "polygons"
	TissueMaskPreviewWidth     TissueMaskField = "preview_width"
	TissueMaskPreviewHeight    TissueMaskField = "preview_height"
	TissueMaskLevel0Width      TissueMaskField = "level0_width"
	TissueMaskLevel0Height     TissueMaskField = "level0_height"
	TissueMaskDownsampleX      TissueMaskField = "downsample_x"
	TissueMaskDownsampleY      TissueMaskField = "downsample_y"
	TissueMaskTissueAreaRatio  TissueMaskField = "tissue_area_ratio"
	TissueMaskEditedBy         TissueMaskField = "edited_by"
	TissueMaskEditedAt         TissueMaskField = "edited_at"
	TissueMaskApprovedBy       TissueMaskField = "approved_by"
	TissueMaskApprovedAt       TissueMaskField = "approved_at"
)

func (f TissueMaskField) APIName() string {
	return string(f)
}

func (f TissueMaskField) FirestoreName() string {
	return string(f)
}

func (f TissueMaskField) DomainName() string {
	switch f {
	case TissueMaskWsID:
		return "WsID"
	case TissueMaskStatus:
		return "Status"
	case TissueMaskAlgorithmVersion:
		return "AlgorithmVersion"
	case TissueMaskParams:
		return "Params"
	case TissueMaskPolygons:
		return "Polygons"
	case TissueMaskPreviewWidth:
		return "PreviewWidth"
	case TissueMaskPreviewHeight:
		return "PreviewHeight"
	case TissueMaskLevel0Width:
		return "Level0Width"
	case TissueMaskLevel0Height:
		return "Level0Height"
	case TissueMaskDownsampleX:
		return "DownsampleX"
	case TissueMaskDownsampleY:
		return "DownsampleY"
	case TissueMaskTissueAreaRatio:
		return "TissueAreaRatio"
	case TissueMaskEditedBy:
		return "EditedBy"
	case TissueMaskEditedAt:
		return "EditedAt"
	case TissueMaskApprovedBy:
		return "ApprovedBy"
	case TissueMaskApprovedAt:
		return "ApprovedAt"
	default:
		return ""
	}
}

func (f TissueMaskField) IsValid() bool {
	return f.DomainName() != ""
}

var TissueMaskFields = []TissueMaskField{
	TissueMaskWsID, TissueMaskStatus, TissueMaskAlgorithmVersion, TissueMaskParams, TissueMaskPolygons,
	TissueMaskPreviewWidth, TissueMaskPreviewHeight, TissueMaskLevel0Width, TissueMaskLevel0Height,
	TissueMaskDownsampleX, TissueMaskDownsampleY, TissueMaskTissueAreaRatio,
	TissueMaskEditedBy, TissueMaskEditedAt, TissueMaskApprovedBy, TissueMaskApprovedAt,
}

// TissueMaskFieldSet lists the fields a tissue mask list can filter and sort by.
type TissueMaskFieldSet struct{}

func NewTissueMaskFieldSet() *TissueMaskFieldSet {
	return &TissueMaskFieldSet{}
}

func (fs *TissueMaskFieldSet) IsValidField(field string) bool {
	switch TissueMaskField(field) {
	case TissueMaskParams, TissueMaskPolygons:
		return false
	}
	return EntityField(field).IsValid() || TissueMaskField(field).IsValid()
}

func (fs *TissueMaskFieldSet) GetAllFields() []string {
	all := make([]string, 0, len(EntityFields)+len(TissueMaskFields))
	for _, f := range EntityFields {
		all = append(all, f.APIName())
	}
	for _, f := range TissueMaskFields {
		if f != TissueMaskParams && f != TissueMaskPolygons {
			all = append(all, f.APIName())
		}
	}
	return all
}
