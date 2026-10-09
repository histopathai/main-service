package request

// RecheckRequestRequest sends an image to Ek Kontrol with one reason.
type RecheckRequestRequest struct {
	Reason string `json:"reason" binding:"required,oneof=subtype polygon polygon_missing global_label_missing other" example:"subtype"`
	// Note is optional, except for the reason "other", whose sentence it is.
	Note string `json:"note" example:"IDC mi ILC mi, E-cadherin kesitine de bakın."`
}

// RecheckStatusRequest marks a request done, or open again.
type RecheckStatusRequest struct {
	Done *bool `json:"done" binding:"required" example:"true"`
}

// RecheckWorkspaceRequest sends every image of a workspace to Ek Kontrol.
type RecheckWorkspaceRequest struct {
	// Note says why; shown once over the workspace in the tab. Required.
	Note string `json:"note" binding:"required" example:"Yeni yüklendi; poligonlar ve global etiketler uzmanca gözden geçirilmeli."`
}
