package query

type FilterOp string

func (op FilterOp) IsValid() bool {
	for _, validOp := range ValidOperators {
		if op == validOp {
			return true
		}
	}
	return false
}

const (
	OpEqual     FilterOp = "=="
	OpNotEqual  FilterOp = "!="
	OpGreater   FilterOp = ">"
	OpLess      FilterOp = "<"
	OpGreaterEq FilterOp = ">="
	OpLessEq    FilterOp = "<="
	OpIn        FilterOp = "in"
	OpNotIn     FilterOp = "not_in"
	OpContains  FilterOp = "contains"
)

var ValidOperators = []FilterOp{
	OpEqual,
	OpNotEqual,
	OpGreater,
	OpLess,
	OpGreaterEq,
	OpLessEq,
	OpIn,
	OpNotIn,
	OpContains,
}

type Filter struct {
	Field    string
	Operator FilterOp
	Value    interface{}
}
