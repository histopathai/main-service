package command

type UpdateCommand interface {
	Validate() error
	GetID() string
	GetUpdates() map[string]interface{}
}
