package repository

import "context"

// Pagination struct
type Pagination struct {
	Limit   int
	Offset  int
	SortBy  string
	Sortdir string
}

// Query response with pagination info
type QueryResult struct {
	Data    []map[string]interface{}
	Total   int
	Limit   int
	Offset  int
	HasMore bool
}

type FilterOp string

const (
	OpEqual           FilterOp = "=="
	OpNotEqual        FilterOp = "!="
	OpGreaterThan     FilterOp = ">"
	OpGreaterThanOrEq FilterOp = ">="
	OpLessThan        FilterOp = "<"
	OpLessThanOrEq    FilterOp = "<="
	OpIn              FilterOp = "in"
	OpNotIn           FilterOp = "not-in"
	OpArrayContains   FilterOp = "array-contains"
)

type Filter struct {
	Field string
	Op    FilterOp
	Value interface{}
}

// Transaction interface for repository layer
type Transaction interface {
	Create(col string, data map[string]interface{}) (string, error)
	Read(col string, docID string) (map[string]interface{}, error)
	Update(col string, docID string, updates map[string]interface{}) error
	Delete(col string, docID string) error
	Set(col string, docID string, data map[string]interface{}) error
}

// TransactionAdapter interface for that adapters must implement
type TransactionAdapter interface {
	Create(col string, data map[string]interface{}) (string, error)
	Read(col string, docID string) (map[string]interface{}, error)
	Update(col string, docID string, updates map[string]interface{}) error
	Delete(col string, docID string) error
	Set(col string, docID string, data map[string]interface{}) error
}

type RepositoryTransaction struct {
	adapterTx TransactionAdapter
}

func (rt *RepositoryTransaction) Create(col string, data map[string]interface{}) (string, error) {
	return rt.adapterTx.Create(col, data)
}

func (rt *RepositoryTransaction) Read(col string, docID string) (map[string]interface{}, error) {
	return rt.adapterTx.Read(col, docID)
}

func (rt *RepositoryTransaction) Update(col string, docID string, updates map[string]interface{}) error {
	return rt.adapterTx.Update(col, docID, updates)
}

func (rt *RepositoryTransaction) Delete(col string, docID string) error {
	return rt.adapterTx.Delete(col, docID)
}

func (rt *RepositoryTransaction) Set(col string, docID string, data map[string]interface{}) error {
	return rt.adapterTx.Set(col, docID, data)
}

// Repository interface for repository layer
type Repository interface {
	Create(ctx context.Context, col string, data map[string]interface{}) (string, error)
	Read(ctx context.Context, col, docID string) (map[string]interface{}, error)
	Update(ctx context.Context, col, docID string, updates map[string]interface{}) error
	Delete(ctx context.Context, col, docID string) error
	Set(ctx context.Context, col, docID string, data map[string]interface{}) error
	Query(ctx context.Context, col string, filters map[string]interface{}, pagination Pagination) (*QueryResult, error)
	List(ctx context.Context, col string, filters []Filter, pagination Pagination) (*QueryResult, error)
	Exists(ctx context.Context, col, docID string) (bool, error)
	RunTransaction(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) error
}

// The RepositoryAdapter interface for that adapters must implement
type RepositoryAdapter interface {
	Create(ctx context.Context, col string, data map[string]interface{}) (string, error)
	Read(ctx context.Context, col, docID string) (map[string]interface{}, error)
	Update(ctx context.Context, col, docID string, updates map[string]interface{}) error
	Delete(ctx context.Context, col, docID string) error
	Set(ctx context.Context, col, docID string, data map[string]interface{}) error
	Query(ctx context.Context, col string, filters map[string]interface{}, pagination Pagination) (*QueryResult, error)
	List(ctx context.Context, col string, filters []Filter, pagination Pagination) (*QueryResult, error)
	Exists(ctx context.Context, col, docID string) (bool, error)
	RunTransaction(ctx context.Context, fn func(ctx context.Context, tx TransactionAdapter) error) error
}

type MainRepository struct {
	adapter RepositoryAdapter
}

func NewMainRepository(adapter RepositoryAdapter) *MainRepository {
	return &MainRepository{
		adapter: adapter,
	}
}

func (r *MainRepository) Create(ctx context.Context, col string, data map[string]interface{}) (string, error) {
	return r.adapter.Create(ctx, col, data)
}

func (r *MainRepository) Read(ctx context.Context, col, docID string) (map[string]interface{}, error) {
	return r.adapter.Read(ctx, col, docID)
}

func (r *MainRepository) Update(ctx context.Context, col, docID string, updates map[string]interface{}) error {
	return r.adapter.Update(ctx, col, docID, updates)
}

func (r *MainRepository) Delete(ctx context.Context, col, docID string) error {
	return r.adapter.Delete(ctx, col, docID)
}

func (r *MainRepository) Set(ctx context.Context, col, docID string, data map[string]interface{}) error {
	return r.adapter.Set(ctx, col, docID, data)
}

func (r *MainRepository) Query(ctx context.Context, col string, filters map[string]interface{}, pagination Pagination) (*QueryResult, error) {
	return r.adapter.Query(ctx, col, filters, pagination)
}

func (r *MainRepository) List(ctx context.Context, col string, filters []Filter, pagination Pagination) (*QueryResult, error) {
	return r.adapter.List(ctx, col, filters, pagination)
}

func (r *MainRepository) Exists(ctx context.Context, col, docID string) (bool, error) {
	return r.adapter.Exists(ctx, col, docID)
}

func (r *MainRepository) RunTransaction(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) error {
	return r.adapter.RunTransaction(ctx, func(ctx context.Context, adapterTx TransactionAdapter) error {
		tx := &RepositoryTransaction{
			adapterTx: adapterTx,
		}
		return fn(ctx, tx)
	})
}
