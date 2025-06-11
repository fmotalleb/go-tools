package decoder

import (
	"reflect"

	"github.com/mitchellh/mapstructure"
)

func DecodeHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, val interface{}) (interface{}, error) {
		opt, ok := reflect.New(to).Interface().(Decodable)
		if !ok {
			return val, nil
		}
		if err := opt.Decode(from, to, val); err != nil {
			return nil, err
		}
		return reflect.ValueOf(opt).Elem().Interface(), nil
	}
}
