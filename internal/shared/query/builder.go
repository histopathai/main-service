package query

type Builder struct {
	spec Specification
}

func NewBuilder() *Builder {
	return &Builder{
		spec: Specification{
			Filters: make([]Filter, 0),
			Sorts:   make([]Sort, 0),
		},
	}
}

func (b *Builder) Where(field string, op Operator, value interface{}) *Builder {
	b.spec.Filters = append(b.spec.Filters, Filter{
		Field:    field,
		Operator: op,
		Value:    value,
	})
	return b
}

func (b *Builder) WhereEqual(field string, value interface{}) *Builder {
	return b.Where(field, OpEqual, value)
}

func (b *Builder) WhereIn(field string, values interface{}) *Builder {
	return b.Where(field, OpIn, values)
}

func (b *Builder) WhereNotDeleted() *Builder {
	return b.Where("deleted", OpEqual, false)
}

func (b *Builder) OrderBy(field string, direction SortDirection) *Builder {
	b.spec.Sorts = append(b.spec.Sorts, Sort{
		Field:     field,
		Direction: direction,
	})
	return b
}

func (b *Builder) OrderByAsc(field string) *Builder {
	return b.OrderBy(field, Asc)
}
func (b *Builder) OrderByDesc(field string) *Builder {
	return b.OrderBy(field, Desc)
}

func (b *Builder) Limit(limit int) *Builder {
	if b.spec.Pagination == nil {
		b.spec.Pagination = &Pagination{}
	}
	b.spec.Pagination.Limit = limit
	return b
}

func (b *Builder) Offset(offset int) *Builder {
	if b.spec.Pagination == nil {
		b.spec.Pagination = &Pagination{}
	}
	b.spec.Pagination.Offset = offset
	return b
}

func (b *Builder) Paginate(limit, offset int) *Builder {
	b.spec.Pagination = &Pagination{
		Limit:  limit,
		Offset: offset,
	}
	return b
}

func (b *Builder) Build() Specification {
	if b.spec.Pagination != nil {
		b.spec.Pagination.Normalize()
	}
	return b.spec
}
