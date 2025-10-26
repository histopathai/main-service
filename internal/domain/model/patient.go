package model

import "time"

type Patient struct {
	ID          string
	WorkspaceID string
	AnonymName  string
	Age         *int
	Gender      *string
	Race        *string
	Disease     *string
	Subtype     *string
	Grade       *int
	History     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
