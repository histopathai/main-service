package model

import "github.com/histopathai/main-service/internal/domain/vobj"

type Workspace struct {
	vobj.Entity
	OrganType       string
	Organization    string
	Description     string
	License         string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}
