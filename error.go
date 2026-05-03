package configutil

import (
	"errors"
	"fmt"
)

// Base error types for common categories.
var (
	ErrInvalidConfig = errors.New("invalid config")
	ErrUnsupported   = errors.New("unsupported")
	ErrRequired      = errors.New("required")
	ErrFile          = errors.New("file error")
	ErrParse         = errors.New("parse error")
	ErrConversion    = errors.New("conversion error")
	ErrReplacement   = errors.New("replacement error")
	ErrTag           = errors.New("tag error")
	ErrSyntax        = errors.New("invalid syntax")
)

// FileTypeValidationError occurs when the .env config file fails to open.
type FileTypeValidationError struct {
	Filepath string
}

func (e *FileTypeValidationError) Error() string {
	return fmt.Sprintf("file extension is not a valid environment file: %q", e.Filepath)
}

func (e *FileTypeValidationError) Unwrap() error { return ErrFile }

// OpenFileError occurs when the .env config file fails to open.
type OpenFileError struct {
	Filepath string
	Err      error
}

func (e *OpenFileError) Error() string {
	return fmt.Sprintf("opening config file %q: %v", e.Filepath, e.Err)
}

func (e *OpenFileError) Unwrap() []error { return []error{e.Err, ErrFile} }

// FieldConversionError occurs when a field on the config struct fails to be set.
type FieldConversionError struct {
	FieldName  string
	TargetType string
	Err        error
}

func (e *FieldConversionError) Error() string {
	return fmt.Sprintf("failed to convert field %q to %v: %v", e.FieldName, e.TargetType, e.Err)
}

func (e *FieldConversionError) Unwrap() []error { return []error{e.Err, ErrConversion} }

// UnsupportedFieldTypeError occurs when the a field type on the config struct is not compatible.
type UnsupportedFieldTypeError struct {
	FieldName string
	FieldType string
}

func (e *UnsupportedFieldTypeError) Error() string {
	return fmt.Sprintf("unsupported field type %q: %s", e.FieldName, e.FieldType)
}

func (e *UnsupportedFieldTypeError) Unwrap() error { return ErrUnsupported }

// InvalidConfigTypeError occurs when config is not a pointer to a struct.
type InvalidConfigTypeError struct {
	ProvidedType any
}

func (e *InvalidConfigTypeError) Error() string {
	return fmt.Sprintf("output must be a pointer to a struct, got %T", e.ProvidedType)
}

func (e *InvalidConfigTypeError) Unwrap() error { return ErrInvalidConfig }

// RequiredFieldError occurs when a required field is not set in the configuration.
type RequiredFieldError struct {
	FieldName string
}

func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("required field is not set in configuration: %q", e.FieldName)
}

func (e *RequiredFieldError) Unwrap() error { return ErrRequired }

// ReplacementError occurs when the configuration variable being used for replacement is not set.
type ReplacementError struct {
	VariableName string
}

func (e *ReplacementError) Error() string {
	return fmt.Sprintf("configuration variable for replacement is not set: %v", e.VariableName)
}

func (e *ReplacementError) Unwrap() error { return ErrReplacement }

// ParseError occurs when a line from the .env config file has been parsed incorrectly.
type ParseError struct {
	Line string
	Err  error
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse line: %v: %v", e.Line, e.Err)
}

func (e *ParseError) Unwrap() []error { return []error{e.Err, ErrParse} }

// FileReadError occurs when an error occurs when scanning the .env file.
type FileReadError struct {
	Filepath string
	Err      error
}

func (e *FileReadError) Error() string {
	return fmt.Sprintf("reading %v: %v", e.Filepath, e.Err)
}

func (e *FileReadError) Unwrap() []error { return []error{e.Err, ErrFile} }

// MalformedTagError occurs when a config struct tag is invalid.
type MalformedTagError struct {
	Tag string
	Err error
}

func (e *MalformedTagError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("malformed tag %q: %v", e.Tag, e.Err)
	}
	return fmt.Sprintf("malformed tag %q", e.Tag)
}

func (e *MalformedTagError) Unwrap() []error { return []error{e.Err, ErrTag} }

// MalformedDefaultError occurs when the default value in a struct tag cannot be
// parsed into the field's type. This is a developer error in the tag definition.
type MalformedDefaultError struct {
	FieldName string
	Default   string
	Err       error
}

func (e *MalformedDefaultError) Error() string {
	return fmt.Sprintf("default value %q is invalid for field %q: %v", e.Default, e.FieldName, e.Err)
}

func (e *MalformedDefaultError) Unwrap() []error { return []error{e.Err, ErrTag} }
