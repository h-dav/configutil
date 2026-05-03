package configutil

type settings struct {
	source  map[string]string
	sources []source
}

type entry struct {
	key, value, fieldName string
}
