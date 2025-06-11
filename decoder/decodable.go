package decoder

import (
	"reflect"
)

type Decodable interface {
	Decode(reflect.Type, reflect.Type, interface{}) error
}

func Of(from any) (Decodable, error) {
	var to any
	opt, ok := any(&to).(Decodable)
	if !ok {
		return opt, nil
	}
	if err := opt.Decode(reflect.TypeOf(from), reflect.TypeOf(to), from); err != nil {
		return opt, err
	}
	return opt, nil
}
