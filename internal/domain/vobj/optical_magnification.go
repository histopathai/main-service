package vobj

type OpticalMagnification struct {
	Objective         *float64
	NativeLevel       *int
	ScanMagnification *float64
}

func (om *OpticalMagnification) GetMap() map[string]interface{} {
	result := make(map[string]interface{})
	if om.Objective != nil {
		result["Objective"] = *om.Objective
	}
	if om.NativeLevel != nil {
		result["NativeLevel"] = *om.NativeLevel
	}
	if om.ScanMagnification != nil {
		result["ScanMagnification"] = *om.ScanMagnification
	}
	return result
}
