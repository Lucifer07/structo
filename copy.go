package structo

import (
	"fmt"
	"reflect"

	"github.com/Lucifer07/Structo/errdefs"
)

// Copy things.
func Copy(toValue interface{}, fromValue interface{}) (err error) {
	return CopyWithOption(toValue, fromValue, CopyOption{
		Converters: DefaultProtoConverters,
	})
}

// Copy With Option
func CopyWithOption(toValue interface{}, fromValue interface{}, opt CopyOption) (err error) {
	var (
		isSlice    bool
		amount     = 1
		from       = indirect(reflect.ValueOf(fromValue))
		to         = indirect(reflect.ValueOf(toValue))
		converters = opt.converters()
		mappings   = opt.fieldNameMapping()
	)

	if !to.CanAddr() {
		return errdefs.ErrInvalidCopyDestination
	}

	// Return is from value is invalid
	if !from.IsValid() {
		return errdefs.ErrInvalidCopyFrom
	}

	fromType, isPtrFrom := indirectType(from.Type())
	toType, _ := indirectType(to.Type())

	if fromType.Kind() == reflect.Interface {
		fromType = reflect.TypeOf(from.Interface())
	}

	if toType.Kind() == reflect.Interface {
		toType, _ = indirectType(reflect.TypeOf(to.Interface()))
		oldTo := to
		to = reflect.New(reflect.TypeOf(to.Interface())).Elem()
		defer func() {
			oldTo.Set(to)
		}()
	}

	// Just set it if possible to assign for normal types
	if from.Kind() != reflect.Slice && from.Kind() != reflect.Struct && from.Kind() != reflect.Map && (from.Type().AssignableTo(to.Type()) || from.Type().ConvertibleTo(to.Type())) {
		if !isPtrFrom || !opt.DeepCopy {
			to.Set(from.Convert(to.Type()))
		} else {
			fromCopy := reflect.New(from.Type())
			fromCopy.Set(from.Elem())
			to.Set(fromCopy.Convert(to.Type()))
		}
		return
	}

	if from.Kind() != reflect.Slice && fromType.Kind() == reflect.Map && toType.Kind() == reflect.Map {
		if !fromType.Key().ConvertibleTo(toType.Key()) {
			return errdefs.ErrMapKeyNotMatch
		}

		if to.IsNil() {
			to.Set(reflect.MakeMapWithSize(toType, from.Len()))
		}

		for _, k := range from.MapKeys() {
			toKey := indirect(reflect.New(toType.Key()))
			isSet, err := set(toKey, k, opt.DeepCopy, converters)
			if err != nil {
				return err
			}
			if !isSet {
				return fmt.Errorf("%w map, old key: %v, new key: %v", errdefs.ErrNotSupported, k.Type(), toType.Key())
			}

			elemType := toType.Elem()
			if elemType.Kind() != reflect.Slice {
				elemType, _ = indirectType(elemType)
			}
			toValue := indirect(reflect.New(elemType))
			isSet, err = set(toValue, from.MapIndex(k), opt.DeepCopy, converters)
			if err != nil {
				return err
			}
			if !isSet {
				if err = CopyWithOption(toValue.Addr().Interface(), from.MapIndex(k).Interface(), opt); err != nil {
					return err
				}
			}

			for {
				if elemType == toType.Elem() {
					to.SetMapIndex(toKey, toValue)
					break
				}
				elemType = reflect.PointerTo(elemType)
				toValue = toValue.Addr()
			}
		}
		return
	}

	if from.Kind() == reflect.Slice && to.Kind() == reflect.Slice {
		if to.IsNil() {
			slice := reflect.MakeSlice(reflect.SliceOf(to.Type().Elem()), from.Len(), from.Cap())
			to.Set(slice)
		}
		if fromType.ConvertibleTo(toType) {
			for i := 0; i < from.Len(); i++ {
				if to.Len() < i+1 {
					to.Set(reflect.Append(to, reflect.New(to.Type().Elem()).Elem()))
				}
				isSet, err := set(to.Index(i), from.Index(i), opt.DeepCopy, converters)
				if err != nil {
					return err
				}
				if !isSet {
					// ignore error while copy slice element
					err = CopyWithOption(to.Index(i).Addr().Interface(), from.Index(i).Interface(), opt)
					if err != nil {
						continue
					}
				}
			}
			return
		}
	}

	if fromType.Kind() != reflect.Struct || toType.Kind() != reflect.Struct {
		// skip not supported type
		return
	}

	if len(converters) > 0 {
		if ok, e := set(to, from, opt.DeepCopy, converters); e == nil && ok {
			// converter supported
			return
		}
	}

	if from.Kind() == reflect.Slice || to.Kind() == reflect.Slice {
		isSlice = true
		if from.Kind() == reflect.Slice {
			amount = from.Len()
		}
	}

	for i := 0; i < amount; i++ {
		var dest, source reflect.Value

		if isSlice {
			// source
			if from.Kind() == reflect.Slice {
				source = indirect(from.Index(i))
			} else {
				source = indirect(from)
			}
			// dest
			dest = indirect(reflect.New(toType).Elem())
		} else {
			source = indirect(from)
			dest = indirect(to)
		}

		if len(converters) > 0 {
			if ok, e := set(dest, source, opt.DeepCopy, converters); e == nil && ok {
				if isSlice {
					// FIXME: maybe should check the other types?
					if to.Type().Elem().Kind() == reflect.Ptr {
						to.Index(i).Set(dest.Addr())
					} else {
						if to.Len() < i+1 {
							reflect.Append(to, dest)
						} else {
							to.Index(i).Set(dest)
						}
					}
				} else {
					to.Set(dest)
				}

				continue
			}
		}

		destKind := dest.Kind()
		initDest := false
		if destKind == reflect.Interface {
			initDest = true
			dest = indirect(reflect.New(toType))
		}

		// Get tag options
		flgs, err := getFlags(dest, source, toType, fromType)
		if err != nil {
			return err
		}

		// check source
		if source.IsValid() {
			copyUnexportedStructFields(dest, source)

			// Copy from source field to dest field or method
			fromTypeFields := deepFields(fromType)
			for _, field := range fromTypeFields {
				name := field.Name

				// Get bit flags for field
				fieldFlags := flgs.BitFlags[name]

				// Check if we should ignore copying
				if (fieldFlags & tagIgnore) != 0 {
					continue
				}

				fieldNamesMapping := getFieldNamesMapping(mappings, fromType, toType)

				srcFieldName, destFieldName := getFieldName(name, flgs, fieldNamesMapping)
				if fromField := fieldByNameOrZeroValue(source, srcFieldName); fromField.IsValid() && !shouldIgnore(fromField, opt.IgnoreEmpty) {
					// process for nested anonymous field
					destFieldNotSet := false
					if f, ok := dest.Type().FieldByName(destFieldName); ok {
						// only initialize parent embedded struct pointer in the path
						for idx := range f.Index[:len(f.Index)-1] {
							destField := dest.FieldByIndex(f.Index[:idx+1])

							if destField.Kind() != reflect.Ptr {
								continue
							}

							if !destField.IsNil() {
								continue
							}
							if !destField.CanSet() {
								destFieldNotSet = true
								break
							}

							// destField is a nil pointer that can be set
							newValue := reflect.New(destField.Type().Elem())
							destField.Set(newValue)
						}
					}

					if destFieldNotSet {
						break
					}

					toField := fieldByName(dest, destFieldName, opt.CaseSensitive)
					if toField.IsValid() {
						if toField.CanSet() {
							isSet, err := set(toField, fromField, opt.DeepCopy, converters)
							if err != nil {
								return err
							}
							if !isSet {
								if err := CopyWithOption(toField.Addr().Interface(), fromField.Interface(), opt); err != nil {
									return err
								}
							}
							if fieldFlags != 0 {
								// Note that a copy was made
								flgs.BitFlags[name] = fieldFlags | hasCopied
							}
						}
					} else {
						// try to set to method
						var toMethod reflect.Value
						if dest.CanAddr() {
							toMethod = dest.Addr().MethodByName(destFieldName)
						} else {
							toMethod = dest.MethodByName(destFieldName)
						}

						if toMethod.IsValid() && toMethod.Type().NumIn() == 1 && fromField.Type().AssignableTo(toMethod.Type().In(0)) {
							toMethod.Call([]reflect.Value{fromField})
						}
					}
				}
			}

			// Copy from from method to dest field
			for _, field := range deepFields(toType) {
				name := field.Name
				srcFieldName, destFieldName := getFieldName(name, flgs, getFieldNamesMapping(mappings, fromType, toType))

				var fromMethod reflect.Value
				if source.CanAddr() {
					fromMethod = source.Addr().MethodByName(srcFieldName)
				} else {
					fromMethod = source.MethodByName(srcFieldName)
				}

				if fromMethod.IsValid() && fromMethod.Type().NumIn() == 0 && fromMethod.Type().NumOut() == 1 && !shouldIgnore(fromMethod, opt.IgnoreEmpty) {
					if toField := fieldByName(dest, destFieldName, opt.CaseSensitive); toField.IsValid() && toField.CanSet() {
						values := fromMethod.Call([]reflect.Value{})
						if len(values) >= 1 {
							set(toField, values[0], opt.DeepCopy, converters)
						}
					}
				}
			}
		}

		if isSlice && to.Kind() == reflect.Slice {
			if dest.Addr().Type().AssignableTo(to.Type().Elem()) {
				if to.Len() < i+1 {
					to.Set(reflect.Append(to, dest.Addr()))
				} else {
					isSet, err := set(to.Index(i), dest.Addr(), opt.DeepCopy, converters)
					if err != nil {
						return err
					}
					if !isSet {
						// ignore error while copy slice element
						err = CopyWithOption(to.Index(i).Addr().Interface(), dest.Addr().Interface(), opt)
						if err != nil {
							continue
						}
					}
				}
			} else if dest.Type().AssignableTo(to.Type().Elem()) {
				if to.Len() < i+1 {
					to.Set(reflect.Append(to, dest))
				} else {
					isSet, err := set(to.Index(i), dest, opt.DeepCopy, converters)
					if err != nil {
						return err
					}
					if !isSet {
						// ignore error while copy slice element
						err = CopyWithOption(to.Index(i).Addr().Interface(), dest.Interface(), opt)
						if err != nil {
							continue
						}
					}
				}
			}
		} else if initDest {
			to.Set(dest)
		}

		// Enabled this if not need silent error 
		// err = checkBitFlags(flgs.BitFlags)
	}

	return
}
