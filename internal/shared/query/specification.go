package query

type Operator string

const (
	OpEqual          Operator = "=="
	OpNotEqual       Operator = "!="
	OpGreaterThan    Operator = ">"
	OpGreaterOrEqual Operator = ">="
	OpLessThan       Operator = "<"
	OpLessOrEqual    Operator = "<="
	OpIn             Operator = "in"
	OpNotIn          Operator = "not_in"
	OpContains       Operator = "array-contains"
	OpContainsAny    Operator = "array-contains-any"
)

func (o Operator) IsValid() bool {
	switch o {
	case OpEqual, OpNotEqual, OpGreaterThan, OpGreaterOrEqual,
		OpLessThan, OpLessOrEqual, OpIn, OpNotIn, OpContains, OpContainsAny:
		return true
	default:
		return false
	}
}

type SortDirection string

const (
	Asc  SortDirection = "asc"
	Desc SortDirection = "desc"
)

func (d SortDirection) IsValid() bool {
	return d == Asc || d == Desc
}

type Filter struct {
	Field    string
	Operator Operator
	Value    interface{}
}

type Sort struct {
	Field     string
	Direction SortDirection
}

type Pagination struct {
	Limit  int
	Offset int
}

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

func (p *Pagination) Normalize() {
	if p.Limit <= 0 {
		p.Limit = DefaultLimit
	}
	if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
}

type Specification struct {
	Filters    []Filter
	Sorts      []Sort
	Pagination *Pagination
}

type Result[T any] struct {
	Data    []T
	Limit   int
	Offset  int
	HasMore bool
}
