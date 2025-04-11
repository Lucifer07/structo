package structo

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type converterPair struct {
	SrcType reflect.Type
	DstType reflect.Type
}

type FieldNameMapping struct {
	SrcType interface{}
	DstType interface{}
	Mapping map[string]string
}

func getFieldNamesMapping(mappings map[converterPair]FieldNameMapping, fromType reflect.Type, toType reflect.Type) map[string]string {
	var fieldNamesMapping map[string]string

	if len(mappings) > 0 {
		pair := converterPair{
			SrcType: fromType,
			DstType: toType,
		}
		if v, ok := mappings[pair]; ok {
			fieldNamesMapping = v.Mapping
		}
	}
	return fieldNamesMapping
}

func fieldByNameOrZeroValue(source reflect.Value, fieldName string) (value reflect.Value) {
	defer func() {
		if err := recover(); err != nil {
			value = reflect.Value{}
		}
	}()

	return source.FieldByName(fieldName)
}

func copyUnexportedStructFields(to, from reflect.Value) {
	if from.Kind() != reflect.Struct || to.Kind() != reflect.Struct || !from.Type().AssignableTo(to.Type()) {
		return
	}

	// create a shallow copy of 'to' to get all fields
	tmp := indirect(reflect.New(to.Type()))
	tmp.Set(from)

	// revert exported fields
	for i := 0; i < to.NumField(); i++ {
		if tmp.Field(i).CanSet() {
			tmp.Field(i).Set(to.Field(i))
		}
	}
	to.Set(tmp)
}

func shouldIgnore(v reflect.Value, ignoreEmpty bool) bool {
	return ignoreEmpty && v.IsZero()
}

var deepFieldsLock sync.RWMutex
var deepFieldsMap = make(map[reflect.Type][]reflect.StructField)

func deepFields(reflectType reflect.Type) []reflect.StructField {
	deepFieldsLock.RLock()
	cache, ok := deepFieldsMap[reflectType]
	deepFieldsLock.RUnlock()
	if ok {
		return cache
	}
	var res []reflect.StructField
	if reflectType, _ = indirectType(reflectType); reflectType.Kind() == reflect.Struct {
		fields := make([]reflect.StructField, 0, reflectType.NumField())

		for i := 0; i < reflectType.NumField(); i++ {
			v := reflectType.Field(i)
			// PkgPath is the package path that qualifies a lower case (unexported)
			// field name. It is empty for upper case (exported) field names.
			// See https://golang.org/ref/spec#Uniqueness_of_identifiers
			if v.PkgPath == "" {
				fields = append(fields, v)
				if v.Anonymous {
					// also consider fields of anonymous fields as fields of the root
					fields = append(fields, deepFields(v.Type)...)
				}
			}
		}
		res = fields
	}

	deepFieldsLock.Lock()
	deepFieldsMap[reflectType] = res
	deepFieldsLock.Unlock()
	return res
}

func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

func indirectType(reflectType reflect.Type) (_ reflect.Type, isPtr bool) {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
		isPtr = true
	}
	return reflectType, isPtr
}

func set(to, from reflect.Value, deepCopy bool, converters map[converterPair]TypeConverter) (bool, error) {
	if !from.IsValid() {
		return true, nil
	}
	if ok, err := lookupAndCopyWithConverter(to, from, converters); err != nil {
		return false, err
	} else if ok {
		return true, nil
	}

	if to.Kind() == reflect.Ptr {
		// set `to` to nil if from is nil
		if from.Kind() == reflect.Ptr && from.IsNil() {
			to.Set(reflect.Zero(to.Type()))
			return true, nil
		} else if to.IsNil() {
			// `from`         -> `to`
			// sql.NullString -> *string
			if fromValuer, ok := driverValuer(from); ok {
				v, err := fromValuer.Value()
				if err != nil {
					return true, nil
				}
				// if `from` is not valid do nothing with `to`
				if v == nil {
					return true, nil
				}
			}
			// allocate new `to` variable with default value (eg. *string -> new(string))
			to.Set(reflect.New(to.Type().Elem()))
		}
		// depointer `to`
		to = to.Elem()
	}

	if deepCopy {
		toKind := to.Kind()
		if toKind == reflect.Interface && to.IsNil() {
			if reflect.TypeOf(from.Interface()) != nil {
				to.Set(reflect.New(reflect.TypeOf(from.Interface())).Elem())
				toKind = reflect.TypeOf(to.Interface()).Kind()
			}
		}
		if from.Kind() == reflect.Ptr && from.IsNil() {
			return true, nil
		}
		if _, ok := to.Addr().Interface().(sql.Scanner); !ok && (toKind == reflect.Struct || toKind == reflect.Map || toKind == reflect.Slice) {
			return false, nil
		}
	}

	if from.Type().ConvertibleTo(to.Type()) {
		to.Set(from.Convert(to.Type()))
	} else if toScanner, ok := to.Addr().Interface().(sql.Scanner); ok {
		// `from`  -> `to`
		// *string -> sql.NullString
		if from.Kind() == reflect.Ptr {
			// if `from` is nil do nothing with `to`
			if from.IsNil() {
				return true, nil
			}
			// depointer `from`
			from = indirect(from)
		}
		// `from` -> `to`
		// string -> sql.NullString
		// set `to` by invoking method Scan(`from`)
		err := toScanner.Scan(from.Interface())
		if err != nil {
			return false, nil
		}
	} else if fromValuer, ok := driverValuer(from); ok {
		// `from`         -> `to`
		// sql.NullString -> string
		v, err := fromValuer.Value()
		if err != nil {
			return false, nil
		}
		// if `from` is not valid do nothing with `to`
		if v == nil {
			return true, nil
		}
		rv := reflect.ValueOf(v)
		if rv.Type().AssignableTo(to.Type()) {
			to.Set(rv)
		} else if to.CanSet() && rv.Type().ConvertibleTo(to.Type()) {
			to.Set(rv.Convert(to.Type()))
		}
	} else if from.Kind() == reflect.Ptr {
		return set(to, from.Elem(), deepCopy, converters)
	} else {
		return false, nil
	}

	return true, nil
}

// lookupAndCopyWithConverter looks up the type pair, on success the TypeConverter Fn func is called to copy src to dst field.
func lookupAndCopyWithConverter(to, from reflect.Value, converters map[converterPair]TypeConverter) (copied bool, err error) {
	pair := converterPair{
		SrcType: from.Type(),
		DstType: to.Type(),
	}

	if cnv, ok := converters[pair]; ok {
		result, err := cnv.Fn(from.Interface())
		if err != nil {
			return false, err
		}

		if result != nil {
			to.Set(reflect.ValueOf(result))
		} else {
			// in case we've got a nil value to copy
			to.Set(reflect.Zero(to.Type()))
		}

		return true, nil
	}

	return false, nil
}

// checkBitFlags Checks flags for error or panic conditions.
func checkBitFlags(flagsList map[string]uint8) (err error) {
	// Check flag conditions were met
	for name, flgs := range flagsList {
		if flgs&hasCopied == 0 {
			switch {
			case flgs&tagMust != 0 && flgs&tagNoPanic != 0:
				err = fmt.Errorf("field %s has must tag but was not copied", name)
				return
			case flgs&(tagMust) != 0:
				panic(fmt.Sprintf("Field %s has must tag but was not copied", name))
			}
		}
	}
	return
}

func getFieldName(fieldName string, flgs flags, fieldNameMapping map[string]string) (srcFieldName string, destFieldName string) {
	// get dest field name
	if name, ok := fieldNameMapping[fieldName]; ok {
		srcFieldName = fieldName
		destFieldName = name
		return
	}

	if srcTagName, ok := flgs.SrcNames.FieldNameToTag[fieldName]; ok {
		destFieldName = srcTagName
		if destTagName, ok := flgs.DestNames.TagToFieldName[srcTagName]; ok {
			destFieldName = destTagName
		}
	} else {
		if destTagName, ok := flgs.DestNames.TagToFieldName[fieldName]; ok {
			destFieldName = destTagName
		}
	}
	if destFieldName == "" {
		destFieldName = fieldName
	}

	// get source field name
	if destTagName, ok := flgs.DestNames.FieldNameToTag[fieldName]; ok {
		srcFieldName = destTagName
		if srcField, ok := flgs.SrcNames.TagToFieldName[destTagName]; ok {
			srcFieldName = srcField
		}
	} else {
		if srcField, ok := flgs.SrcNames.TagToFieldName[fieldName]; ok {
			srcFieldName = srcField
		}
	}

	if srcFieldName == "" {
		srcFieldName = fieldName
	}
	return
}

func driverValuer(v reflect.Value) (i driver.Valuer, ok bool) {
	if !v.CanAddr() {
		i, ok = v.Interface().(driver.Valuer)
		return
	}

	i, ok = v.Addr().Interface().(driver.Valuer)
	return
}

func fieldByName(v reflect.Value, name string, caseSensitive bool) reflect.Value {
	if caseSensitive {
		return v.FieldByName(name)
	}

	return v.FieldByNameFunc(func(n string) bool { return strings.EqualFold(n, name) })
}
