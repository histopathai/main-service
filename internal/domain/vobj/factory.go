package vobj

import (
	"errors"
	"time"
)

//============================ ContentType Factory=========================//

func NewContentTypeFromString(s string) (ContentType, error) {

	if s == "" {
		return "", errors.New("content type string is empty")
	}
	value := ContentType(s)
	if value.IsValid() {
		return value, nil
	} else {
		return "", errors.New("invalid content type: " + s)
	}
}

func NewContentProviderFromString(s string) (ContentProvider, error) {
	if s == "" {
		return "", errors.New("content provider string is empty")
	}
	value := ContentProvider(s)
	if value.IsValid() {
		return value, nil
	} else {
		return "", errors.New("invalid content provider: " + s)
	}
}

func GetContentTypeFromExtension(ext string) ContentType {
	switch ext {
	case ".svs":
		return ContentTypeImageSVS
	case ".tif", ".tiff":
		return ContentTypeImageTIFF
	case ".ndpi":
		return ContentTypeImageNDPI
	case ".vms":
		return ContentTypeImageVMS
	case ".vmu":
		return ContentTypeImageVMU
	case ".scn":
		return ContentTypeImageSCN
	case ".mrz":
		return ContentTypeImageMIRAX
	case ".bif":
		return ContentTypeImageBIF
	case ".dng":
		return ContentTypeImageDNG
	case ".bmp":
		return ContentTypeImageBMP
	case ".jpg", ".jpeg":
		return ContentTypeImageJPEG
	case ".png":
		return ContentTypeImagePNG
	case ".zip":
		return ContentTypeApplicationZip
	case ".json":
		return ContentTypeApplicationJSON
	case ".dzi":
		return ContentTypeApplicationDZI
	default:
		return ContentTypeApplicationOctetStream
	}
}

//============================ EntityType Factory=========================//

func NewEntityTypeFromString(s string) (EntityType, error) {
	if s == "" {
		return "", errors.New("entity type cannot be empty")
	}
	value := EntityType(s)
	if value.IsValid() {
		return value, nil
	} else {
		return "", errors.New("invalid entity type: " + s)
	}
}

func NewEntity(entityType EntityType, name string, creatorID string, parent *ParentRef) (*Entity, error) {
	if !entityType.IsValid() {
		return nil, errors.New("invalid entity type: " + string(entityType))
	}

	if creatorID == "" {
		return nil, errors.New("creator ID is required")
	}

	return &Entity{
		EntityType: entityType,
		Name:       name,
		CreatorID:  creatorID,
		Parent:     *parent,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil
}

// ============================ Parent Factory=========================//

func NewParentTypeFromString(s string) (ParentType, error) {
	if s == "" {
		return ParentTypeNone, nil
	}
	value := ParentType(s)
	if value.IsValid() {
		return value, nil
	} else {
		return "", errors.New("invalid parent type")
	}
}

func NewParentRef(id string, parentType ParentType) (*ParentRef, error) {
	if !parentType.IsValid() {
		return nil, errors.New("invalid parent type")
	}

	if parentType != ParentTypeNone && id == "" {
		return nil, errors.New("parent ID is required")
	}

	return &ParentRef{
		ID:   id,
		Type: parentType,
	}, nil
}

// ============================ TagType Factory=========================//

func NewTagTypeFromString(s string) (TagType, error) {
	if s == "" {
		return "", errors.New("tag type cannot be empty")
	}
	tagType := TagType(s)
	if !tagType.IsValid() {
		return "", errors.New("invalid tag type")
	}
	return tagType, nil
}

// ============================ ImageStatus Factory=========================//

func NewImageStatusFromString(s string) (ImageStatus, error) {
	is := ImageStatus(s)
	if !is.IsValid() {
		return "", errors.New("invalid ImageStatus: " + s)
	}
	return is, nil
}

// ============================ OrganType Factory=========================//

func NewOrganTypeFromString(s string) (OrganType, error) {
	if s == "" {
		return OrganUnknown, nil
	}

	normalized := OrganType(normalizeOrganString(s))

	if !normalized.IsValid() {
		return "", errors.New("invalid organ type")
	}

	return normalized, nil
}

// =========================== Point Factory=========================//

func NewPoint(x, y float64) Point {
	return Point{X: x, Y: y}
}

func ToJSONPoints(points []Point) []map[string]float64 {
	jsonPoints := make([]map[string]float64, len(points))
	for i, point := range points {
		jsonPoints[i] = map[string]float64{
			"X": point.X,
			"Y": point.Y,
		}
	}
	return jsonPoints
}

func FromJSONPoints(jsonPoints []map[string]float64) []Point {
	points := make([]Point, len(jsonPoints))
	for i, jp := range jsonPoints {
		points[i] = Point{
			X: jp["X"],
			Y: jp["Y"],
		}
	}
	return points
}
