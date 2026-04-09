package configutil

import (
	"fmt"
	"strings"
)

const tagConfig = "config"

// tagMetadata holds the parsed components of a config struct tag.
type tagMetadata struct {
	Name     string
	Required bool
	Default  string
	Prefix   string
}

// parseTag parses a "config" struct tag value into tagMetadata.
func parseTag(tag string) (tagMetadata, error) {
	if tag == "" {
		return tagMetadata{}, nil
	}

	parts := strings.Split(tag, ",")
	metadata := tagMetadata{
		Name: strings.TrimSpace(parts[0]),
	}

	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		key, value, found := strings.Cut(part, "=")
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if found && key == "" {
			return tagMetadata{}, &MalformedTagError{Tag: tag}
		}

		switch key {
		case "required":
			metadata.Required = true
		case "default":
			metadata.Default = value
		case "prefix":
			metadata.Prefix = value
		default:
			return tagMetadata{}, &MalformedTagError{Tag: tag, Err: fmt.Errorf("unknown option %q", key)}
		}
	}

	return metadata, nil
}
