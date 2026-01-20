package command

import "github.com/histopathai/main-service/internal/shared/query"

type ListCommand struct {
	Filters    *[]FilterCommand
	Pagination *PaginateCommand
}

func (c *ListCommand) Validate() error {

	if c.Filters != nil {
		for _, filter := range *c.Filters {
			if err := filter.Validate(); err != nil {
				return err
			}
		}
	}

	if c.Pagination != nil {
		if err := c.Pagination.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c *ListCommand) ToQueryFilters(deleted bool) []query.Filter {
	if c.Filters == nil {
		if !deleted {
			return []query.Filter{
				{
					Field:    "deleted",
					Operator: query.OpEqual,
					Value:    false,
				},
			}
		}
	}

	var filters []query.Filter
	for _, filterCmd := range *c.Filters {
		filter, err := filterCmd.ToFilter()
		if err == nil {
			filters = append(filters, filter)
		}
	}
	if !deleted {
		filters = append(filters, query.Filter{
			Field:    "deleted",
			Operator: query.OpEqual,
			Value:    false,
		})
	}

	return filters
}
func (c *ListCommand) ToPagination() *query.Pagination {
	if c.Pagination == nil {
		return nil
	}

	return &query.Pagination{
		Limit:  c.Pagination.Limit,
		Offset: c.Pagination.Offset,
	}
}
