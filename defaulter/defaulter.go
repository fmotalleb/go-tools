package defaulter

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/fmotalleb/go-tools/decoder"
	"github.com/fmotalleb/go-tools/env"
	"github.com/fmotalleb/go-tools/template"
)

// ApplyDefaults walks v - which should be a pointer, typically to a struct -
// and fills every zero-valued field tagged with `default:"..."`. If the
// field also carries an `env:"..."` tag, the environment variable it names
// takes precedence over the static default (but a value already set by the
// caller, e.g. from a decoded config file, is never overwritten - only zero
// fields are touched). Default values may be Go templates (e.g.
// `default:"{{.Some.Field}}"`), evaluated against data.
//
// It recurses into nested structs, slices/arrays, maps, and pointers.
// Nil struct pointers are allocated so that defaults can still be applied
// deep inside optional sub-config trees. Pointer cycles are guarded against
// via a visited-address set.
//
// Any errors encountered while decoding a default value into its target
// field are collected and returned via errors.Join rather than discarded;
// a nil return means every tagged field that needed a default got one.
func ApplyDefaults(v any, data any) error {
	visited := make(map[uintptr]bool)
	var errs []error
	applyRecursive(data, reflect.ValueOf(v), visited, &errs)
	return errors.Join(errs...)
}

func applyRecursive(data any, val reflect.Value, visited map[uintptr]bool, errs *[]error) {
	if !val.IsValid() {
		return
	}

	// Dereference pointers, allocating nil ones (when we can) so defaults
	// can still reach fields inside optional sub-structs.
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			if !val.CanSet() {
				return
			}
			val.Set(reflect.New(val.Type().Elem()))
		}
		ptr := val.Pointer()
		if visited[ptr] {
			return
		}
		visited[ptr] = true
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Struct:
		t := val.Type()
		for i := 0; i < val.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			fv := val.Field(i)

			envKey := f.Tag.Get("env")
			defVal := f.Tag.Get("default")
			if def := env.Or(envKey, defVal); def != "" {
				if err := applyDefault(fv, def, data); err != nil {
					*errs = append(*errs, fmt.Errorf("%s: %w", f.Name, err))
				}
			}

			if fv.CanAddr() {
				applyRecursive(data, fv.Addr(), visited, errs)
			} else {
				applyRecursive(data, fv, visited, errs)
			}
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if elem.CanAddr() {
				applyRecursive(data, elem.Addr(), visited, errs)
			} else {
				applyRecursive(data, elem, visited, errs)
			}
		}

	case reflect.Map:
		elemType := val.Type().Elem()
		for _, key := range val.MapKeys() {
			// Map values obtained via MapIndex are never addressable/settable,
			// so struct (or nested pointer/slice) values living inside a map
			// can't be defaulted in place. Work on an addressable copy and
			// write it back into the map afterward.
			copyVal := reflect.New(elemType).Elem()
			copyVal.Set(val.MapIndex(key))
			if copyVal.CanAddr() {
				applyRecursive(data, copyVal.Addr(), visited, errs)
			} else {
				applyRecursive(data, copyVal, visited, errs)
			}
			val.SetMapIndex(key, copyVal)
		}
	}
}

// applyDefault decodes def (after template evaluation) into field, but only
// if field is still at its zero value. Errors from decoding are returned
// rather than swallowed, except for the plain-string fallback case, which
// always succeeds.
func applyDefault(field reflect.Value, def string, data any) error {
	if !field.CanSet() || !field.IsZero() {
		return nil
	}

	defValue, err := template.EvaluateTemplate(def, data)
	if err != nil {
		defValue = def
	}

	newValue := reflect.New(field.Type())
	if decodeErr := decoder.Decode(newValue.Interface(), defValue); decodeErr == nil {
		field.Set(newValue.Elem())
		return nil
	} else if field.Kind() == reflect.String {
		// decoder.Decode can fail for plain/named string types depending on
		// how it's configured; falling back to a direct set still respects
		// named string types via Convert, avoiding a reflect panic on type
		// mismatch (e.g. type Foo string vs the raw `string` defValue).
		field.Set(reflect.ValueOf(defValue).Convert(field.Type()))
		return nil
	} else {
		return fmt.Errorf("failed to decode default value %q into %s: %w", def, field.Type(), decodeErr)
	}
}