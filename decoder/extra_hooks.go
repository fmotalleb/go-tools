package decoder

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cast"
)

func LooseTypeCaster() mapstructure.DecodeHookFunc {
	return looseTypeCasterImpl
}

func looseTypeCasterImpl(from, to reflect.Type, val interface{}) (interface{}, error) {
	if from.Kind() == to.Kind() {
		return val, nil
	}

	switch to.Kind() {
	// Integer types
	case reflect.Int:
		return cast.ToIntE(val)
	case reflect.Int8:
		return cast.ToInt8E(val)
	case reflect.Int16:
		return cast.ToInt16E(val)
	case reflect.Int32:
		return cast.ToInt32E(val)
	case reflect.Int64:
		return cast.ToInt64E(val)

	// Unsigned integer types
	case reflect.Uint:
		return cast.ToUintE(val)
	case reflect.Uint8:
		return cast.ToUint8E(val)
	case reflect.Uint16:
		return cast.ToUint16E(val)
	case reflect.Uint32:
		return cast.ToUint32E(val)
	case reflect.Uint64:
		return cast.ToUint64E(val)

	// Floating point types
	case reflect.Float32:
		return cast.ToFloat32E(val)
	case reflect.Float64:
		return cast.ToFloat64E(val)

	// Complex types
	case reflect.Complex64:
		return convertToComplex64(val)
	case reflect.Complex128:
		return convertToComplex128(val)

	// String type
	case reflect.String:
		return cast.ToStringE(val)

	// Boolean type
	case reflect.Bool:
		return cast.ToBoolE(val)

	default:
		return val, nil
	}
}

func convertToComplex64(val interface{}) (interface{}, error) {
	if str, ok := val.(string); ok {
		var realN, imagN float32
		n, err := fmt.Sscanf(str, "(%g%+gi)", &realN, &imagN)
		if err != nil || n != 2 {
			// Try alternative formats
			n, err = fmt.Sscanf(str, "%g%+gi", &realN, &imagN)
			if err != nil || n != 2 {
				return val, fmt.Errorf("cannot parse complex64 from string: %s", str)
			}
		}
		return complex(realN, imagN), nil
	}
	return val, fmt.Errorf("cannot convert %T to complex64", val)
}

func convertToComplex128(val interface{}) (interface{}, error) {
	if str, ok := val.(string); ok {
		var realN, imagN float64
		n, err := fmt.Sscanf(str, "(%g%+gi)", &realN, &imagN)
		if err != nil || n != 2 {
			n, err = fmt.Sscanf(str, "%g%+gi", &realN, &imagN)
			if err != nil || n != 2 {
				return val, fmt.Errorf("cannot parse complex128 from string: %s", str)
			}
		}
		return complex(realN, imagN), nil
	}
	return val, fmt.Errorf("cannot convert %T to complex128", val)
}

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
