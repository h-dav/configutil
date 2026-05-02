package configutil

import (
	"regexp"
	"strings"
)

// textReplacementRegex matches ${VAR} where VAR is a valid identifier.
// Patterns with spaces or other non-identifier characters are passed through as literal text.
var textReplacementRegex = regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_]*\}`)

// resolveReplacement expands ${VAR} references in value using s.source.
func (s *settings) resolveReplacement(value string) (string, error) {
	matches := textReplacementRegex.FindAllString(value, -1)

	for _, m := range matches {
		varName := strings.TrimPrefix(m, "${")
		varName = strings.TrimSuffix(varName, "}")

		replacement, exists := s.source[varName]
		if !exists {
			return "", &ReplacementError{VariableName: varName}
		}

		value = strings.ReplaceAll(value, m, replacement)
	}

	return value, nil
}
