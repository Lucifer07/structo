package structo

import (
	"reflect"

	"github.com/Lucifer07/Structo/errdefs"
)

// Diff returns a map of field names and their [old, new] values that differ.
func Diff(oldStruct, newStruct interface{}) (map[string][2]interface{}, error) {
	oldVal, newVal, err := getComparableValues(oldStruct, newStruct)
	if err != nil {
		return nil, err
	}

	differences := make(map[string][2]interface{})
	for i := 0; i < oldVal.NumField(); i++ {
		field := oldVal.Type().Field(i).Name
		oldValue := oldVal.Field(i).Interface()
		newValue := newVal.Field(i).Interface()

		if !reflect.DeepEqual(oldValue, newValue) {
			differences[field] = [2]interface{}{oldValue, newValue}
		}
	}
	return differences, nil
}

// getComparableValues returns reflect.Value of two structs after dereferencing pointers.
func getComparableValues(a, b interface{}) (reflect.Value, reflect.Value, error) {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	if va.Kind() == reflect.Ptr {
		va = va.Elem()
	}
	if vb.Kind() == reflect.Ptr {
		vb = vb.Elem()
	}
	if va.Kind() != reflect.Struct || vb.Kind() != reflect.Struct {
		return reflect.Value{}, reflect.Value{}, errdefs.ErrInvalidStructType
	}
	if va.Type() != vb.Type() {
		return reflect.Value{}, reflect.Value{}, errdefs.ErrMismatchedStructTypes
	}
	return va, vb, nil
}
