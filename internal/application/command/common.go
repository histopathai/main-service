package command

type ReadCommand struct {
	ID string
}

type ReadAllCommand struct {
	IDs []string
}

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

type CountCommand struct {
	Filters []FilterCommand
}
