package decoder

import (
	"errors"

	"github.com/go-viper/mapstructure/v2"
)

func GetHooksDefault() []mapstructure.DecodeHookFunc {
	return []mapstructure.DecodeHookFunc{
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		DecodeHookFunc(),
	}
}

func Build(item any) (*mapstructure.Decoder, error) {
	hook := mapstructure.ComposeDecodeHookFunc(
		GetHooksDefault(),
	)

	decoderConfig := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           &item,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		DecodeHook:       hook,
		DecodeNil:        true,
		ZeroFields:       true,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	return decoder, err
}

func Decode[T any](dst *T, src any) error {
	decoder, err := Build(dst)
	if err != nil {
		return errors.Join(
			errors.New("failed to create decoder"),
			err,
		)
	}
	if err := decoder.Decode(src); err != nil {
		return errors.Join(
			errors.New("failed to decode"),
			err,
		)
	}
	return nil
}
