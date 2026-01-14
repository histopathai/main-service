package common

type CreateCommand interface {
	Validate() (interface{}, interface{}, error)
	ToEntity() (interface{}, error)
}
