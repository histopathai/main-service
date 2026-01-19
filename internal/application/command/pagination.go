package command

import (
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

var MaxLimit = 100

type PaginateCommand struct {
	Limit  int
	Offset int
	Sort   *SortCommand
}

func (c *PaginateCommand) Validate() error {
	detail := make(map[string]interface{})
	if c.Limit < 0 {
		detail["limit"] = "Limit must be non-negative"

	}
	if c.Offset < 0 {
		detail["offset"] = "Offset must be non-negative"
	}
	if c.Sort != nil {
		if _, ok := c.Sort.Validate(); !ok {
			detail["sort"] = "Invalid sort command"
		}
	}
	if len(detail) > 0 {
		return errors.NewValidationError("validation failed", detail)
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
