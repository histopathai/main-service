package command

import "github.com/histopathai/main-service/internal/domain/vobj"

type TransferCommand struct {
	OldParent  string
	NewParent  string
	ParentType string
	IDs        []string
}

func (c *TransferCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if c.OldParent == "" {
		details["old_parent"] = "Old parent ID is required"
	}

	if c.NewParent == "" {
		details["new_parent"] = "New parent ID is required"
	}

	if c.ParentType == "" {
		details["parent_type"] = "Parent type is required"
	}

	_, err := vobj.NewParentTypeFromString(c.ParentType)
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
