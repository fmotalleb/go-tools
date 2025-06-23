package decoder

import (
	"errors"
	"time"

	"github.com/go-viper/mapstructure/v2"
)

func GetHooks() []mapstructure.DecodeHookFunc {
	return []mapstructure.DecodeHookFunc{
		// disabled for now as the loose type caster supports more types
		StringToSliceHookFunc(","),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		LooseTypeCaster(),
		mapstructure.StringToNetIPAddrPortHookFunc(),
		mapstructure.StringToNetIPAddrHookFunc(),
		mapstructure.TextUnmarshallerHookFunc(),
		mapstructure.StringToURLHookFunc(),
		mapstructure.StringToIPHookFunc(),
		mapstructure.StringToIPNetHookFunc(),
		mapstructure.RecursiveStructToMapHookFunc(),
		// mapstructure.StringToBasicTypeHookFunc(),
		DecodeHookFunc(),
	}
}

func Build(item any) (*mapstructure.Decoder, error) {
	allHooks := registeredHooks
	allHooks = append(allHooks, GetHooks()...)

	hook := mapstructure.ComposeDecodeHookFunc(
		allHooks...,
	)

	decoderConfig := &mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           &item,
		TagName:          "mapstructure",
		WeaklyTypedInput: true,
		DecodeHook:       hook,
		DecodeNil:        true,
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
