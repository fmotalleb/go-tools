package decoder

import (
	"errors"
	"time"

	"github.com/fmotalleb/go-tools/decoder/hooks"
	"github.com/fmotalleb/go-tools/template"
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

func Build[T any](item T, extraHooks ...mapstructure.DecodeHookFunc) (*mapstructure.Decoder, error) {
	allHooks := GetHooks()
	allHooks = append(allHooks, hooks.GetExtraHooks()...)
	if len(extraHooks) != 0 {
		allHooks = append(allHooks, extraHooks...)
	}
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

func Decode(dst any, src any) error {
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

func DecodeWithTemplate(dst any, src any, data any) error {
	hook := template.StringTemplateEvaluate(data)
	decoder, err := Build(dst, hook)
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
