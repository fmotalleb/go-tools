package hooks

import (
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

// stringToSliceHookFunc converts strings to slices.
func StringToSliceHookFunc(sep string) mapstructure.DecodeHookFunc {
	return func(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
		if from.Kind() != reflect.String || to.Kind() != reflect.Slice {
			return data, nil
		}
		slice := strings.Split(data.(string), sep)
		return slice, nil
	}
}
