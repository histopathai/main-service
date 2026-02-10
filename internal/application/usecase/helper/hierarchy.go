package helper

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

type HierarchyService struct {
	uow port.UnitOfWorkFactory
}

func NewHierarchyService(uow port.UnitOfWorkFactory) *HierarchyService {
	return &HierarchyService{uow: uow}
}

// GetChildIDs fetches all child entity IDs under a parent
func (s *HierarchyService) GetChildIDs(
	ctx context.Context,
	parentType vobj.EntityType,
	parentID string,
) ([]string, error) {
	switch parentType {
	case vobj.EntityTypeWorkspace:
		return s.getPatientIDsUnderWorkspace(ctx, parentID)
	case vobj.EntityTypePatient:
		return s.getImageIDsUnderPatient(ctx, parentID)
	case vobj.EntityTypeImage:
		return s.getAnnotationIDsUnderImage(ctx, parentID)
	default:
		return nil, fmt.Errorf("unsupported parent type: %s", parentType)
	}
}

func (s *HierarchyService) getPatientIDsUnderWorkspace(ctx context.Context, workspaceID string) ([]string, error) {
	patientRepo := s.uow.GetPatientRepo()

	builder := query.NewBuilder()
	builder.Where(fields.EntityParentID.DomainName(), query.OpEqual, workspaceID)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
	return fetchAllIDs(ctx, patientRepo, builder.Build())
}

func (s *HierarchyService) getImageIDsUnderPatient(ctx context.Context, patientID string) ([]string, error) {
	imageRepo := s.uow.GetImageRepo()

	builder := query.NewBuilder()
	builder.Where(fields.EntityParentID.DomainName(), query.OpEqual, patientID)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	return fetchAllIDs(ctx, imageRepo, builder.Build())
}

func (s *HierarchyService) getAnnotationIDsUnderImage(ctx context.Context, imageID string) ([]string, error) {
	annotationRepo := s.uow.GetAnnotationRepo()

	builder := query.NewBuilder()
	builder.Where(fields.EntityParentID.DomainName(), query.OpEqual, imageID)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	return fetchAllIDs(ctx, annotationRepo, builder.Build())
}

func (s *HierarchyService) GetChildIDsByWsID(ctx context.Context, wsID string, entityType vobj.EntityType) ([]string, error) {
	switch entityType {
	case vobj.EntityTypePatient:
		return s.getPatientIDsUnderWorkspace(ctx, wsID)
	case vobj.EntityTypeImage:
		return s.getImageIDsHasWsID(ctx, wsID)
	case vobj.EntityTypeAnnotation:
		return s.getAnnotationIDsHasWsID(ctx, wsID)
	default:
		return nil, fmt.Errorf("unsupported entity type: %s", entityType)
	}
}

func (s *HierarchyService) getImageIDsHasWsID(ctx context.Context, wsID string) ([]string, error) {
	imageRepo := s.uow.GetImageRepo()

	builder := query.NewBuilder()
	builder.Where(fields.ImageWsID.DomainName(), query.OpEqual, wsID)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
	return fetchAllIDs(ctx, imageRepo, builder.Build())
}

func (s *HierarchyService) getAnnotationIDsHasWsID(ctx context.Context, wsID string) ([]string, error) {
	annotationRepo := s.uow.GetAnnotationRepo()

	builder := query.NewBuilder()
	builder.Where(fields.AnnotationWsID.DomainName(), query.OpEqual, wsID)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
	return fetchAllIDs(ctx, annotationRepo, builder.Build())
}

// fetchAllIDs - generic paginated ID fetcher
func fetchAllIDs[T port.Entity](
	ctx context.Context,
	repo port.Repository[T],
	spec query.Specification,
) ([]string, error) {
	const limit = 1000
	offset := 0
	var allIDs []string

	for {
		spec.Pagination = &query.Pagination{Limit: limit, Offset: offset}

		result, err := repo.Find(ctx, spec)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch entities: %w", err)
		}

		for _, entity := range result.Data {
			allIDs = append(allIDs, entity.GetID())
		}

		if !result.HasMore {
			break
		}

		offset += limit
	}

	return allIDs, nil
}
