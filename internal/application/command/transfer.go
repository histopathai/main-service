package command

import "github.com/histopathai/main-service/internal/domain/vobj"

type TransferManyCommand struct {
	NewParent  string
	ParentType string
	IDs        []string
}

func (c *TransferManyCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if c.NewParent == "" {
		details["new_parent"] = "New parent ID is required"
	}

	if c.ParentType == "" {
		details["parent_type"] = "Parent type is required"
	}

	_, err := vobj.NewParentTypeFromString(string(c.ParentType))
	if err != nil {
		details["parent_type"] = "Invalid parent type"
	}

	if len(c.IDs) == 0 {
		details["ids"] = "At least one ID is required"
	}

	if len(details) > 0 {
		return details, false
	}

	return nil, true
}

func (c *TransferManyCommand) GetNewParent() string {
	return c.NewParent
}

func (c *TransferManyCommand) GetParentType() string {
	return c.ParentType
}

func (c *TransferManyCommand) GetIDs() []string {
	return c.IDs
}

type TransferCommand struct {
	NewParent  string
	ParentType string
	ID         string
}

func (c *TransferCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if c.NewParent == "" {
		details["new_parent"] = "New parent ID is required"
	}

	if c.ParentType == "" {
		details["parent_type"] = "Parent type is required"
	}

	_, err := vobj.NewParentTypeFromString(string(c.ParentType))
	if err != nil {
		details["parent_type"] = "Invalid parent type"
	}

	if c.ID == "" {
		details["id"] = "ID is required"
	}

	if len(details) > 0 {
		return details, false
	}

	return nil, true
}

func (c *TransferCommand) GetNewParent() string {
	return c.NewParent
}

func (c *TransferCommand) GetParentType() string {
	return c.ParentType
}

func (c *TransferCommand) GetID() string {
	return c.ID
}
