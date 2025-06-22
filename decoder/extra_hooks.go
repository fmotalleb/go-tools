package decoder

import (
	"fmt"
	"reflect"

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

	// // Slice types
	// case reflect.Slice:
	// 	return convertToSlice(from, to, val)

	// // Array types
	// case reflect.Array:
	// 	return convertToArray(from, to, val)

	// // Map types
	// case reflect.Map:
	// 	return convertToMap(from, to, val)

	// // Pointer types
	// case reflect.Ptr:
	// 	return convertToPointer(from, to, val)

	// // Interface types
	// case reflect.Interface:
	// 	if to == reflect.TypeOf((*interface{})(nil)).Elem() {
	// 		return val, nil
	// 	}
	// 	return val, nil

	// // Struct types
	// case reflect.Struct:
	// 	return convertToStruct(from, to, val)

	// Uintptr type
	// case reflect.Uintptr:
	// 	if num, err := cast.ToUint64E(val); err == nil {
	// 		return uintptr(num), nil
	// 	}
	// 	return val, fmt.Errorf("cannot convert %T to uintptr", val)

	// Unsupported types
	// case reflect.Chan, reflect.Func, reflect.UnsafePointer:
	// 	return val, fmt.Errorf("conversion to %s not supported", to.Kind())

	default:
		return val, nil
	}
}

func convertToComplex64(val interface{}) (interface{}, error) {
	if str, ok := val.(string); ok {
		var real, imag float32
		n, err := fmt.Sscanf(str, "(%g%+gi)", &real, &imag)
		if err != nil || n != 2 {
			// Try alternative formats
			n, err = fmt.Sscanf(str, "%g%+gi", &real, &imag)
			if err != nil || n != 2 {
				return val, fmt.Errorf("cannot parse complex64 from string: %s", str)
			}
		}
		return complex(real, imag), nil
	}
	return val, fmt.Errorf("cannot convert %T to complex64", val)
}

func convertToComplex128(val interface{}) (interface{}, error) {
	if str, ok := val.(string); ok {
		var real, imag float64
		n, err := fmt.Sscanf(str, "(%g%+gi)", &real, &imag)
		if err != nil || n != 2 {
			n, err = fmt.Sscanf(str, "%g%+gi", &real, &imag)
			if err != nil || n != 2 {
				return val, fmt.Errorf("cannot parse complex128 from string: %s", str)
			}
		}
		return complex(real, imag), nil
	}
	return val, fmt.Errorf("cannot convert %T to complex128", val)
}

func convertToSlice(from, to reflect.Type, val interface{}) (interface{}, error) {
	elemType := to.Elem()

	// Handle common slice types using cast library
	switch elemType.Kind() {
	case reflect.String:
		return cast.ToStringSliceE(val)
	case reflect.Int:
		return cast.ToIntSliceE(val)
	case reflect.Bool:
		return cast.ToBoolSliceE(val)
	case reflect.Interface:
		return cast.ToSliceE(val)
	}

	// For other types, convert manually
	sourceSlice, err := cast.ToSliceE(val)
	if err != nil {
		return val, err
	}

	resultSlice := reflect.MakeSlice(to, len(sourceSlice), len(sourceSlice))
	for i, item := range sourceSlice {
		converted, err := convertValue(reflect.TypeOf(item), elemType, item)
		if err != nil {
			return val, fmt.Errorf("failed to convert slice element at index %d: %w", i, err)
		}
		resultSlice.Index(i).Set(reflect.ValueOf(converted))
	}

	return resultSlice.Interface(), nil
}

func convertToArray(from, to reflect.Type, val interface{}) (interface{}, error) {
	sourceSlice, err := cast.ToSliceE(val)
	if err != nil {
		return val, err
	}

	arrayLen := to.Len()
	arrayVal := reflect.New(to).Elem()
	elemType := to.Elem()

	minLen := arrayLen
	if len(sourceSlice) < minLen {
		minLen = len(sourceSlice)
	}

	for i := 0; i < minLen; i++ {
		converted, err := convertValue(reflect.TypeOf(sourceSlice[i]), elemType, sourceSlice[i])
		if err != nil {
			return val, fmt.Errorf("failed to convert array element at index %d: %w", i, err)
		}
		arrayVal.Index(i).Set(reflect.ValueOf(converted))
	}

	return arrayVal.Interface(), nil
}

func convertToMap(from, to reflect.Type, val interface{}) (interface{}, error) {
	keyType := to.Key()
	valueType := to.Elem()

	// Handle common map types
	if keyType.Kind() == reflect.String {
		switch valueType.Kind() {
		case reflect.Interface:
			return cast.ToStringMapE(val)
		case reflect.String:
			return cast.ToStringMapStringE(val)
		case reflect.Slice:
			if valueType.Elem().Kind() == reflect.String {
				return cast.ToStringMapStringSliceE(val)
			}
		}
	}

	// Manual conversion for other map types
	sourceMap, err := cast.ToStringMapE(val)
	if err != nil {
		return val, err
	}

	resultMap := reflect.MakeMap(to)
	for k, v := range sourceMap {
		convertedKey, err := convertValue(reflect.TypeOf(k), keyType, k)
		if err != nil {
			return val, fmt.Errorf("failed to convert map key %v: %w", k, err)
		}

		convertedValue, err := convertValue(reflect.TypeOf(v), valueType, v)
		if err != nil {
			return val, fmt.Errorf("failed to convert map value for key %v: %w", k, err)
		}

		resultMap.SetMapIndex(reflect.ValueOf(convertedKey), reflect.ValueOf(convertedValue))
	}

	return resultMap.Interface(), nil
}

func convertToPointer(from, to reflect.Type, val interface{}) (interface{}, error) {
	if val == nil {
		return nil, nil
	}

	elemType := to.Elem()

	// Dereference if source is pointer
	if from.Kind() == reflect.Ptr {
		sourceVal := reflect.ValueOf(val)
		if sourceVal.IsNil() {
			return nil, nil
		}
		val = sourceVal.Elem().Interface()
		from = reflect.TypeOf(val)
	}

	converted, err := convertValue(from, elemType, val)
	if err != nil {
		return val, err
	}

	ptr := reflect.New(elemType)
	ptr.Elem().Set(reflect.ValueOf(converted))

	return ptr.Interface(), nil
}

func convertToStruct(from, to reflect.Type, val interface{}) (interface{}, error) {
	if from.Kind() != reflect.Map {
		return val, fmt.Errorf("can only convert map to struct, got %s", from.Kind())
	}

	// Use mapstructure without recursive decode hook to avoid stack overflow
	result := reflect.New(to).Interface()
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           result,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	})
	if err != nil {
		return val, err
	}

	if err := decoder.Decode(val); err != nil {
		return val, err
	}

	return reflect.ValueOf(result).Elem().Interface(), nil
}

// Helper function for direct value conversion without decode hooks
func convertValue(from, to reflect.Type, val interface{}) (interface{}, error) {
	if from == to {
		return val, nil
	}

	// Use the same logic as the main function but without recursion
	return looseTypeCasterImpl(from, to, val)
}
