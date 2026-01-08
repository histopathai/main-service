package service

import (
	"context"

	"github.com/histopathai/main-service/internal/application/commands"
	"github.com/histopathai/main-service/internal/application/usecases/common"
	"github.com/histopathai/main-service/internal/application/usecases/composite"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

type BaseService[T port.Entity] struct {
	readUc          *common.ReadUseCase[T]
	listUc          *common.ListUseCase[T]
	countUc         *common.CountUseCase[T]
	sDeleteUc       *common.SoftDeleteUseCase[T]
	filterUc        *common.FilterUseCase[T]
	findByParentUc  *common.FilterByParentUseCase[T]
	findByCreatorUc *common.FilterByCreatorUseCase[T]
	findByNameUc    *common.FilterByNameUseCase[T]
	hdeleteUc       *composite.DeleteUseCase
	createUC        *composite.CreateUseCase[T]
	// updateUC        *composite.UpdateUseCase[T]
	entityType vobj.EntityType
}

func NewBaseService[T port.Entity](
	readUc *common.ReadUseCase[T],
	listUc *common.ListUseCase[T],
	countUc *common.CountUseCase[T],
	sDeleteUc *common.SoftDeleteUseCase[T],
	filterUc *common.FilterUseCase[T],
	findByParentUc *common.FilterByParentUseCase[T],
	findByCreatorUc *common.FilterByCreatorUseCase[T],
	findByNameUc *common.FilterByNameUseCase[T],
	hdeleteUc *composite.DeleteUseCase,
	createUC *composite.CreateUseCase[T],
	// updateUC *composite.UpdateUseCase[T],
	entityType vobj.EntityType,
) *BaseService[T] {
	return &BaseService[T]{
		readUc:          readUc,
		listUc:          listUc,
		countUc:         countUc,
		sDeleteUc:       sDeleteUc,
		filterUc:        filterUc,
		findByParentUc:  findByParentUc,
		findByCreatorUc: findByCreatorUc,
		findByNameUc:    findByNameUc,
		hdeleteUc:       hdeleteUc,
		createUC:        createUC,
		// updateUC:        updateUC,
		entityType: entityType,
	}
}

func (bs *BaseService[T]) Create(ctx context.Context, cmd commands.CreateCommand[T]) (T, error) {
	entity, err := cmd.ToEntity()
	if err != nil {
		var zero T
		return zero, err
	}

	return bs.createUC.Execute(ctx, entity)
}

func (bs *BaseService[T]) GetByID(ctx context.Context, cmd commands.ReadCommand[T]) (T, error) {
	return bs.readUc.Execute(ctx, cmd.ID)
}

// func (bs *BaseService[T]) Update(ctx context.Context, cmd commands.UpdateCommand[T]) (T, error) {
// 	updates, err := cmd.GetUpdates()
// 	if err != nil {
// 		var zero T
// 		return zero, err
// 	}
//
// 	return bs.updateUC.Execute(ctx, cmd.GetID(), updates)
// }

func (bs *BaseService[T]) HardDelete(ctx context.Context, cmd commands.DeleteCommand[T]) error {
	return bs.hdeleteUc.Execute(ctx, cmd.ID, bs.entityType)
}

func (bs *BaseService[T]) SoftDelete(ctx context.Context, cmd commands.DeleteCommand[T]) error {
	return bs.sDeleteUc.Execute(ctx, cmd.ID)
}

func (bs *BaseService[T]) List(ctx context.Context, cmd commands.ListCommand[T]) (*query.Result[T], error) {
	return bs.listUc.Execute(ctx, cmd.Filters, cmd.Pagination)
}

func (bs *BaseService[T]) Count(ctx context.Context, cmd commands.CountCommand[T]) (int64, error) {
	return bs.countUc.Execute(ctx, cmd.Filters)
}

func (bs *BaseService[T]) FindByParent(ctx context.Context, cmd commands.FindByParentCommand[T]) (*query.Result[T], error) {
	return bs.findByParentUc.Execute(ctx, cmd.ParentRef.GetID(), cmd.ParentRef.Type.String(), cmd.Pagination)
}

func (bs *BaseService[T]) FindByCreator(ctx context.Context, cmd commands.FindByCreatorCommand[T]) (*query.Result[T], error) {
	return bs.findByCreatorUc.Execute(ctx, cmd.CreatorID, cmd.Pagination)
}

func (bs *BaseService[T]) FindByName(ctx context.Context, cmd commands.FindByNameCommand[T]) (*query.Result[T], error) {
	return bs.findByNameUc.Execute(ctx, cmd.Name, cmd.Pagination, &bs.entityType)
}
