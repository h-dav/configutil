package configutil

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"strings"
)

// source provides configuration key-value pairs.
type source interface {
	Load() (map[string]string, error)
}

// flagSource loads values from command-line flags.
// It only reads flags if [flag.Parse] has been called.
type flagSource struct{}

func (flagSource) Load() (map[string]string, error) {
	m := make(map[string]string)
	if !flag.Parsed() {
		return m, nil
	}
	flag.Visit(func(f *flag.Flag) {
		m[f.Name] = f.Value.String()
	})
	return m, nil
}

// fileSource loads values from a .env file.
type fileSource struct {
	filepath string
}

func (s fileSource) Load() (map[string]string, error) {
	if !strings.EqualFold(filepath.Ext(s.filepath), ".env") {
		return nil, &FileTypeValidationError{Filepath: s.filepath}
	}
	return parseEnvFile(s.filepath)
}

// parseEnvFile reads a .env file and returns its key-value pairs.
func parseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, &OpenFileError{Err: err}
	}
	defer file.Close() //nolint:errcheck // read-only handle; close error is non-consequential

	m := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, err := parseEnvLine(line)
		if err != nil {
			return nil, err
		}
		m[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, &FileReadError{Filepath: path, Err: err}
	}

	return m, nil
}

// parseEnvLine parses a single KEY=VALUE line from a .env file.
// Quoted values preserve inline # characters. Unquoted values
// treat " #" as an inline comment delimiter.
func parseEnvLine(line string) (string, string, error) {
	key, value, found := strings.Cut(line, "=")
	if !found {
		return "", "", &ParseError{Line: line, Err: ErrSyntax}
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') {
		quote := value[0]
		if end := strings.IndexByte(value[1:], quote); end != -1 {
			value = value[1 : end+1]
		}
	} else {
		if idx := strings.Index(value, " #"); idx != -1 {
			value = value[:idx]
		}
		value = strings.TrimSpace(value)
	}

	return key, value, nil
}

// environmentVariableSource loads values from the process environment.
type environmentVariableSource struct{}

func (environmentVariableSource) Load() (map[string]string, error) {
	m := make(map[string]string)
	for _, kv := range os.Environ() {
		key, value, found := strings.Cut(kv, "=")
		if found {
			m[key] = value
		}
	}
	return m, nil
}
