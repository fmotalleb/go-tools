package decoder

import (
	"errors"
	"time"

	"github.com/go-viper/mapstructure/v2"
)

func GetHooks() []mapstructure.DecodeHookFunc {
	return []mapstructure.DecodeHookFunc{
		// disabled for now as the loose type caster supports more types
		// mapstructure.StringToBasicTypeHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToNetIPAddrPortHookFunc(),
		mapstructure.StringToNetIPAddrHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		mapstructure.TextUnmarshallerHookFunc(),
		mapstructure.StringToURLHookFunc(),
		mapstructure.StringToIPHookFunc(),
		mapstructure.StringToIPNetHookFunc(),
		mapstructure.StringToTimeHookFunc(time.RFC3339),
		mapstructure.RecursiveStructToMapHookFunc(),
		LooseTypeCaster(),
		DecodeHookFunc(),
	}
}

func Build(item any) (*mapstructure.Decoder, error) {
	hook := mapstructure.ComposeDecodeHookFunc(
		GetHooks(),
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
