package structo

import (
	"reflect"
	"strconv"
)

// Actions represent the type of change
type Actions string

const (
	Add    Actions = "add"
	Remove Actions = "remove"
	Change Actions = "change"
)

// resultData stores from/to data
type resultData struct {
	dataFrom any
	dataTo   any
}

func (r *resultData) GetData() any {
	if r.dataFrom != nil && r.dataTo != nil {
		return []any{r.dataFrom, r.dataTo}
	} else if r.dataFrom != nil {
		return r.dataFrom
	}
	return r.dataTo
}

// result is the main output for each changed path
type result struct {
	Action Actions
	Data   resultData
}

// TrackWithHistory returns a map of field paths and their detected actions
func TrackWithHistory(oldStruct, newStruct interface{}) (map[string]result, error) {
	oldVal, newVal, err := getComparableValues(oldStruct, newStruct)
	if err != nil {
		return nil, err
	}

	changes := make(map[string]result)
	trackRecursive(oldVal, newVal, "", changes)
	return changes, nil
}

func trackRecursive(oldVal, newVal reflect.Value, prefix string, resultMap map[string]result) {
	if oldVal.Kind() == reflect.Ptr {
		if oldVal.IsNil() || newVal.IsNil() {
			if !reflect.DeepEqual(deref(oldVal), deref(newVal)) {
				resultMap[prefix] = result{
					Action: Change,
					Data:   resultData{dataFrom: derefInterface(oldVal), dataTo: derefInterface(newVal)},
				}
			}
			return
		}
		oldVal = oldVal.Elem()
		newVal = newVal.Elem()
	}

	switch oldVal.Kind() {
	case reflect.Struct:
		for i := 0; i < oldVal.NumField(); i++ {
			field := oldVal.Type().Field(i)
			if !field.IsExported() {
				continue
			}
			fieldPath := joinKey(prefix, field.Name)
			trackRecursive(oldVal.Field(i), newVal.Field(i), fieldPath, resultMap)
		}

	case reflect.Slice:
		oldLen := oldVal.Len()
		newLen := newVal.Len()
		minLen := oldLen
		if newLen < oldLen {
			minLen = newLen
		}

		for i := 0; i < minLen; i++ {
			oldItem := oldVal.Index(i)
			newItem := newVal.Index(i)
			if !reflect.DeepEqual(oldItem.Interface(), newItem.Interface()) {
				itemPath := joinKey(prefix, "["+strconv.Itoa(i)+"]")
				if isStructLike(oldItem) {
					trackRecursive(oldItem, newItem, itemPath, resultMap)
				} else {
					resultMap[itemPath] = result{
						Action: Change,
						Data: resultData{
							dataFrom: oldItem.Interface(),
							dataTo:   newItem.Interface(),
						},
					}
				}
			}
		}

		for i := minLen; i < newLen; i++ {
			itemPath := joinKey(prefix, "["+strconv.Itoa(i)+"]")
			resultMap[itemPath] = result{
				Action: Add,
				Data:   resultData{dataTo: newVal.Index(i).Interface()},
			}
		}

		for i := minLen; i < oldLen; i++ {
			itemPath := joinKey(prefix, "["+strconv.Itoa(i)+"]")
			resultMap[itemPath] = result{
				Action: Remove,
				Data:   resultData{dataFrom: oldVal.Index(i).Interface()},
			}
		}

	default:
		if !reflect.DeepEqual(oldVal.Interface(), newVal.Interface()) {
			resultMap[prefix] = result{
				Action: Change,
				Data: resultData{
					dataFrom: oldVal.Interface(),
					dataTo:   newVal.Interface(),
				},
			}
		}
	}
}

func deref(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		return v.Elem()
	}
	return v
}

func derefInterface(v reflect.Value) interface{} {
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		return v.Elem().Interface()
	}
	if v.IsValid() {
		return v.Interface()
	}
	return nil
}



func isStructLike(v reflect.Value) bool {
	return (v.Kind() == reflect.Struct) || (v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct)
}
