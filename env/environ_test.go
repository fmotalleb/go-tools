package env_test

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/fmotalleb/go-tools/env"
)

func TestOr(t *testing.T) {
	const key = "TEST_OR"

	// unset
	os.Unsetenv(key)
	if got := env.Or(key, "default"); got != "default" {
		t.Errorf("Or() = %v; want %v", got, "default")
	}

	// set
	os.Setenv(key, "value")
	if got := env.Or(key, "default"); got != "value" {
		t.Errorf("Or() = %v; want %v", got, "value")
	}

	// empty key
	if got := env.Or("", "default"); got != "default" {
		t.Errorf("Or() with empty key = %v; want %v", got, "default")
	}
}

func TestBoolOr(t *testing.T) {
	const key = "TEST_BOOL_OR"

	os.Unsetenv(key)
	if got := env.BoolOr(key, true); got != true {
		t.Errorf("BoolOr() = %v; want %v", got, true)
	}

	os.Setenv(key, "false")
	if got := env.BoolOr(key, true); got != false {
		t.Errorf("BoolOr() = %v; want %v", got, false)
	}

	os.Setenv(key, "invalid")
	if got := env.BoolOr(key, true); got != true {
		t.Errorf("BoolOr() = %v; want %v", got, true)
	}

	if got := env.BoolOr("", false); got != false {
		t.Errorf("BoolOr() with empty key = %v; want %v", got, false)
	}
}

func TestIntOr(t *testing.T) {
	const key = "TEST_INT_OR"

	os.Unsetenv(key)
	if got := env.IntOr(key, 42); got != 42 {
		t.Errorf("IntOr() = %v; want %v", got, 42)
	}

	os.Setenv(key, "123")
	if got := env.IntOr(key, 42); got != 123 {
		t.Errorf("IntOr() = %v; want %v", got, 123)
	}

	os.Setenv(key, "invalid")
	if got := env.IntOr(key, 42); got != 42 {
		t.Errorf("IntOr() = %v; want %v", got, 42)
	}

	if got := env.IntOr("", 7); got != 7 {
		t.Errorf("IntOr() with empty key = %v; want %v", got, 7)
	}
}

func TestSliceOr(t *testing.T) {
	const key = "TEST_SLICE_OR"

	os.Unsetenv(key)
	def := []string{"a", "b"}
	if got := env.SliceOr(key, def); !reflect.DeepEqual(got, def) {
		t.Errorf("SliceOr() = %v; want %v", got, def)
	}

	os.Setenv(key, "x,y,z")
	if got := env.SliceOr(key, def); !reflect.DeepEqual(got, []string{"x", "y", "z"}) {
		t.Errorf("SliceOr() = %v; want %v", got, []string{"x", "y", "z"})
	}

	if got := env.SliceOr("", def); !reflect.DeepEqual(got, def) {
		t.Errorf("SliceOr() = %v; want %v", got, def)
	}
}

func TestSliceSeparatorOr(t *testing.T) {
	const key = "TEST_SLICE_SEP_OR"

	os.Unsetenv(key)
	def := []string{"1", "2"}
	if got := env.SliceSeparatorOr(key, ";", def); !reflect.DeepEqual(got, def) {
		t.Errorf("SliceSeparatorOr() = %v; want %v", got, def)
	}

	os.Setenv(key, "a;b;c")
	if got := env.SliceSeparatorOr(key, ";", def); !reflect.DeepEqual(got, []string{"a", "b", "c"}) {
		t.Errorf("SliceSeparatorOr() = %v; want %v", got, []string{"a", "b", "c"})
	}
}

func TestDurationOr(t *testing.T) {
	const key = "TEST_DURATION_OR"

	os.Unsetenv(key)
	def := time.Second * 5
	if got := env.DurationOr(key, def); got != def {
		t.Errorf("DurationOr() = %v; want %v", got, def)
	}

	os.Setenv(key, "10s")
	if got := env.DurationOr(key, def); got != 10*time.Second {
		t.Errorf("DurationOr() = %v; want %v", got, 10*time.Second)
	}

	os.Setenv(key, "invalid")
	if got := env.DurationOr(key, def); got != def {
		t.Errorf("DurationOr() = %v; want %v", got, def)
	}

	if got := env.DurationOr("", time.Minute); got != time.Minute {
		t.Errorf("DurationOr() with empty key = %v; want %v", got, time.Minute)
	}
}
