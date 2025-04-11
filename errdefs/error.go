package errdefs

import "errors"

var (
	ErrInvalidCopyDestination        = errors.New("copy destination must be non-nil and addressable")
	ErrInvalidCopyFrom               = errors.New("copy from must be non-nil and addressable")
	ErrMapKeyNotMatch                = errors.New("map's key type doesn't match")
	ErrNotSupported                  = errors.New("not supported")
	ErrFieldNameTagStartNotUpperCase = errors.New("copier field name tag must be start upper case")
	ErrInvalidStructType             = errors.New("provided value is not a struct or pointer to struct")
	ErrUnsupportedKind               = errors.New("unsupported reflect.Kind encountered")
	ErrFieldNotSettable              = errors.New("cannot set value to field")
	ErrMismatchedStructTypes         = errors.New("structs must be of the same type")
	ErrNotPointerToStruct            = errors.New("input must be a pointer to a struct")
)
