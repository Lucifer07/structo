package structo

import "reflect"

// Option sets copy options
type CopyOption struct {
	// setting this value to true will ignore copying zero values of all the fields, including bools, as well as a
	// struct having all it's fields set to their zero values respectively (see IsZero() in reflect/value.go)
	IgnoreEmpty   bool
	CaseSensitive bool
	DeepCopy      bool
	Converters    []TypeConverter
	// Custom field name mappings to copy values with different names in `fromValue` and `toValue` types.
	// Examples can be found in `copier_field_name_mapping_test.go`.
	FieldNameMapping []FieldNameMapping
}

func (opt CopyOption) converters() map[converterPair]TypeConverter {
	var converters = map[converterPair]TypeConverter{}

	// save converters into map for faster lookup
	for i := range opt.Converters {
		pair := converterPair{
			SrcType: reflect.TypeOf(opt.Converters[i].SrcType),
			DstType: reflect.TypeOf(opt.Converters[i].DstType),
		}

		converters[pair] = opt.Converters[i]
	}

	return converters
}

func (opt CopyOption) fieldNameMapping() map[converterPair]FieldNameMapping {
	var mapping = map[converterPair]FieldNameMapping{}

	for i := range opt.FieldNameMapping {
		pair := converterPair{
			SrcType: reflect.TypeOf(opt.FieldNameMapping[i].SrcType),
			DstType: reflect.TypeOf(opt.FieldNameMapping[i].DstType),
		}

		mapping[pair] = opt.FieldNameMapping[i]
	}

	return mapping
}

func WithOptionsProtobuf(options ...CopyOption) CopyOption {
	option := CopyOption{}
	option.Converters = DefaultProtoConverters

	if len(options) > 0 {
		option.DeepCopy = options[0].DeepCopy
		option.IgnoreEmpty = options[0].IgnoreEmpty
		option.FieldNameMapping = options[0].FieldNameMapping
		option.Converters = append(option.Converters, options[0].Converters...)
	}

	return option
}
