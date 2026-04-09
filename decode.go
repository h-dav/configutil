package configutil

import (
	"reflect"
	"strconv"
	"strings"
)

// setFieldValue decodes entry.value into the target field using kind-based dispatch.
// This supports named types (e.g. type Port int, type Env string) in addition to
// primitive types, and handles all integer and float bit sizes.
func (s *settings) setFieldValue(field reflect.Value, e entry) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(e.value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(e.value, 10, field.Type().Bits())
		if err != nil {
			return &FieldConversionError{FieldName: e.key, TargetType: field.Type().String(), Err: err}
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(e.value, 10, field.Type().Bits())
		if err != nil {
			return &FieldConversionError{FieldName: e.key, TargetType: field.Type().String(), Err: err}
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(e.value, field.Type().Bits())
		if err != nil {
			return &FieldConversionError{FieldName: e.key, TargetType: field.Type().String(), Err: err}
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(e.value)
		if err != nil {
			return &FieldConversionError{FieldName: e.key, TargetType: "bool", Err: err}
		}
		field.SetBool(b)
	case reflect.Slice:
		return setSliceField(field, e)
	default:
		return &UnsupportedFieldTypeError{FieldType: field.Interface()}
	}
	return nil
}

func setSliceField(field reflect.Value, e entry) error {
	switch field.Type().Elem().Kind() {
	case reflect.String:
		field.Set(splitStringSlice(e.value, field.Type()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := splitIntSlice(e.key, e.value, field.Type())
		if err != nil {
			return err
		}
		field.Set(v)
	case reflect.Float32, reflect.Float64:
		v, err := splitFloatSlice(e.key, e.value, field.Type())
		if err != nil {
			return err
		}
		field.Set(v)
	default:
		return &UnsupportedFieldTypeError{FieldType: field.Interface()}
	}
	return nil
}

func splitStringSlice(value string, sliceType reflect.Type) reflect.Value {
	parts := strings.Split(value, ",")
	result := reflect.MakeSlice(sliceType, len(parts), len(parts))
	for i, v := range parts {
		result.Index(i).SetString(strings.TrimSpace(v))
	}
	return result
}

func splitIntSlice(key, value string, sliceType reflect.Type) (reflect.Value, error) {
	parts := strings.Split(value, ",")
	result := reflect.MakeSlice(sliceType, len(parts), len(parts))
	bits := sliceType.Elem().Bits()
	for i, v := range parts {
		n, err := strconv.ParseInt(strings.TrimSpace(v), 10, bits)
		if err != nil {
			return reflect.Value{}, &FieldConversionError{FieldName: key, TargetType: sliceType.String(), Err: err}
		}
		result.Index(i).SetInt(n)
	}
	return result, nil
}

func splitFloatSlice(key, value string, sliceType reflect.Type) (reflect.Value, error) {
	parts := strings.Split(value, ",")
	result := reflect.MakeSlice(sliceType, len(parts), len(parts))
	bits := sliceType.Elem().Bits()
	for i, v := range parts {
		f, err := strconv.ParseFloat(strings.TrimSpace(v), bits)
		if err != nil {
			return reflect.Value{}, &FieldConversionError{FieldName: key, TargetType: sliceType.String(), Err: err}
		}
		result.Index(i).SetFloat(f)
	}
	return result, nil
}
