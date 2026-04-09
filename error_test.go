package configutil_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/h-dav/configutil"
)

func TestErrorsIs(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
	}{
		{"FileTypeValidationError is ErrFile", &configutil.FileTypeValidationError{Filepath: "bad.txt"}, configutil.ErrFile},
		{"OpenFileError is ErrFile", &configutil.OpenFileError{Err: fmt.Errorf("open failed")}, configutil.ErrFile},
		{"FieldConversionError is ErrConversion", &configutil.FieldConversionError{FieldName: "Port", TargetType: "int", Err: fmt.Errorf("bad")}, configutil.ErrConversion},
		{"UnsupportedFieldTypeError is ErrUnsupported", &configutil.UnsupportedFieldTypeError{FieldType: "complex128"}, configutil.ErrUnsupported},
		{"InvalidConfigTypeError is ErrInvalidConfig", &configutil.InvalidConfigTypeError{ProvidedType: "string"}, configutil.ErrInvalidConfig},
		{"RequiredFieldError is ErrRequired", &configutil.RequiredFieldError{FieldName: "Name"}, configutil.ErrRequired},
		{"ReplacementError is ErrReplacement", &configutil.ReplacementError{VariableName: "HOST"}, configutil.ErrReplacement},
		{"ParseError is ErrParse", &configutil.ParseError{Line: "bad", Err: configutil.ErrSyntax}, configutil.ErrParse},
		{"FileReadError is ErrFile", &configutil.FileReadError{Filepath: "test.env", Err: fmt.Errorf("read")}, configutil.ErrFile},
		{"MalformedTagError is ErrTag", &configutil.MalformedTagError{Tag: "bad"}, configutil.ErrTag},
		{"MalformedDefaultError is ErrTag", &configutil.MalformedDefaultError{FieldName: "Port", Default: "abc", Err: fmt.Errorf("bad")}, configutil.ErrTag},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !errors.Is(tc.err, tc.target) {
				t.Errorf("errors.Is(%T, %v) = false, want true", tc.err, tc.target)
			}
		})
	}
}

func TestErrorsIs_Negative(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
	}{
		{"RequiredFieldError is not ErrFile", &configutil.RequiredFieldError{FieldName: "X"}, configutil.ErrFile},
		{"FileTypeValidationError is not ErrRequired", &configutil.FileTypeValidationError{Filepath: "x"}, configutil.ErrRequired},
		{"FieldConversionError is not ErrParse", &configutil.FieldConversionError{FieldName: "X", TargetType: "int", Err: fmt.Errorf("x")}, configutil.ErrParse},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if errors.Is(tc.err, tc.target) {
				t.Errorf("errors.Is(%T, %v) = true, want false", tc.err, tc.target)
			}
		})
	}
}

func TestErrorsUnwrap(t *testing.T) {
	inner := fmt.Errorf("inner error")

	tests := []struct {
		name string
		err  error
	}{
		{"OpenFileError", &configutil.OpenFileError{Err: inner}},
		{"FieldConversionError", &configutil.FieldConversionError{FieldName: "X", TargetType: "int", Err: inner}},
		{"ParseError", &configutil.ParseError{Line: "bad", Err: inner}},
		{"FileReadError", &configutil.FileReadError{Filepath: "x.env", Err: inner}},
		{"MalformedTagError", &configutil.MalformedTagError{Tag: "bad", Err: inner}},
		{"FieldError", &configutil.FieldError{FieldName: "X", Err: inner}},
		{"MalformedDefaultError", &configutil.MalformedDefaultError{FieldName: "X", Default: "bad", Err: inner}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			unwrapped := errors.Unwrap(tc.err)
			if unwrapped != inner {
				t.Errorf("Unwrap() = %v, want %v", unwrapped, inner)
			}
		})
	}
}

func TestErrorsAs_ThroughFieldError(t *testing.T) {
	inner := &configutil.RequiredFieldError{FieldName: "Name"}
	wrapped := &configutil.FieldError{FieldName: "Name", Err: inner}

	var target *configutil.RequiredFieldError
	if !errors.As(wrapped, &target) {
		t.Error("errors.As through FieldError failed to find RequiredFieldError")
	}
	if target.FieldName != "Name" {
		t.Errorf("FieldName = %q, want %q", target.FieldName, "Name")
	}

	// FieldError wrapping RequiredFieldError should match ErrRequired via the chain.
	if !errors.Is(wrapped, configutil.ErrRequired) {
		t.Error("errors.Is(FieldError{RequiredFieldError}, ErrRequired) = false, want true")
	}
}

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"FileTypeValidationError", &configutil.FileTypeValidationError{Filepath: "bad.txt"}, `file extension is not a valid environment file: "bad.txt"`},
		{"OpenFileError", &configutil.OpenFileError{Err: fmt.Errorf("no such file")}, "failed to open config file: no such file"},
		{"UnsupportedFieldTypeError", &configutil.UnsupportedFieldTypeError{FieldType: "test"}, "unsupported field type: string"},
		{"InvalidConfigTypeError", &configutil.InvalidConfigTypeError{ProvidedType: "test"}, "output must be a pointer to a struct, got string"},
		{"RequiredFieldError", &configutil.RequiredFieldError{FieldName: "Name"}, "required field is not set in configuration: Name"},
		{"ReplacementError", &configutil.ReplacementError{VariableName: "HOST"}, "configuration variable for replacement is not set: HOST"},
		{"MalformedTagError with err", &configutil.MalformedTagError{Tag: "bad", Err: fmt.Errorf("reason")}, `malformed tag "bad": reason`},
		{"MalformedTagError without err", &configutil.MalformedTagError{Tag: "bad"}, `malformed tag "bad"`},
		{"FieldError with name", &configutil.FieldError{FieldName: "X", Err: fmt.Errorf("fail")}, `field "X": fail`},
		{"FieldError without name", &configutil.FieldError{Err: fmt.Errorf("fail")}, "fail"},
		{"MalformedDefaultError", &configutil.MalformedDefaultError{FieldName: "Port", Default: "abc", Err: fmt.Errorf("bad")}, `default value "abc" is invalid for field "Port": bad`},
		{"FieldConversionError", &configutil.FieldConversionError{FieldName: "Port", TargetType: "int", Err: fmt.Errorf("bad")}, "failed to convert field Port to int: bad"},
		{"ParseError", &configutil.ParseError{Line: "BADLINE", Err: configutil.ErrSyntax}, "parse line: BADLINE: invalid syntax"},
		{"FileReadError", &configutil.FileReadError{Filepath: "x.env", Err: fmt.Errorf("read err")}, "reading x.env: read err"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.want {
				t.Errorf("Error() = %q, want %q", got, tc.want)
			}
		})
	}
}
