package configutil

type settings struct {
	source     map[string]string
	sources    []source
	provenance map[string]string
	summary    *LoadSummary
}

type entry struct {
	key, value, fieldName string
}
