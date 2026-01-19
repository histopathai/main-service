package command

import (
	"github.com/histopathai/main-service/internal/shared/query"
)

var MaxLimit = 100

type PaginateCommand struct {
	Limit  int
	Offset int
	Sort   *SortCommand
}

func (c *PaginateCommand) Validate() error {
	if c.Limit < 0 {
		c.Limit = MaxLimit
	}
	if c.Offset < 0 {
		c.Offset = 0
	}
	if c.Sort != nil {
		if err := c.Sort.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *PaginateCommand) ToPagination() (*query.Pagination, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	pagination := &query.Pagination{
		Limit:  c.Limit,
		Offset: c.Offset,
	}

	if c.Sort != nil {
		field, direction, err := c.Sort.ToSort()
		if err != nil {
			return nil, err
		}
		pagination.SortBy = field
		pagination.SortDir = direction
	}

	return pagination, nil
}
