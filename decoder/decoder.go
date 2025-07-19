package decoder

import (
	"errors"
	"time"

	"github.com/fmotalleb/go-tools/decoder/hooks"
	"github.com/go-viper/mapstructure/v2"
)

func GetHooks() []mapstructure.DecodeHookFunc {
	return []mapstructure.DecodeHookFunc{
		hooks.EnvSubst(),
		hooks.StringToSliceHookFunc(","),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		hooks.LooseTypeCaster(),
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
	allHooks := GetHooks()
	allHooks = append(allHooks, hooks.GetExtraHooks()...)

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
