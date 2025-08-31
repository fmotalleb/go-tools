package defaulter

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/fmotalleb/go-tools/env"
	"github.com/fmotalleb/go-tools/template"
)

func ApplyDefaults(v any, data any) {
	visited := make(map[uintptr]bool)
	findTemplateFieldsRecursive(data, reflect.ValueOf(v), visited, "")
}

func findTemplateFieldsRecursive(data any, val reflect.Value, visited map[uintptr]bool, path string) {
	if !val.IsValid() {
		return
	}

	// Dereference pointers
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return
		}
		ptr := val.Pointer()
		if visited[ptr] {
			return
		}
		visited[ptr] = true
		val = val.Elem()
	}

	t := val.Type()

	switch val.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			f := t.Field(i)
			fv := val.Field(i)

			currentPath := path + "." + f.Name
			envKey := f.Tag.Get("env")
			defVal := f.Tag.Get("default")
			if def := env.Or(envKey, defVal); def != "" {
				applyDefault(val, i, def, data)
			}

			if fv.CanAddr() {
				findTemplateFieldsRecursive(data, fv.Addr(), visited, currentPath)
			} else {
				findTemplateFieldsRecursive(data, fv, visited, currentPath)
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			findTemplateFieldsRecursive(data, val.Index(i), visited, fmt.Sprintf("%s[%d]", path, i))
		}

	case reflect.Map:
		for _, key := range val.MapKeys() {
			findTemplateFieldsRecursive(data, val.MapIndex(key), visited, fmt.Sprintf("%s[%v]", path, key))
		}
	}
}

func applyDefault(val reflect.Value, i int, def string, data any) {
	field := val.Field(i)
	defValue, err := template.EvaluateTemplate(def, data)
	if err != nil {
		defValue = def
	}
	if field.CanSet() && field.IsZero() {
		newValue := reflect.New(field.Type()).Interface()
		if err := json.Unmarshal([]byte(defValue), newValue); err == nil {
			field.Set(reflect.ValueOf(newValue).Elem())
		} else if field.Kind() == reflect.String {
			field.Set(reflect.ValueOf(defValue))
		}
	}
}
