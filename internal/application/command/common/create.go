package common

type CreateCommand interface {
	Validate() error
	ToEntity() (interface{}, error)
}
