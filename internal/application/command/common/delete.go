package common

type DeleteCommand struct {
	ID string
}

type DeleteCommands struct {
	IDs []string
}

type SoftDeleteCommand struct {
	ID string
}

type SoftDeleteCommands struct {
	IDs []string
}
