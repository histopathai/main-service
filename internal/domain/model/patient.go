package model

type Patient struct {
	ID          string
	WorkspaceID string
	AnonymName  string
	Age         *int
	Gender      *string
	Race        *string
	Disease     *string
	Subtype     *string
	Grade       *string
	History     *string
	CreatedAt   string
	UpdatedAt   string
}
