package copier

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	TimestampToTime = TypeConverter{
		DstType: timestamppb.Timestamp{},
		SrcType: time.Time{},
		Fn: func(src interface{}) (interface{}, error) {
			return *timestamppb.New(src.(time.Time)), nil
		},
	}

	TimeToTimestamp = TypeConverter{
		DstType: time.Time{},
		SrcType: timestamppb.Timestamp{},
		Fn: func(src interface{}) (interface{}, error) {
			return src.(*timestamppb.Timestamp).AsTime(), nil
		},
	}
)

var (
	DefaultProtoConverters = []TypeConverter{
		TimestampToTime,
		TimeToTimestamp,
	}
)

func WithOptionsProtobuf(options ...Option) Option {
	option := Option{}
	option.Converters = DefaultProtoConverters

	if len(options) > 0 {
		option.DeepCopy = options[0].DeepCopy
		option.IgnoreEmpty = options[0].IgnoreEmpty
		option.FieldNameMapping = options[0].FieldNameMapping
		option.Converters = append(option.Converters, options[0].Converters...)
	}

	return option
}
