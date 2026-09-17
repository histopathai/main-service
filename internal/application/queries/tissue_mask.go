package queries

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

type TissueMaskQuery struct {
	*BaseQuery[*model.TissueMask]
	*HierarchicalQueries[*model.TissueMask]
	wsRepo    port.WorkspaceRepository
	imageRepo port.ImageRepository
}

func NewTissueMaskQuery(tmRepo port.TissueMaskRepository, wsRepo port.WorkspaceRepository, imageRepo port.ImageRepository) *TissueMaskQuery {
	return &TissueMaskQuery{
		BaseQuery:           &BaseQuery[*model.TissueMask]{repo: tmRepo},
		HierarchicalQueries: &HierarchicalQueries[*model.TissueMask]{repo: tmRepo},
		wsRepo:              wsRepo,
		imageRepo:           imageRepo,
	}
}

// GetWorkspaceStats iterates all workspaces and counts images needing masks
// plus tissue_mask documents by status. Large images (>3000px) that are
// processed and not deleted are counted.
func (q *TissueMaskQuery) GetWorkspaceStats(ctx context.Context) ([]port.TissueMaskWorkspaceStats, error) {
	// List all non-deleted workspaces.
	wsResult, err := q.wsRepo.Find(ctx, query.Specification{
		Filters: []query.Filter{{
			Field:    fields.EntityIsDeleted.DomainName(),
			Operator: query.OpEqual,
			Value:    false,
		}},
	})
	if err != nil {
		return nil, err
	}

	stats := make([]port.TissueMaskWorkspaceStats, 0, len(wsResult.Data))

	for _, ws := range wsResult.Data {
		// Count processed large images for this workspace.
		imgResult, imgErr := q.imageRepo.Find(ctx, query.Specification{
			Filters: []query.Filter{
				{Field: "WsID", Operator: query.OpEqual, Value: ws.ID},
				{Field: fields.EntityIsDeleted.DomainName(), Operator: query.OpEqual, Value: false},
			},
		})
		if imgErr != nil {
			return nil, imgErr
		}

		totalNeedsMasking := 0
		for _, img := range imgResult.Data {
			if img.Processing != nil && img.Processing.Status == vobj.StatusProcessed &&
				((img.Width != nil && *img.Width > 3000) || (img.Height != nil && *img.Height > 3000)) {
				totalNeedsMasking++
			}
		}

		if totalNeedsMasking == 0 {
			continue // skip workspaces with no masking-eligible images
		}

		// Count tissue masks by status for this workspace.
		tmResult, tmErr := q.BaseQuery.repo.Find(ctx, query.Specification{
			Filters: []query.Filter{
				{Field: "WsID", Operator: query.OpEqual, Value: ws.ID},
				{Field: fields.EntityIsDeleted.DomainName(), Operator: query.OpEqual, Value: false},
			},
		})
		if tmErr != nil {
			return nil, tmErr
		}

		var approved, edited, autoCount, rejected, missing int
		masksByImageID := make(map[string]*model.TissueMask)
		for _, tm := range tmResult.Data {
			masksByImageID[tm.ID] = tm
		}

		// Walk eligible images and classify each.
		for _, img := range imgResult.Data {
			if img.Processing == nil || img.Processing.Status != vobj.StatusProcessed {
				continue
			}
			if (img.Width == nil || *img.Width <= 3000) && (img.Height == nil || *img.Height <= 3000) {
				continue
			}

			tm, ok := masksByImageID[img.ID]
			if !ok {
				missing++
				continue
			}
			switch tm.Status {
			case vobj.TissueMaskStatusApproved:
				approved++
			case vobj.TissueMaskStatusEdited:
				edited++
			case vobj.TissueMaskStatusAuto:
				autoCount++
			case vobj.TissueMaskStatusRejected:
				rejected++
			}
		}

		done := approved + rejected
		remaining := totalNeedsMasking - done

		stats = append(stats, port.TissueMaskWorkspaceStats{
			WorkspaceID:   ws.ID,
			WorkspaceName: ws.Name,
			TotalImages:   totalNeedsMasking,
			Approved:      approved,
			Edited:        edited,
			Auto:          autoCount,
			Rejected:      rejected,
			Missing:       missing,
			Done:          done,
			Remaining:     remaining,
		})
	}

	return stats, nil
}
