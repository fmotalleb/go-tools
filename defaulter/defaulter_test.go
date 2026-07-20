package defaulter_test

import (
	"strings"
	"testing"
	"time"

	"github.com/fmotalleb/go-tools/defaulter"
)

// --- basic tag-driven defaulting ---------------------------------------------------

type basic struct {
	Name    string        `default:"anon"`
	Count   int           `default:"5"`
	Timeout time.Duration `default:"30s"`
}

func TestApplyDefaults_FillsZeroFields(t *testing.T) {
	v := &basic{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Name != "anon" {
		t.Errorf("Name = %q, want %q", v.Name, "anon")
	}
	if v.Count != 5 {
		t.Errorf("Count = %d, want 5", v.Count)
	}
	if v.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", v.Timeout)
	}
}

// NOTE: fields left at their type's zero value are indistinguishable from
// "never set", by design of tag-based defaulting (the same caveat documented
// by libraries like mcuadros/go-defaults). This test only exercises
// non-zero explicit values, which are the only case where overwrite
// protection is actually observable.
func TestApplyDefaults_DoesNotOverwriteExplicitNonZeroValues(t *testing.T) {
	v := &basic{Name: "explicit", Count: 42, Timeout: 5 * time.Second}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Name != "explicit" || v.Count != 42 || v.Timeout != 5*time.Second {
		t.Errorf("ApplyDefaults overwrote explicitly-set fields: %+v", v)
	}
}

// KNOWN LIMITATION (not a bug introduced by the refactor): because the
// zero value of bool is false, a default:"true" tag cannot distinguish
// "explicitly set to false" from "never set" - it will always win. Anyone
// tagging a bool default should be aware an explicit `false` in config
// will be overridden back to the tag's default.
type boolDefault struct {
	Enabled bool `default:"true"`
}

func TestApplyDefaults_BoolZeroValueAmbiguity(t *testing.T) {
	v := &boolDefault{Enabled: false} // indistinguishable from "unset"
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v.Enabled {
		t.Fatalf("expected known bool-zero-value limitation: Enabled should have been forced to true, got false")
	}
}

type untagged struct {
	Optional string
}

func TestApplyDefaults_LeavesUntaggedFieldsAlone(t *testing.T) {
	v := &untagged{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Optional != "" {
		t.Errorf("Optional = %q, want empty string (field has no default tag)", v.Optional)
	}
}

// --- env precedence -----------------------------------------------------------------

type envOverride struct {
	Value string `default:"fallback" env:"DEFAULTER_TEST_VALUE"`
}

func TestApplyDefaults_FallsBackToDefaultWhenEnvUnset(t *testing.T) {
	v := &envOverride{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Value != "fallback" {
		t.Errorf("Value = %q, want %q", v.Value, "fallback")
	}
}

func TestApplyDefaults_EnvOverridesDefaultTag(t *testing.T) {
	t.Setenv("DEFAULTER_TEST_VALUE", "from-env")
	v := &envOverride{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Value != "from-env" {
		t.Errorf("Value = %q, want %q (env var should take precedence over the static default)", v.Value, "from-env")
	}
}

func TestApplyDefaults_EnvDoesNotOverwriteExplicitValue(t *testing.T) {
	t.Setenv("DEFAULTER_TEST_VALUE", "from-env")
	v := &envOverride{Value: "explicitly-set"}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Value != "explicitly-set" {
		t.Errorf("Value = %q, want %q (an already-set field must win over both env and default)", v.Value, "explicitly-set")
	}
}

type envMissing struct {
	Value *string `env:"DEFAULTER_TEST_VALUE"`
}

func TestApplyDefaults_NoChangeWhenEnvUnset(t *testing.T) {
	v := &envMissing{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Value != nil {
		t.Errorf("Value = %v, want %v", v.Value, nil)
	}
}

type envMissingComplex struct {
	Value *struct {
		test string
	}
}

func TestApplyDefaults_NoChangeWhenEnvUnsetComplex(t *testing.T) {
	v := &envMissingComplex{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Value != nil {
		t.Errorf("Value = %v, want %v", v.Value, nil)
	}
}

// --- nested structs, slices, maps -----------------------------------------------------

type inner struct {
	Port int `default:"8080"`
}

type outer struct {
	Name  string `default:"outer"`
	Inner inner
}

func TestApplyDefaults_NestedStruct(t *testing.T) {
	v := &outer{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Name != "outer" {
		t.Errorf("Name = %q, want %q", v.Name, "outer")
	}
	if v.Inner.Port != 8080 {
		t.Errorf("Inner.Port = %d, want 8080", v.Inner.Port)
	}
}

type item struct {
	Weight int `default:"1"`
}

type withSlice struct {
	Items []item
}

func TestApplyDefaults_SliceOfStructs(t *testing.T) {
	v := &withSlice{Items: []item{{}, {Weight: 9}, {}}}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []int{1, 9, 1}
	for i, it := range v.Items {
		if it.Weight != want[i] {
			t.Errorf("Items[%d].Weight = %d, want %d", i, it.Weight, want[i])
		}
	}
}

type mapVal struct {
	Retries int `default:"3"`
}

type withMap struct {
	Backends map[string]mapVal
}

// This specifically exercises the map-write-back fix: MapIndex returns a
// non-addressable copy, so without copying out / SetMapIndex back in,
// defaults silently never reach struct values stored in a map.
func TestApplyDefaults_MapOfStructsGetsDefaultsWrittenBack(t *testing.T) {
	v := &withMap{
		Backends: map[string]mapVal{
			"a": {},           // should get the default
			"b": {Retries: 7}, // should be left alone
		},
	}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := v.Backends["a"].Retries; got != 3 {
		t.Errorf(`Backends["a"].Retries = %d, want 3 (map struct values must receive defaults)`, got)
	}
	if got := v.Backends["b"].Retries; got != 7 {
		t.Errorf(`Backends["b"].Retries = %d, want 7 (explicit map value must be preserved)`, got)
	}
}

// --- nil pointer allocation ----------------------------------------------------------

type subConfig struct {
	Level string `default:"info"`
}

type withPointer struct {
	Sub *subConfig
}

func TestApplyDefaults_AllocatesNilStructPointer(t *testing.T) {
	v := &withPointer{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Sub == nil {
		t.Fatal("Sub was not allocated; defaults inside a nil pointer sub-struct can't be applied without allocation")
	}
	if v.Sub.Level != "info" {
		t.Errorf("Sub.Level = %q, want %q", v.Sub.Level, "info")
	}
}

func TestApplyDefaults_LeavesAlreadyAllocatedPointerAlone(t *testing.T) {
	v := &withPointer{Sub: &subConfig{Level: "debug"}}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Sub.Level != "debug" {
		t.Errorf("Sub.Level = %q, want %q (pre-set pointer target must not be overwritten)", v.Sub.Level, "debug")
	}
}

// --- pointer cycles --------------------------------------------------------------------

type node struct {
	Name string `default:"node"`
	Next *node
}

func TestApplyDefaults_HandlesPointerCyclesWithoutHanging(t *testing.T) {
	a := &node{}
	b := &node{}
	a.Next = b
	b.Next = a // cycle

	done := make(chan error, 1)
	go func() {
		done <- defaulter.ApplyDefaults(a, nil)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ApplyDefaults did not return in time - likely an infinite loop walking the pointer cycle")
	}

	if a.Name != "node" || b.Name != "node" {
		t.Errorf("defaults not applied on cyclic nodes: a=%+v b=%+v", a, b)
	}
}

// --- named string types ---------------------------------------------------------------

type logLevel string

type withNamedString struct {
	LogLevel logLevel `default:"warn"`
}

// Guards against the reflect panic risk in the string-fallback path:
// reflect.ValueOf(defValue) produces a plain `string` Value, which can't be
// assigned directly to a named string type without an explicit Convert.
func TestApplyDefaults_NamedStringTypeDoesNotPanic(t *testing.T) {
	v := &withNamedString{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.LogLevel != "warn" {
		t.Errorf("LogLevel = %q, want %q", v.LogLevel, "warn")
	}
}

// --- unexported fields ------------------------------------------------------------------

type withUnexportedDefault struct {
	Public  string `default:"public-default"`
	private string `default:"should-never-be-set"` //nolint:unused // exercised via reflection only
}

// Unexported fields obtained via reflection are read-only; attempting to
// .Set() one panics. This test's main assertion is simply that ApplyDefaults
// does not panic on a struct containing a tagged unexported field, while
// still confirming exported sibling fields are defaulted normally.
func TestApplyDefaults_DoesNotPanicOnUnexportedTaggedField(t *testing.T) {
	v := &withUnexportedDefault{}
	if err := defaulter.ApplyDefaults(v, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Public != "public-default" {
		t.Errorf("Public = %q, want %q", v.Public, "public-default")
	}
}

// --- error propagation ------------------------------------------------------------------

type badDefault struct {
	Port int `default:"not-a-number"`
}

func TestApplyDefaults_ReturnsErrorOnUnparseableDefault(t *testing.T) {
	v := &badDefault{}
	err := defaulter.ApplyDefaults(v, nil)
	if err == nil {
		t.Fatal("expected an error for an unparseable default value targeting a non-string field, got nil")
	}
	if !strings.Contains(err.Error(), "Port") {
		t.Errorf("error %q does not mention the offending field name %q", err.Error(), "Port")
	}
}

// --- templated defaults ------------------------------------------------------------------
// NOTE: this assumes github.com/fmotalleb/go-tools/template supports
// standard Go text/template {{.Field}} syntax against the data argument.
// Adjust the template string below if the package's syntax differs.

type templateData struct {
	Env string
}

type withTemplatedDefault struct {
	Greeting string `default:"hello-{{.Env}}"`
}

func TestApplyDefaults_EvaluatesTemplatedDefaults(t *testing.T) {
	v := &withTemplatedDefault{}
	data := templateData{Env: "prod"}
	if err := defaulter.ApplyDefaults(v, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Greeting != "hello-prod" {
		t.Errorf("Greeting = %q, want %q (templated default was not evaluated against data)", v.Greeting, "hello-prod")
	}
}

// --- non-pointer input ------------------------------------------------------------------

func TestApplyDefaults_NonPointerInputDoesNotPanic(t *testing.T) {
	v := basic{}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ApplyDefaults panicked on non-pointer input: %v", r)
		}
	}()
	// A non-pointer struct is not addressable, so this is expected to be a
	// no-op (nothing settable), not an error - just confirming no panic.
	_ = defaulter.ApplyDefaults(v, nil)
}
