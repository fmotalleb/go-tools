package decoder

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
)

type Decodable interface {
	Decode(reflect.Type, interface{}) (any, error)
}

func Of(from any) (Decodable, error) {
	var to any
	opt, ok := any(&to).(Decodable)
	if !ok {
		return opt, nil
	}
	if _, err := opt.Decode(reflect.TypeOf(from), from); err != nil {
		return opt, err
	}
	return opt, nil
}

func DecodeHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, val interface{}) (interface{}, error) {
		opt, ok := reflect.New(to).Interface().(Decodable)
		if !ok {
			return val, nil
		}
		return opt.Decode(from, val)
	}
}
