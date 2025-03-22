package structo

import "github.com/Lucifer07/Structo/copier"

// Copy copy things
func Copy(toValue interface{}, fromValue interface{}) (err error) {
	return copier.Copier(toValue, fromValue, copier.Option{
		Converters: copier.DefaultProtoConverters,
	})
}

// CopyWithOption copy with option
func CopyWithOption(toValue interface{}, fromValue interface{}, opt copier.Option) (err error) {
	return copier.Copier(toValue, fromValue, opt)
}


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


