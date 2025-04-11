package structo

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/Lucifer07/Structo/errdefs"
)

// InjectDefaults.
func InjectDefaults(ptr interface{}) error {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return errdefs.ErrUnsupportedKind
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return errdefs.ErrUnsupportedKind
	}

	return injectDefaultsRecursive(v)
}

func injectDefaultsRecursive(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		if !structField.IsExported() {
			continue
		}

		// Handle nested struct or pointer to struct
		switch field.Kind() {
		case reflect.Struct:
			err := injectDefaultsRecursive(field)
			if err != nil {
				return err
			}
		case reflect.Ptr:
			if field.IsNil() && structField.Type.Elem().Kind() == reflect.Struct {
				field.Set(reflect.New(structField.Type.Elem()))
			}
			if !field.IsNil() && field.Elem().Kind() == reflect.Struct {
				err := injectDefaultsRecursive(field.Elem())
				if err != nil {
					return err
				}
			}
		}

		// Inject default value if empty
		defaultVal := structField.Tag.Get("default")
		if defaultVal != "" && isZeroValue(field) {
			err := setFieldWithString(field, defaultVal)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
func setFieldWithString(field reflect.Value, val string) error {
	if field.Kind() == reflect.Ptr {
		elemType := field.Type().Elem()
		ptrValue := reflect.New(elemType).Elem()

		err := setFieldWithString(ptrValue, val)
		if err != nil {
			return err
		}

		field.Set(ptrValue.Addr())
		return nil
	}

	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.Slice:
		strs := strings.Split(val, ",")
		slice := reflect.MakeSlice(field.Type(), len(strs), len(strs))
		for i, s := range strs {
			item := slice.Index(i)
			err := setFieldWithString(item, strings.TrimSpace(s))
			if err != nil {
				return err
			}
		}
		field.Set(slice)
	case reflect.String:
		field.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		field.SetBool(b)
	}
	return nil
}
