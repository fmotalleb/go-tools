package decoder

import "fmt"

type Parsable[T any] interface {
	Parse(val interface{}) (T, error)
}

func UsingParserOf[T any](from any) (*T, error) {
	var to T
	opt, ok := any(&to).(Parsable[T])
	if !ok {
		return nil, nil
	}
	to, err := opt.Parse(from)
	return &to, err
}

func UsingParserOfStrict[T any](from any) (T, error) {
	var to T
	opt, ok := any(&to).(Parsable[T])
	if !ok {
		return to, fmt.Errorf("%T not a parsable type", to)
	}
	to, err := opt.Parse(from)
	return to, err
}
