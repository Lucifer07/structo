package structo

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Lucifer07/Structo/errdefs"
)

// Flatten returns a flat map of a struct's fields using dot notation.
func Flatten(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	flattenHelper(reflect.ValueOf(data), "", result)
	return result
}

// Unflatten sets values in a struct based on keys in a flat map (supports nested dot notation).
func Unflatten(flatMap map[string]interface{}, result interface{}) error {
	v := reflect.ValueOf(result)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errdefs.ErrInvalidStructType
	}
	v = v.Elem()

	for key, val := range flatMap {
		if err := setNestedField(v, key, val); err != nil {
			return err
		}
	}
	return nil
}

// setNestedField sets a value in a nested struct based on a dot-notated key.
func setNestedField(v reflect.Value, key string, val interface{}) error {
	parts := strings.Split(key, ".")
	for i, part := range parts {
		// Dereference pointer
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}

		// Handle final value
		if i == len(parts)-1 {
			// Slice index case
			if v.Kind() == reflect.Slice {
				index, err := strconv.Atoi(part)
				if err != nil {
					return fmt.Errorf("invalid slice index: %q", part)
				}
				// Extend slice if necessary
				if v.Len() <= index {
					newSlice := reflect.MakeSlice(v.Type(), index+1, index+1)
					reflect.Copy(newSlice, v)
					v.Set(newSlice)
				}
				elem := v.Index(index)
				return setValue(elem, val, key)
			}

			// Struct field case
			if v.Kind() != reflect.Struct {
				return fmt.Errorf("cannot set field %q: not a struct", strings.Join(parts[:i+1], "."))
			}
			field := v.FieldByName(part)
			if !field.IsValid() {
				return fmt.Errorf("field %q not found", strings.Join(parts[:i+1], "."))
			}
			return setValue(field, val, key)
		}

		// Traverse deeper: slice index or struct field
		if v.Kind() == reflect.Slice {
			index, err := strconv.Atoi(part)
			if err != nil {
				return fmt.Errorf("invalid slice index: %q", part)
			}
			if v.Len() <= index {
				newSlice := reflect.MakeSlice(v.Type(), index+1, index+1)
				reflect.Copy(newSlice, v)
				v.Set(newSlice)
			}
			v = v.Index(index)
			continue
		}

		if v.Kind() != reflect.Struct {
			return fmt.Errorf("cannot traverse into %q: not a struct", strings.Join(parts[:i+1], "."))
		}

		field := v.FieldByName(part)
		if !field.IsValid() {
			return fmt.Errorf("field %q not found", strings.Join(parts[:i+1], "."))
		}
		v = field
	}
	return nil
}

func setValue(field reflect.Value, val interface{}, key string) error {
	fv := reflect.ValueOf(val)

	// Handle if field is a pointer type
	if field.Kind() == reflect.Ptr {
		if fv.Type().AssignableTo(field.Type().Elem()) {
			ptr := reflect.New(field.Type().Elem())
			ptr.Elem().Set(fv)
			field.Set(ptr)
			return nil
		} else if fv.Type().ConvertibleTo(field.Type().Elem()) {
			ptr := reflect.New(field.Type().Elem())
			ptr.Elem().Set(fv.Convert(field.Type().Elem()))
			field.Set(ptr)
			return nil
		}
	}

	// Normal assignment fallback
	if fv.Type().AssignableTo(field.Type()) {
		field.Set(fv)
	} else if fv.Type().ConvertibleTo(field.Type()) {
		field.Set(fv.Convert(field.Type()))
	} else {
		return fmt.Errorf("cannot assign value of type %s to field %q (type %s)",
			fv.Type(), key, field.Type())
	}
	return nil
}


// flattenHelper recursively flattens structs and arrays/slices.
func flattenHelper(v reflect.Value, prefix string, result map[string]interface{}) {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			fieldName := field.Name
			flattenHelper(v.Field(i), joinKey(prefix, fieldName), result)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			flattenHelper(v.Index(i), joinKey(prefix, fmt.Sprintf("%d", i)), result)
		}
	default:
		if prefix != "" {
			result[prefix] = v.Interface()
		}
	}
}



// joinKey combines parent and child keys using dot notation.
func joinKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}
