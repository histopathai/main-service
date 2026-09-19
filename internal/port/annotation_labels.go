package port

import "context"

// AnnotationLabelRow is what a label-set summary needs of one region
// annotation: who drew it, for which annotation type, on which image — and
// nothing of the polygon, which is what makes reading a whole workspace cheap.
type AnnotationLabelRow struct {
	CreatorID        string
	AnnotationTypeID string
	Resource         string
	Name             string
	ImageID          string
}

// AnnotationLabelReader reads those rows for every region annotation of a workspace.
type AnnotationLabelReader interface {
	LabelRows(ctx context.Context, wsID string) ([]AnnotationLabelRow, error)
}

// LabelSet is what one annotator labelled with one annotation type in a
// workspace. Labels of different annotators, or of different types, are never
// evaluated together: patches are made from one label set at a time.
type LabelSet struct {
	CreatorID        string
	AnnotationTypeID string
	// Resources are the kinds of annotation behind the set: "manual", "model", "imported".
	Resources []string
	// Name is the annotations' own name when every annotation of the type (in
	// the whole workspace) agrees on it, otherwise empty — then the annotation
	// type record carries the name. The same rule dev-ingestor applies.
	Name     string
	Polygons int
	ImageIDs []string
}
