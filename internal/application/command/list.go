package command

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
