// Package configutil populates a struct from environment variables, flags, and .env files.
package configutil

// Option configures the behaviour of [Set].
type Option func(*settings)

// WithFilepath loads key-value pairs from the file at path.
// Only .env files are supported.
func WithFilepath(path string) Option {
	return func(s *settings) {
		s.sources = append(s.sources, fileSource{filepath: path})
	}
}

// WithSummary registers out to receive provenance information after Set returns.
// Each config field that receives a value will have a corresponding LoadEntry
// in out.Entries describing the field name, key, resolved value, and source.
// WithSummary(nil) is a no-op.
func WithSummary(out *LoadSummary) Option {
	return func(s *settings) {
		if out != nil {
			s.summary = out
		}
	}
}

// Set populates config from the registered sources.
// Sources are evaluated in order: files, environment variables, flags.
// Later sources overwrite earlier ones.
func Set(config any, opts ...Option) error {
	s := &settings{
		source:     make(map[string]string),
		provenance: make(map[string]string),
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

		for k, v := range values {
			s.source[k] = v
			s.provenance[k] = src.Name()
		}
	}

	return s.populateStruct(config)
}
