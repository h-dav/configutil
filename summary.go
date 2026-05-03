package configutil

// LoadEntry describes the value written to a single config field by [Set].
type LoadEntry struct {
	FieldName string // Go struct field name, e.g. "Port"
	Key       string // config key looked up, e.g. "SERVER_PORT"
	Value     string // resolved string value that was decoded into the field
	Source    string // "env", "flag", a file path from WithFilepath, or "default"
}

// LoadSummary is populated when [WithSummary] is passed to [Set].
// Entries contains one record for each field that received a value.
// Fields that were neither set from a source nor had a default are absent.
// LoadSummary.Entries is only meaningful when [Set] returns nil (no error).
type LoadSummary struct {
	Entries []LoadEntry
}
