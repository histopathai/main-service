package entityspecific

type CommandPoint struct {
	X float64
	Y float64
}

type CreateAnnotationCommand struct {
	Name       string
	Type       string
	CreatorID  string
	ParentID   string
	ParentType string
	TagType    string
	TagValue   any
	TagColor   *string
	Global     bool
	Points     []CommandPoint
}

func (c *CreateAnnotationCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *CreateAnnotationCommand) ToEntity() (interface{}, error) {
	// Implement conversion logic here
	return nil, nil
}

type UpdateAnnotationCommand struct {
	ID        string
	CreatorID *string
	TagType   *string
	TagName   *string
	TagValue  *any
	TagColor  *string
	Global    *bool
	Points    []CommandPoint
}

func (c *UpdateAnnotationCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *UpdateAnnotationCommand) GetID() string {
	return c.ID
}

func (c *UpdateAnnotationCommand) GetUpdates() map[string]interface{} {
	// Implement logic to return updates as a map
	return nil
}

func (c *UpdateAnnotationCommand) GetUpdatebleFields() []string {
	// Implement logic to return a list of updatable fields
	return nil
}
