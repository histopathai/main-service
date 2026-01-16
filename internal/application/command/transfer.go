package command

type TransferCommand struct {
	OldParent  string
	NewParent  string
	ParentType string
	IDs        []string
}
