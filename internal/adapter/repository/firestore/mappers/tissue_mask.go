package mappers

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type TissueMaskMapper struct {
	*EntityMapper[*model.TissueMask]
}

func NewTissueMaskMapper() *TissueMaskMapper {
	return &TissueMaskMapper{
		EntityMapper: NewEntityMapper[*model.TissueMask](),
	}
}

func (m *TissueMaskMapper) ToFirestoreMap(entity *model.TissueMask) map[string]interface{} {
	doc := m.EntityMapper.ToFirestoreMap(entity)
	for k, v := range tissueMaskValues(entity) {
		doc[k] = v
	}
	return doc
}

// tissueMaskValues returns every tissue-mask-specific Firestore field. Nil
// pointers are written as null so a full update clears previous values.
func tissueMaskValues(e *model.TissueMask) map[string]interface{} {
	return map[string]interface{}{
		fields.TissueMaskWsID.FirestoreName():             e.WsID,
		fields.TissueMaskStatus.FirestoreName():           e.Status.String(),
		fields.TissueMaskAlgorithmVersion.FirestoreName(): e.AlgorithmVersion,
		fields.TissueMaskParams.FirestoreName():           tissueParamsToMap(e.Params),
		fields.TissueMaskPolygons.FirestoreName():         tissuePolygonsToMaps(e.Polygons),
		fields.TissueMaskPreviewWidth.FirestoreName():     e.PreviewWidth,
		fields.TissueMaskPreviewHeight.FirestoreName():    e.PreviewHeight,
		fields.TissueMaskLevel0Width.FirestoreName():      e.Level0Width,
		fields.TissueMaskLevel0Height.FirestoreName():     e.Level0Height,
		fields.TissueMaskDownsampleX.FirestoreName():      e.DownsampleX,
		fields.TissueMaskDownsampleY.FirestoreName():      e.DownsampleY,
		fields.TissueMaskTissueAreaRatio.FirestoreName():  e.TissueAreaRatio,
		fields.TissueMaskEditedBy.FirestoreName():         nullable(e.EditedBy),
		fields.TissueMaskEditedAt.FirestoreName():         nullable(e.EditedAt),
		fields.TissueMaskApprovedBy.FirestoreName():       nullable(e.ApprovedBy),
		fields.TissueMaskApprovedAt.FirestoreName():       nullable(e.ApprovedAt),
		fields.TissueMaskRevision.FirestoreName():         e.Revision,
		fields.TissueMaskRejectedBy.FirestoreName():       nullable(e.RejectedBy),
		fields.TissueMaskRejectedAt.FirestoreName():       nullable(e.RejectedAt),
		fields.TissueMaskRejectReason.FirestoreName():     nullable(e.RejectReason),
	}
}

func nullable[T any](v *T) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func tissueParamsToMap(p vobj.TissueParams) map[string]interface{} {
	return map[string]interface{}{
		"method":               p.Method,
		"saturation_threshold": p.SaturationThreshold,
		"gray_threshold":       p.GrayThreshold,
		"closing_radius":       p.ClosingRadius,
		"min_object_area":      p.MinObjectArea,
		"min_hole_area":        p.MinHoleArea,
		"simplify_tolerance":   p.SimplifyTolerance,
	}
}

func tissuePolygonsToMaps(polygons []vobj.TissuePolygon) []map[string]interface{} {
	out := make([]map[string]interface{}, len(polygons))
	for i, p := range polygons {
		holes := make([]map[string]interface{}, len(p.Holes))
		for j, h := range p.Holes {
			holes[j] = map[string]interface{}{"points": vobj.ToJSONPoints(h)}
		}
		out[i] = map[string]interface{}{
			"exterior": vobj.ToJSONPoints(p.Exterior),
			"holes":    holes,
		}
	}
	return out
}

func (m *TissueMaskMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.TissueMask, error) {
	entity, err := m.EntityMapper.ParseEntity(doc)
	if err != nil {
		return nil, err
	}
	return TissueMaskFromData(*entity, doc.Data()), nil
}

// TissueMaskFromData builds a mask from Firestore field values.
func TissueMaskFromData(entity vobj.Entity, data map[string]interface{}) *model.TissueMask {
	mask := &model.TissueMask{Entity: entity}

	mask.WsID, _ = data[fields.TissueMaskWsID.FirestoreName()].(string)
	if s, ok := data[fields.TissueMaskStatus.FirestoreName()].(string); ok {
		mask.Status = vobj.TissueMaskStatus(s)
	}
	mask.AlgorithmVersion, _ = data[fields.TissueMaskAlgorithmVersion.FirestoreName()].(string)
	if p, ok := data[fields.TissueMaskParams.FirestoreName()].(map[string]interface{}); ok {
		mask.Params = vobj.TissueParams{
			Method:              stringValue(p["method"]),
			SaturationThreshold: floatValue(p["saturation_threshold"]),
			GrayThreshold:       floatValue(p["gray_threshold"]),
			ClosingRadius:       int(floatValue(p["closing_radius"])),
			MinObjectArea:       int(floatValue(p["min_object_area"])),
			MinHoleArea:         int(floatValue(p["min_hole_area"])),
			SimplifyTolerance:   floatValue(p["simplify_tolerance"]),
		}
	}
	if raw, ok := data[fields.TissueMaskPolygons.FirestoreName()].([]interface{}); ok {
		mask.Polygons = make([]vobj.TissuePolygon, 0, len(raw))
		for _, item := range raw {
			pm, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			polygon := vobj.TissuePolygon{Exterior: pointsValue(pm["exterior"]), Holes: [][]vobj.Point{}}
			if holes, ok := pm["holes"].([]interface{}); ok {
				for _, h := range holes {
					if hm, ok := h.(map[string]interface{}); ok {
						polygon.Holes = append(polygon.Holes, pointsValue(hm["points"]))
					}
				}
			}
			mask.Polygons = append(mask.Polygons, polygon)
		}
	}
	mask.PreviewWidth = int(floatValue(data[fields.TissueMaskPreviewWidth.FirestoreName()]))
	mask.PreviewHeight = int(floatValue(data[fields.TissueMaskPreviewHeight.FirestoreName()]))
	mask.Level0Width = int(floatValue(data[fields.TissueMaskLevel0Width.FirestoreName()]))
	mask.Level0Height = int(floatValue(data[fields.TissueMaskLevel0Height.FirestoreName()]))
	mask.DownsampleX = floatValue(data[fields.TissueMaskDownsampleX.FirestoreName()])
	mask.DownsampleY = floatValue(data[fields.TissueMaskDownsampleY.FirestoreName()])
	mask.TissueAreaRatio = floatValue(data[fields.TissueMaskTissueAreaRatio.FirestoreName()])
	if v, ok := data[fields.TissueMaskEditedBy.FirestoreName()].(string); ok {
		mask.EditedBy = &v
	}
	if v, ok := data[fields.TissueMaskEditedAt.FirestoreName()].(time.Time); ok {
		mask.EditedAt = &v
	}
	if v, ok := data[fields.TissueMaskApprovedBy.FirestoreName()].(string); ok {
		mask.ApprovedBy = &v
	}
	if v, ok := data[fields.TissueMaskApprovedAt.FirestoreName()].(time.Time); ok {
		mask.ApprovedAt = &v
	}
	mask.Revision = int(floatValue(data[fields.TissueMaskRevision.FirestoreName()]))
	if v, ok := data[fields.TissueMaskRejectedBy.FirestoreName()].(string); ok {
		mask.RejectedBy = &v
	}
	if v, ok := data[fields.TissueMaskRejectedAt.FirestoreName()].(time.Time); ok {
		mask.RejectedAt = &v
	}
	if v, ok := data[fields.TissueMaskRejectReason.FirestoreName()].(string); ok {
		mask.RejectReason = &v
	}
	return mask
}

// Firestore returns integers as int64 and doubles as float64.
func floatValue(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case int:
		return float64(n)
	default:
		return 0
	}
}

func stringValue(v interface{}) string {
	s, _ := v.(string)
	return s
}

func pointsValue(v interface{}) []vobj.Point {
	raw, _ := v.([]interface{})
	points := make([]vobj.Point, 0, len(raw))
	for _, item := range raw {
		if pm, ok := item.(map[string]interface{}); ok {
			points = append(points, vobj.Point{X: floatValue(pm["X"]), Y: floatValue(pm["Y"])})
		}
	}
	return points
}

// MapUpdates accepts a "Mask" key holding a *model.TissueMask, which replaces
// every tissue-mask-specific field at once. Tissue masks are always written
// whole so the status, its audit fields and the revision stay consistent.
func (m *TissueMaskMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mapped, err := m.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}
	if v, ok := updates["Mask"]; ok {
		mask, ok := v.(*model.TissueMask)
		if !ok {
			return nil, errors.NewValidationError("invalid type for tissue mask update", nil)
		}
		for fk, fv := range tissueMaskValues(mask) {
			mapped[fk] = fv
		}
	}
	return mapped, nil
}

func (m *TissueMaskMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mapped, err := m.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}
	for _, f := range filters {
		for _, tf := range fields.TissueMaskFields {
			if f.Field == tf.APIName() || f.Field == tf.DomainName() {
				mapped = append(mapped, query.Filter{Field: tf.FirestoreName(), Operator: f.Operator, Value: f.Value})
				break
			}
		}
	}
	return mapped, nil
}
