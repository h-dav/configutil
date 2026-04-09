// Package configutil populates a struct from environment variables, flags, and .env files.
package configutil

import (
	"maps"
)

// Option configures the behaviour of [Set].
type Option func(*settings)

// WithFilepath loads key-value pairs from the file at path.
// Only .env files are supported.
func WithFilepath(path string) Option {
	return func(s *settings) {
		s.sources = append(s.sources, fileSource{filepath: path})
	}
}

// Set populates config from the registered sources.
// Sources are evaluated in order: files, environment variables, flags.
// Later sources overwrite earlier ones.
func Set(config any, opts ...Option) error {
	s := &settings{
		source: make(map[string]string),
	}

	for _, opt := range opts {
		opt(s)
	}

	// Default sources appended after file sources so they take precedence.
	s.sources = append(s.sources, environmentVariableSource{}, flagSource{})

	for _, src := range s.sources {
		values, err := src.Load()
		if err != nil {
			return err
		}

		maps.Copy(s.source, values)
	}

	return s.populateStruct(config)
}
