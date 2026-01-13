package entityspecific

type CreatePatientCommand struct {
	Name       string
	Type       string
	CreatorID  string
	ParentID   string
	ParentType string
	Age        *int
	Gender     *string
	Race       *string
	Disease    *string
	Subtype    *string
	Grade      *string
	History    *string
}

func (c *CreatePatientCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *CreatePatientCommand) ToEntity() (interface{}, error) {
	// Implement conversion logic here
	return nil, nil
}

type UpdatePatientCommand struct {
	ID        string
	Name      *string
	CreatorID *string
	Age       *int
	Gender    *string
	Race      *string
	Disease   *string
	Subtype   *string
	Grade     *string
	History   *string
}

func (c *UpdatePatientCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *UpdatePatientCommand) GetID() string {
	return c.ID
}

func (c *UpdatePatientCommand) GetUpdates() map[string]interface{} {
	// Implement logic to return updates as a map
	return nil
}

func (c *UpdatePatientCommand) GetUpdatebleFields() []string {
	// Implement logic to return a list of updatable fields
	return nil
}
