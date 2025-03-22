/*
Package structo provides utility functions for copying data between structures.

It wraps the copier package to provide convenient functions for copying
with default and customizable options.
*/
package structo

import "github.com/Lucifer07/Structo/copier"

// Copy copies data from fromValue to toValue using default options.
//
// Parameters:
//   - toValue: The destination object where data will be copied.
//   - fromValue: The source object from which data is copied.
//
// Returns:
//   - error: Returns an error if the copying process fails.
func Copy(toValue interface{}, fromValue interface{}) (err error) {
	return copier.Copier(toValue, fromValue, copier.Option{
		Converters: copier.DefaultProtoConverters,
	})
}

// CopyWithOption copies data from fromValue to toValue using custom options.
//
// Parameters:
//   - toValue: The destination object where data will be copied.
//   - fromValue: The source object from which data is copied.
//   - opt: Custom copier options to modify the copying behavior.
//
// Returns:
//   - error: Returns an error if the copying process fails.
func CopyWithOption(toValue interface{}, fromValue interface{}, opt copier.Option) (err error) {
	return copier.Copier(toValue, fromValue, opt)
}

// CreateProtobufOption creates a copier option with default protobuf converters.
// Additional options can be provided to override specific behaviors.
//
// Parameters:
//   - options: (Optional) Additional copier options to be merged.
//
// Returns:
//   - copier.Option: A copier option struct configured with protobuf converters.
func CreateProtobufOption(options ...copier.Option) copier.Option {
	option := copier.Option{}
	option.Converters = copier.DefaultProtoConverters

	if len(options) > 0 {
		option.DeepCopy = options[0].DeepCopy
		option.IgnoreEmpty = options[0].IgnoreEmpty
		option.FieldNameMapping = options[0].FieldNameMapping
		option.Converters = append(option.Converters, options[0].Converters...)
	}

	return option
}
