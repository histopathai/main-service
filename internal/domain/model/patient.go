package model

type Patient struct {
	BaseEntity
	WorkspaceID string
	Age         *int
	Gender      *string
	Race        *string
	Disease     *string
	Subtype     *string
	Grade       *int
	History     *string
}
