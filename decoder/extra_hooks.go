package decoder

import (
	"fmt"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/cast"
)

// var looseTypeCaster func(reflect.Type, reflect.Type, interface{}) (interface{}, error) = LooseTypeCaster()

func LooseTypeCaster() func(reflect.Type, reflect.Type, interface{}) (interface{}, error) {
	return func(from, to reflect.Type, val interface{}) (interface{}, error) {
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
			// Cast doesn't have direct complex support, handle manually
			if str, ok := val.(string); ok {
				var c complex64
				_, err := fmt.Sscanf(str, "(%g%+gi)", &c)
				if err != nil {
					return val, err
				}
				return c, nil
			}
			return val, fmt.Errorf("cannot convert %T to complex64", val)
		case reflect.Complex128:
			if str, ok := val.(string); ok {
				var c complex128
				_, err := fmt.Sscanf(str, "(%g%+gi)", &c)
				if err != nil {
					return val, err
				}
				return c, nil
			}
			return val, fmt.Errorf("cannot convert %T to complex128", val)

		// String type
		case reflect.String:
			return cast.ToStringE(val)

		// Boolean type
		case reflect.Bool:
			return cast.ToBoolE(val)

		// Slice types
		case reflect.Slice:
			return handleSliceConversion(from, to, val)

		// Array types
		case reflect.Array:
			return handleArrayConversion(from, to, val)

		// Map types
		case reflect.Map:
			return handleMapConversion(from, to, val)

		// Pointer types
		case reflect.Ptr:
			return handlePointerConversion(from, to, val)

		// Interface types
		case reflect.Interface:
			// For interface{}, just return the value as-is
			if to == reflect.TypeOf((*interface{})(nil)).Elem() {
				return val, nil
			}
			return val, nil

		// Struct types
		case reflect.Struct:
			return handleStructConversion(from, to, val)

		// Channel, Function, UnsafePointer - not supported
		case reflect.Chan, reflect.Func, reflect.UnsafePointer:
			return val, fmt.Errorf("conversion to %s not supported", to.Kind())

		// Uintptr type
		case reflect.Uintptr:
			if num, err := cast.ToUint64E(val); err == nil {
				return uintptr(num), nil
			}
			return val, fmt.Errorf("cannot convert %T to uintptr", val)

		default:
			return val, nil
		}
	}
}

func handleSliceConversion(from, to reflect.Type, val interface{}) (interface{}, error) {
	elemType := to.Elem()

	switch elemType.Kind() {
	case reflect.String:
		return cast.ToStringSliceE(val)
	case reflect.Int:
		return cast.ToIntSliceE(val)
	case reflect.Bool:
		return cast.ToBoolSliceE(val)
	case reflect.Float64:
		// Cast doesn't have ToFloat64SliceE, handle manually
		if slice, err := cast.ToSliceE(val); err == nil {
			result := make([]float64, len(slice))
			for i, item := range slice {
				if f, err := cast.ToFloat64E(item); err == nil {
					result[i] = f
				} else {
					return val, err
				}
			}
			return result, nil
		}
		return val, fmt.Errorf("cannot convert %T to []float64", val)
	case reflect.Interface:
		// For []interface{}
		return cast.ToSliceE(val)
	default:
		// For other slice types, try to convert each element
		if slice, err := cast.ToSliceE(val); err == nil {
			resultSlice := reflect.MakeSlice(to, len(slice), len(slice))
			for i, item := range slice {
				// Recursively convert each element
				converted, err := LooseTypeCaster()(reflect.TypeOf(item), elemType, item)
				if err != nil {
					return val, err
				}
				resultSlice.Index(i).Set(reflect.ValueOf(converted))
			}
			return resultSlice.Interface(), nil
		}
		return val, fmt.Errorf("cannot convert %T to %s", val, to)
	}
}

func handleArrayConversion(from, to reflect.Type, val interface{}) (interface{}, error) {
	// Convert to slice first, then to array
	sliceType := reflect.SliceOf(to.Elem())
	slice, err := handleSliceConversion(from, sliceType, val)
	if err != nil {
		return val, err
	}

	sliceVal := reflect.ValueOf(slice)
	arrayVal := reflect.New(to).Elem()

	// Copy elements from slice to array
	minLen := arrayVal.Len()
	if sliceVal.Len() < minLen {
		minLen = sliceVal.Len()
	}

	for i := 0; i < minLen; i++ {
		arrayVal.Index(i).Set(sliceVal.Index(i))
	}

	return arrayVal.Interface(), nil
}

func handleMapConversion(from, to reflect.Type, val interface{}) (interface{}, error) {
	// Handle common map types
	keyType := to.Key()
	valueType := to.Elem()

	if keyType.Kind() == reflect.String && valueType.Kind() == reflect.Interface {
		// map[string]interface{}
		return cast.ToStringMapE(val)
	}

	if keyType.Kind() == reflect.String && valueType.Kind() == reflect.String {
		// map[string]string
		return cast.ToStringMapStringE(val)
	}

	if keyType.Kind() == reflect.String && valueType.Kind() == reflect.Slice &&
		valueType.Elem().Kind() == reflect.String {
		// map[string][]string
		return cast.ToStringMapStringSliceE(val)
	}

	// For other map types, try manual conversion
	if sourceMap, err := cast.ToStringMapE(val); err == nil {
		resultMap := reflect.MakeMap(to)

		for k, v := range sourceMap {
			// Convert key
			convertedKey, err := LooseTypeCaster()(reflect.TypeOf(k), keyType, k)
			if err != nil {
				return val, err
			}

			// Convert value
			convertedValue, err := LooseTypeCaster()(reflect.TypeOf(v), valueType, v)
			if err != nil {
				return val, err
			}

			resultMap.SetMapIndex(reflect.ValueOf(convertedKey), reflect.ValueOf(convertedValue))
		}

		return resultMap.Interface(), nil
	}

	return val, fmt.Errorf("cannot convert %T to %s", val, to)
}

func handlePointerConversion(from, to reflect.Type, val interface{}) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	elemType := to.Elem()

	// If the value is already a pointer, dereference it
	if from.Kind() == reflect.Ptr {
		val = reflect.ValueOf(val).Elem().Interface()
		from = reflect.TypeOf(val)
	}

	// Convert the underlying value
	converted, err := LooseTypeCaster()(from, elemType, val)
	if err != nil {
		return val, err
	}

	// Create a pointer to the converted value
	ptr := reflect.New(elemType)
	ptr.Elem().Set(reflect.ValueOf(converted))

	return ptr.Interface(), nil
}

func handleStructConversion(from, to reflect.Type, val interface{}) (interface{}, error) {
	// For struct conversion, we typically need mapstructure itself
	// This is a basic implementation - you might want to use mapstructure.Decode here

	if from.Kind() == reflect.Map {
		// Convert map to struct using mapstructure
		result := reflect.New(to).Interface()
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:     result,
			DecodeHook: LooseTypeCaster(),
		})
		if err != nil {
			return val, err
		}

		if err := decoder.Decode(val); err != nil {
			return val, err
		}

		return reflect.ValueOf(result).Elem().Interface(), nil
	}

	return val, fmt.Errorf("cannot convert %T to %s", val, to)
}
