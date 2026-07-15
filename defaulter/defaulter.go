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

	switch val.Kind() {
	case reflect.Pointer:
		// Cycle/visited tracking must only ever key off values that are
		// genuinely pointer-typed in the data model (e.g. `Next *Node`).
		// Earlier revisions also wrapped every traversed struct/slice/map
		// field in .Addr() purely to get a settable handle, and fed that
		// synthetic pointer through this same visited-address check. That's
		// unsound: in Go, a struct's first field starts at the same address
		// as the struct itself, so &container == &container.FirstField for
		// any single/first-field struct. That collision made the visited
		// set report a false "already seen" on the very first visit to
		// that field, silently skipping it. By only entering this branch
		// for values that were *already* reflect.Pointer before we ever
		// touched them (never for our own .Addr() wrapping, which no longer
		// happens - see the other cases below), the addresses we track are
		// always genuine distinct allocations, so no false positives.
		if val.IsNil() {
			if !val.CanSet() {
				return
			}
			// Only allocate the pointer when its element type could
			// contain nested tagged defaults (structs, maps, slices,
			// or further pointers). For basic types like *string or
			// *int there is nothing to recurse into, and keeping them
			// nil lets callers distinguish "unset" from "set to zero".
			switch val.Type().Elem().Kind() {
			case reflect.Struct, reflect.Map, reflect.Slice, reflect.Ptr:
				val.Set(reflect.New(val.Type().Elem()))
			default:
				return
			}
		}
		ptr := val.Pointer()
		if visited[ptr] {
			return
		}
		visited[ptr] = true
		applyRecursive(data, val.Elem(), visited, errs)

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

			// fv is already addressable/settable whenever val (the parent
			// struct) is - no .Addr()/re-dereference wrapping needed. If fv
			// is itself pointer-typed, it lands in the Pointer case above on
			// its own terms, with its own genuine target address.
			applyRecursive(data, fv, visited, errs)
		}

	case reflect.Slice, reflect.Array:
		// Slice elements are always addressable regardless of whether the
		// slice header itself is addressable (they live in the backing
		// array); array elements are addressable iff the array is. Either
		// way, no synthetic pointer wrapping is needed here.
		for i := 0; i < val.Len(); i++ {
			applyRecursive(data, val.Index(i), visited, errs)
		}

	case reflect.Map:
		elemType := val.Type().Elem()
		for _, key := range val.MapKeys() {
			// Map values obtained via MapIndex are never addressable/settable,
			// so struct (or nested pointer/slice) values living inside a map
			// can't be defaulted in place. Work on a freshly allocated,
			// inherently addressable copy and write it back into the map
			// afterward - no .Addr() wrapping/visited tracking needed since
			// this is a brand-new allocation, never colliding with any
			// existing address.
			copyVal := reflect.New(elemType).Elem()
			copyVal.Set(val.MapIndex(key))
			applyRecursive(data, copyVal, visited, errs)
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
