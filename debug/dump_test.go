package debug_test

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/fmotalleb/go-tools/debug"
)

// test types used across multiple tests.
type person struct {
	Name string
	Age  int
}

type nested struct {
	Title string
	Meta  person
}

// captureLog captures log output written via log.Printf/Print/Println.
func captureLog(fn func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)
	fn()
	return buf.String()
}

// ── DumpJSON ────────────────────────────────────────────────────────────────────────

func TestDumpJSON(t *testing.T) {
	p := person{Name: "Alice", Age: 30}
	out := captureOutput(func() { debug.DumpJSON(p) })
	if !strings.Contains(out, `"Name"`) || !strings.Contains(out, "Alice") {
		t.Errorf("DumpJSON output missing expected fields:\n%s", out)
	}
}

// ── DumpYAML ────────────────────────────────────────────────────────────────────────

func TestDumpYAML(t *testing.T) {
	p := person{Name: "Bob", Age: 25}
	out := captureOutput(func() { debug.DumpYAML(p) })
	if !strings.Contains(out, "name: Bob") {
		t.Errorf("DumpYAML output missing expected content:\n%s", out)
	}
}

// ── DumpTOML ─────────────────────────────────────────────────────────────────────────

func TestDumpTOML(t *testing.T) {
	p := person{Name: "Charlie", Age: 35}
	out := captureOutput(func() { debug.DumpTOML(p) })
	if !strings.Contains(out, "Charlie") {
		t.Errorf("DumpTOML output missing expected content:\n%s", out)
	}
}

// ── DumpPretty ──────────────────────────────────────────────────────────────────────

func TestDumpPretty(t *testing.T) {
	p := person{Name: "Diana", Age: 28}
	out := captureOutput(func() { debug.DumpPretty(p) })
	if !strings.Contains(out, "Name:Diana") && !strings.Contains(out, "Name: Diana") {
		t.Errorf("DumpPretty output missing expected content:\n%s", out)
	}
}

// ── DumpCompactJSON ─────────────────────────────────────────────────────────────────

func TestDumpCompactJSON(t *testing.T) {
	p := person{Name: "Eve", Age: 22}
	out := captureOutput(func() { debug.DumpCompactJSON(p) })
	// Compact JSON is single-line; check the line has no leading whitespace and
	// contains the values.
	if strings.Contains(out, "\n  ") {
		t.Errorf("DumpCompactJSON output is indented (should be compact):\n%s", out)
	}
	if !strings.Contains(out, "Eve") {
		t.Errorf("DumpCompactJSON output missing value:\n%s", out)
	}
}

// ── DumpType ────────────────────────────────────────────────────────────────────────

func TestDumpType(t *testing.T) {
	out := captureOutput(func() { debug.DumpType(person{}) })
	if !strings.Contains(out, "debug_test.person") {
		t.Errorf("DumpType expected debug_test.person, got:\n%s", out)
	}
}

func TestDumpType_Builtin(t *testing.T) {
	out := captureOutput(func() { debug.DumpType(42) })
	if !strings.Contains(out, "int") {
		t.Errorf("DumpType expected int, got:\n%s", out)
	}
}

// ── DumpFields ──────────────────────────────────────────────────────────────────────

func TestDumpFields(t *testing.T) {
	p := person{Name: "Frank", Age: 40}
	out := captureOutput(func() { debug.DumpFields(p) })
	if !strings.Contains(out, "Name = Frank") || !strings.Contains(out, "Age = 40") {
		t.Errorf("DumpFields output missing expected content:\n%s", out)
	}
}

func TestDumpFields_NonStruct(t *testing.T) {
	out := captureOutput(func() { debug.DumpFields("hello") })
	if !strings.Contains(out, "<not a struct") {
		t.Errorf("DumpFields should report non-struct input:\n%s", out)
	}
}

func TestDumpFields_NilPointer(t *testing.T) {
	out := captureOutput(func() { debug.DumpFields((*person)(nil)) })
	if !strings.Contains(out, "<nil>") {
		t.Errorf("DumpFields should report nil pointer:\n%s", out)
	}
}

// ── DumpStack ───────────────────────────────────────────────────────────────────────

func TestDumpStack(t *testing.T) {
	out := captureOutput(func() { debug.DumpStack() })
	if !strings.Contains(out, "goroutine") {
		t.Errorf("DumpStack should contain goroutine header:\n%s", out)
	}
}

// ── DumpChan ────────────────────────────────────────────────────────────────────────

func TestDumpChan(t *testing.T) {
	ch := make(chan int, 3)
	ch <- 10
	ch <- 20
	ch <- 30

	out := captureOutput(func() { debug.DumpChan(ch) })
	if !strings.Contains(out, "10") || !strings.Contains(out, "20") || !strings.Contains(out, "30") {
		t.Errorf("DumpChan should print all buffered values:\n%s", out)
	}
}

func TestDumpChan_NonChan(t *testing.T) {
	out := captureOutput(func() { debug.DumpChan(42) })
	if !strings.Contains(out, "<not a channel") {
		t.Errorf("DumpChan should report non-channel input:\n%s", out)
	}
}

// ── Dump (smart) ────────────────────────────────────────────────────────────────────

func TestDump_Struct(t *testing.T) {
	p := person{Name: "Grace", Age: 32}
	out := captureOutput(func() { debug.Dump(p) })
	// Struct uses DumpPretty → %+v format.
	if !strings.Contains(out, "Name:Grace") && !strings.Contains(out, "Name: Grace") {
		t.Errorf("Dump on struct should use pretty format:\n%s", out)
	}
}

func TestDump_Slice(t *testing.T) {
	items := []string{"a", "b"}
	out := captureOutput(func() { debug.Dump(items) })
	// Slice uses DumpJSON → JSON format.
	if !strings.Contains(out, `"a"`) {
		t.Errorf("Dump on slice should use JSON:\n%s", out)
	}
}

func TestDump_Chan(t *testing.T) {
	ch := make(chan int, 1)
	ch <- 99
	out := captureOutput(func() { debug.Dump(ch) })
	if !strings.Contains(out, "99") {
		t.Errorf("Dump on channel should use DumpChan:\n%s", out)
	}
}

func TestDump_NilPtr(t *testing.T) {
	out := captureOutput(func() { debug.Dump((*person)(nil)) })
	if !strings.Contains(out, "<nil>") {
		t.Errorf("Dump on nil pointer should print <nil>:\n%s", out)
	}
}

// ── Sdump* string-returning variants ────────────────────────────────────────────────

func TestSdumpJSON(t *testing.T) {
	s := debug.SdumpJSON(person{Name: "Hank", Age: 50})
	if !strings.Contains(s, "Hank") {
		t.Errorf("SdumpJSON missing value, got:\n%s", s)
	}
}

func TestSdumpYAML(t *testing.T) {
	s := debug.SdumpYAML(person{Name: "Ivy", Age: 20})
	if !strings.Contains(s, "ivy") && !strings.Contains(s, "Ivy") {
		t.Errorf("SdumpYAML missing value, got:\n%s", s)
	}
}

func TestSdumpTOML(t *testing.T) {
	s := debug.SdumpTOML(person{Name: "Jack", Age: 45})
	if !strings.Contains(s, "Jack") {
		t.Errorf("SdumpTOML missing value, got:\n%s", s)
	}
}

func TestSdumpPretty(t *testing.T) {
	s := debug.SdumpPretty(person{Name: "Kate", Age: 33})
	if !strings.Contains(s, "Kate") {
		t.Errorf("SdumpPretty missing value, got:\n%s", s)
	}
}

func TestSdumpCompactJSON(t *testing.T) {
	s := debug.SdumpCompactJSON(person{Name: "Leo", Age: 27})
	if strings.Contains(s, "\n  ") {
		t.Errorf("SdumpCompactJSON should return compact JSON (no indentation):\n%s", s)
	}
	if !strings.Contains(s, "Leo") {
		t.Errorf("SdumpCompactJSON missing value:\n%s", s)
	}
}

// ── Log* variants ───────────────────────────────────────────────────────────────────

func TestLogJSON(t *testing.T) {
	out := captureLog(func() { debug.LogJSON(person{Name: "Mia", Age: 19}) })
	if !strings.Contains(out, "Mia") {
		t.Errorf("LogJSON missing value:\n%s", out)
	}
}

func TestLogYAML(t *testing.T) {
	out := captureLog(func() { debug.LogYAML(person{Name: "Noah", Age: 31}) })
	if !strings.Contains(out, "Noah") && !strings.Contains(out, "noah") {
		t.Errorf("LogYAML missing value:\n%s", out)
	}
}

func TestLogTOML(t *testing.T) {
	out := captureLog(func() { debug.LogTOML(person{Name: "Olivia", Age: 29}) })
	if !strings.Contains(out, "Olivia") {
		t.Errorf("LogTOML missing value:\n%s", out)
	}
}

func TestLogPretty(t *testing.T) {
	out := captureLog(func() { debug.LogPretty(person{Name: "Paul", Age: 41}) })
	if !strings.Contains(out, "Paul") {
		t.Errorf("LogPretty missing value:\n%s", out)
	}
}

func TestLogCompactJSON(t *testing.T) {
	out := captureLog(func() { debug.LogCompactJSON(person{Name: "Quinn", Age: 36}) })
	if !strings.Contains(out, "Quinn") {
		t.Errorf("LogCompactJSON missing value:\n%s", out)
	}
}

func TestLogType(t *testing.T) {
	out := captureLog(func() { debug.LogType(person{}) })
	if !strings.Contains(out, "debug_test.person") {
		t.Errorf("LogType missing type:\n%s", out)
	}
}

func TestLogFields(t *testing.T) {
	out := captureLog(func() { debug.LogFields(person{Name: "Ray", Age: 44}) })
	if !strings.Contains(out, "Ray") || !strings.Contains(out, "44") {
		t.Errorf("LogFields missing values:\n%s", out)
	}
}

func TestLogFields_NonStruct(t *testing.T) {
	out := captureLog(func() { debug.LogFields("hello") })
	if !strings.Contains(out, "not a struct") {
		t.Errorf("LogFields should report non-struct input:\n%s", out)
	}
}

// ── SprintFields ────────────────────────────────────────────────────────────────────

func TestSprintFields(t *testing.T) {
	s := debug.SprintFields(person{Name: "Sam", Age: 38})
	if !strings.Contains(s, "Name=Sam") || !strings.Contains(s, "Age=38") {
		t.Errorf("SprintFields missing content:\n%s", s)
	}
}

func TestSprintFields_NilPtr(t *testing.T) {
	s := debug.SprintFields((*person)(nil))
	if s != "<nil>" {
		t.Errorf("SprintFields on nil pointer should return <nil>, got: %s", s)
	}
}

func TestSprintFields_NonStruct(t *testing.T) {
	s := debug.SprintFields("hello")
	if !strings.Contains(s, "not a struct") {
		t.Errorf("SprintFields should indicate non-struct input:\n%s", s)
	}
}

// ── Error handling ──────────────────────────────────────────────────────────────────

func TestDumpJSON_Error(t *testing.T) {
	// Channels can't be JSON-marshalled; verify we don't panic.
	captureOutput(func() { debug.DumpJSON(make(chan int)) })
}

func TestSdumpJSON_Error(t *testing.T) {
	s := debug.SdumpJSON(make(chan int))
	if !strings.Contains(s, "<error:") {
		t.Errorf("SdumpJSON should return error string for non-marshalable types:\n%s", s)
	}
}

// captureOutput runs fn and returns everything written to stdout.
func captureOutput(fn func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = stdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}
