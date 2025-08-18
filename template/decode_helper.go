package template

import (
	"reflect"

	"github.com/go-viper/mapstructure/v2"
)

func StringTemplateEvaluate() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, val interface{}) (interface{}, error) {
		if from.Kind() != reflect.String {
			return val, nil
		}
		str := val.(string)
		return EvaluateTemplate(str, struct{}{})
	}
}

// func EvaluateOnStruct(v any, data any) {
// 	visited := make(map[uintptr]bool)
// 	findTemplateFieldsRecursive(data, reflect.ValueOf(v), visited, "")
// }

// func findTemplateFieldsRecursive(data any, val reflect.Value, visited map[uintptr]bool, path string) {
// 	if !val.IsValid() {
// 		return
// 	}

// 	// Dereference pointers
// 	for val.Kind() == reflect.Pointer {
// 		if val.IsNil() {
// 			return
// 		}
// 		ptr := val.Pointer()
// 		if visited[ptr] {
// 			return
// 		}
// 		visited[ptr] = true
// 		val = val.Elem()
// 	}

// 	t := val.Type()

// 	switch val.Kind() {
// 	case reflect.Struct:
// 		for i := 0; i < val.NumField(); i++ {
// 			f := t.Field(i)
// 			fv := val.Field(i)

// 			currentPath := path + "." + f.Name
// 			if f.Tag.Get("template") == "true" {
// 				field := val.Field(i)
// 				input := field.text
// 				result := EvaluateTemplate(input, data)
// 				field.text = result
// 			}

// 			findTemplateFieldsRecursive(data, fv, visited, currentPath)
// 		}

// 	case reflect.Slice, reflect.Array:
// 		for i := 0; i < val.Len(); i++ {
// 			findTemplateFieldsRecursive(data, val.Index(i), visited, fmt.Sprintf("%s[%d]", path, i))
// 		}

// 	case reflect.Map:
// 		for _, key := range val.MapKeys() {
// 			findTemplateFieldsRecursive(data, val.MapIndex(key), visited, fmt.Sprintf("%s[%v]", path, key))
// 		}
// 	}
// }
