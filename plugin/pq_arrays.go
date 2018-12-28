package plugin

// IsAbleToMakePQArray tells us if the specific field-type can automatically be turned into a PQ array:
func (p *OrmPlugin) IsAbleToMakePQArray(fieldType string) bool {
	switch fieldType {
	case "[]bool":
		return true
	case "[]float64":
		return true
	case "[]int64":
		return true
	case "[]string":
		return true
	default:
		return false
	}
}
