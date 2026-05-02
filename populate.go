package configutil

import (
	"reflect"
)

// populateStruct validates config and walks its fields.
func (s *settings) populateStruct(config any) error {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return &InvalidConfigTypeError{ProvidedType: config}
	}

	return s.walkFields(v.Elem(), "")
}

// walkFields iterates over struct fields and populates them.
func (s *settings) walkFields(v reflect.Value, prefix string) error {
	for i := range v.NumField() {
		field := v.Type().Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		if err := s.handleField(field, fieldVal, prefix); err != nil {
			return &FieldError{FieldName: field.Name, Err: err}
		}
	}
	return nil
}

// handleField parses the struct tag and populates the field.
func (s *settings) handleField(field reflect.StructField, value reflect.Value, prefix string) error {
	tag := field.Tag.Get(tagConfig)
	metadata, err := parseTag(tag)
	if err != nil {
		return err
	}

	// Nested struct with prefix.
	if field.Type.Kind() == reflect.Struct && metadata.Prefix != "" {
		return s.walkFields(value, prefix+metadata.Prefix)
	}

	var wasSet bool
	if metadata.Name != "" {
		key := prefix + metadata.Name
		if val, exists := s.source[key]; exists {
			resolved, err := s.resolveReplacement(val)
			if err != nil {
				return err
			}
			if err := s.setFieldValue(value, entry{key: key, value: resolved}); err != nil {
				return err
			}
			wasSet = true
		}
	}

	if !wasSet && value.IsZero() && metadata.Default != "" {
		if err := s.setFieldValue(value, entry{key: field.Name, value: metadata.Default}); err != nil {
			return &MalformedDefaultError{FieldName: field.Name, Default: metadata.Default, Err: err}
		}
	}

	if metadata.Required && !wasSet {
		return &RequiredFieldError{FieldName: field.Name}
	}

	return nil
}
