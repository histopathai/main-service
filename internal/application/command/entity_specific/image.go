package entityspecific

type CreateImageCommand struct {
	ID            string
	Name          string
	Type          string
	ContentType   string
	CreatorID     string
	ParentID      string
	ParentType    string
	Format        string
	OriginPath    string
	Size          *int64
	Width         *int
	Height        *int
	Status        *string
	ProcessedPath *string
}

func (c *CreateImageCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *CreateImageCommand) ToEntity() (interface{}, error) {
	// Implement conversion logic here
	return nil, nil
}

func (c *CreateImageCommand) GetID() string {
	return c.ID
}

type UpdateImageCommand struct {
	CreatorID     *string
	Status        *string
	Width         *int
	Height        *int
	Size          *int64
	ProcessedPath *string
}

func (c *UpdateImageCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *UpdateImageCommand) GetID() string {
	return ""
}

func (c *UpdateImageCommand) GetUpdates() map[string]interface{} {
	// Implement logic to return updates as a map
	return nil
}

func (c *UpdateImageCommand) GetUpdatebleFields() []string {
	// Implement logic to return a list of updatable fields
	return nil
}
