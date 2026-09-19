package queries

import (
	"context"
	"sort"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type AnnotationQuery struct {
	*BaseQuery[*model.Annotation]
	*HierarchicalQueries[*model.Annotation]
	labels port.AnnotationLabelReader
}

func NewAnnotationQuery(repo port.AnnotationRepository, labels port.AnnotationLabelReader) *AnnotationQuery {
	return &AnnotationQuery{
		labels: labels,
		BaseQuery: &BaseQuery[*model.Annotation]{
			repo: repo,
		},
		HierarchicalQueries: &HierarchicalQueries[*model.Annotation]{
			repo: repo,
		},
	}
}

// LabelSetsByWsID groups the region annotations of a workspace by who drew
// them and for which annotation type. Sets come sorted by creator, then type;
// image ids sorted too, so the answer is the same on every call.
func (q *AnnotationQuery) LabelSetsByWsID(ctx context.Context, wsID string) ([]port.LabelSet, error) {
	rows, err := q.labels.LabelRows(ctx, wsID)
	if err != nil {
		return nil, err
	}
	return groupLabelSets(rows), nil
}

func groupLabelSets(rows []port.AnnotationLabelRow) []port.LabelSet {
	type key struct{ creator, annotationType string }
	type group struct {
		resources map[string]struct{}
		images    map[string]struct{}
		polygons  int
	}
	groups := map[key]*group{}
	// The name of a type is judged over the whole workspace, not per annotator.
	names := map[string]map[string]struct{}{}

	for _, r := range rows {
		k := key{r.CreatorID, r.AnnotationTypeID}
		g, ok := groups[k]
		if !ok {
			g = &group{resources: map[string]struct{}{}, images: map[string]struct{}{}}
			groups[k] = g
		}
		g.polygons++
		g.resources[r.Resource] = struct{}{}
		g.images[r.ImageID] = struct{}{}
		if names[r.AnnotationTypeID] == nil {
			names[r.AnnotationTypeID] = map[string]struct{}{}
		}
		names[r.AnnotationTypeID][r.Name] = struct{}{}
	}

	sets := make([]port.LabelSet, 0, len(groups))
	for k, g := range groups {
		set := port.LabelSet{
			CreatorID:        k.creator,
			AnnotationTypeID: k.annotationType,
			Resources:        sortedKeys(g.resources),
			Polygons:         g.polygons,
			ImageIDs:         sortedKeys(g.images),
		}
		if agreed := names[k.annotationType]; len(agreed) == 1 {
			for name := range agreed {
				set.Name = name
			}
		}
		sets = append(sets, set)
	}
	sort.Slice(sets, func(i, j int) bool {
		if sets[i].CreatorID != sets[j].CreatorID {
			return sets[i].CreatorID < sets[j].CreatorID
		}
		return sets[i].AnnotationTypeID < sets[j].AnnotationTypeID
	})
	return sets
}

func sortedKeys(m map[string]struct{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
