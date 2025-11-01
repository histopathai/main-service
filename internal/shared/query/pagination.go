package query

type Pagination struct {
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}
