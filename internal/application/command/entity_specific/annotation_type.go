package entityspecific

type CreateAnnotationType struct {
	Name       string
	Type       string
	CreatorID  string
	TagType    string
	IsGlobal   bool
	IsRequired bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}

func (c *CreateAnnotationType) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *CreateAnnotationType) ToEntity() (interface{}, error) {
	// Implement conversion logic here
	return nil, nil
}

type UpdateAnnotationType struct {
	CreatorID  *string
	Type       *string
	TagType    *string
	IsGlobal   *bool
	IsRequired *bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}

func (c *UpdateAnnotationType) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *UpdateAnnotationType) GetUpdates() map[string]interface{} {
	// Implement logic to return updates as a map
	return nil
}

func (c *UpdateAnnotationType) GetUpdatebleFields() []string {
	// Implement logic to return a list of updatable fields
	return nil
}
