package command

import "github.com/histopathai/main-service/internal/shared/query"

type CountCommand struct {
	Filters []FilterCommand
}

func (c *CountCommand) ToQueryFilters(deleted bool) []query.Filter {
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
	for _, filterCmd := range c.Filters {
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
