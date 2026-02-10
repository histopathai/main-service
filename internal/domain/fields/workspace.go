package fields

type WorkspaceField string

const (
	WorkspaceOrganType       WorkspaceField = "organ_type"
	WorkspaceOrganization    WorkspaceField = "organization"
	WorkspaceDescription     WorkspaceField = "description"
	WorkspaceLicense         WorkspaceField = "license"
	WorkspaceResourceURL     WorkspaceField = "resource_url"
	WorkspaceReleaseYear     WorkspaceField = "release_year"
	WorkspaceAnnotationTypes WorkspaceField = "annotation_types"
)

func (f WorkspaceField) APIName() string {
	return string(f)
}

func (f WorkspaceField) FirestoreName() string {
	return string(f)
}

func (f WorkspaceField) DomainName() string {
	switch f {
	case WorkspaceOrganType:
		return "OrganType"
	case WorkspaceOrganization:
		return "Organization"
	case WorkspaceDescription:
		return "Description"
	case WorkspaceLicense:
		return "License"
	case WorkspaceResourceURL:
		return "ResourceURL"
	case WorkspaceReleaseYear:
		return "ReleaseYear"
	case WorkspaceAnnotationTypes:
		return "AnnotationTypes"
	default:
		return ""
	}
}

func (f WorkspaceField) IsValid() bool {
	switch f {
	case WorkspaceOrganType, WorkspaceOrganization, WorkspaceDescription, WorkspaceLicense, WorkspaceResourceURL, WorkspaceReleaseYear, WorkspaceAnnotationTypes:
		return true
	default:
		return false
	}
}

var WorkspaceFields = []WorkspaceField{
	WorkspaceOrganType, WorkspaceOrganization, WorkspaceDescription, WorkspaceLicense, WorkspaceResourceURL, WorkspaceReleaseYear, WorkspaceAnnotationTypes,
}
