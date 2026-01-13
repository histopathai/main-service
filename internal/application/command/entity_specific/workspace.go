package entityspecific

type CreateWorkspaceCommand struct {
	Name            string
	Type            string
	CreatorID       string
	OrganType       string
	Organization    string
	Description     string
	License         string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

func (c *CreateWorkspaceCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *CreateWorkspaceCommand) ToEntity() (interface{}, error) {
	// Implement conversion logic here
	return nil, nil
}

type UpdateWorkspaceCommand struct {
	ID              string
	CreatorID       *string
	Name            *string
	OrganType       *string
	Organization    *string
	Description     *string
	License         *string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

func (c *UpdateWorkspaceCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *UpdateWorkspaceCommand) GetID() string {
	return c.ID
}

func (c *UpdateWorkspaceCommand) GetUpdates() map[string]interface{} {
	// Implement logic to return updates as a map
	return nil
}

func (c *UpdateWorkspaceCommand) GetUpdatebleFields() []string {
	// Implement logic to return a list of updatable fields
	return nil
}
