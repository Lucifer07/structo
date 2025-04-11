package structo

import (
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type TypeConverter struct {
	SrcType interface{}
	DstType interface{}
	Fn      func(src interface{}) (dst interface{}, err error)
}

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

	DefaultProtoConverters = []TypeConverter{
		TimestampToTime,
		TimeToTimestamp,
	}
)
