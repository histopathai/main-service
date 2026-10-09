package request

// RecheckRequestRequest sends an image to Ek Kontrol with one reason.
type RecheckRequestRequest struct {
	Reason string `json:"reason" binding:"required,oneof=subtype subtype_missing polygon polygon_missing global_label_missing other" example:"subtype"`
	// Note is optional, except for the reason "other", whose sentence it is.
	Note string `json:"note" example:"IDC mi ILC mi, E-cadherin kesitine de bakın."`
}

// RecheckStatusRequest marks a request done with the expert's answer, or open again.
type RecheckStatusRequest struct {
	Done *bool `json:"done" binding:"required" example:"true"`
	// Outcome is required when done: corrected, no_change, undecided or unsuitable.
	Outcome string `json:"outcome" example:"no_change"`
	// Note says why; required for every outcome but corrected.
	Note string `json:"note" example:"Kanal yapıları belirgin, tek sıra dizilim yok; IDC ile uyumlu."`
}

// RecheckWorkspaceRequest sends every image of a workspace to Ek Kontrol.
type RecheckWorkspaceRequest struct {
	// Note says why; shown once over the workspace in the tab. Required.
	Note string `json:"note" binding:"required" example:"Yeni yüklendi; poligonlar ve global etiketler uzmanca gözden geçirilmeli."`
}
