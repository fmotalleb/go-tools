package hooks

import (
	"reflect"

	"github.com/fmotalleb/go-tools/env"
	"github.com/go-viper/mapstructure/v2"
)

// EnvSubst applies bash env substitution on the given input.
func EnvSubst() mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.String {
			return data, nil
		}
		strVal := data.(string)
		out := env.Subst(strVal)
		return out, nil
	}
}
