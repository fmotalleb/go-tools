package debug

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"go.yaml.in/yaml/v3"
)

// DumpJSON marshals item to indented JSON and prints it.
func DumpJSON(item any) {
	s, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(string(s))
}

// DumpYAML marshals item to YAML and prints it.
func DumpYAML(item any) {
	s, err := yaml.Marshal(item)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(string(s))
}

// DumpTOML marshals item to TOML and prints it.
func DumpTOML(item any) {
	s, err := toml.Marshal(item)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(string(s))
}

// DumpPretty prints the item using %+v (struct fields with names).
func DumpPretty(item any) {
	fmt.Printf("%+v\n", item)
}

// DumpCompactJSON marshals item to single-line JSON and prints it.
func DumpCompactJSON(item any) {
	s, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		return
	}
	fmt.Println(string(s))
}

// DumpType prints the concrete Go type of item.
func DumpType(item any) {
	fmt.Printf("%T\n", item)
}

// DumpFields reflectively enumerates the exported fields of a struct and prints
// each field in "name=value" form, one per line.
func DumpFields(item any) {
	v := reflect.ValueOf(item)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			fmt.Println("<nil>")
			return
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		fmt.Printf("<not a struct: %T>\n", item)
		return
	}

	t := v.Type()
	for i := range v.NumField() {
		fv := v.Field(i)
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}
		fmt.Printf("  %s = %+v\n", ft.Name, fv.Interface())
	}
}

// DumpStack prints the current goroutine stack trace.
func DumpStack() {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	os.Stdout.Write(buf[:n])
}

// DumpChan reads all currently available values from a channel (non-blocking)
// and prints them as JSON lines (one per value). It does NOT close the channel.
func DumpChan(ch any) {
	v := reflect.ValueOf(ch)
	if v.Kind() != reflect.Chan {
		fmt.Printf("<not a channel: %T>\n", ch)
		return
	}

	for {
		cases := []reflect.SelectCase{
			{Dir: reflect.SelectRecv, Chan: v},
			{Dir: reflect.SelectDefault},
		}
		chosen, recv, ok := reflect.Select(cases)
		if chosen == 1 {
			break
		}
		if !ok {
			fmt.Println("<channel closed>")
			break
		}
		DumpJSON(recv.Interface())
	}
}

// Dump is a smart dump that auto-selects a format:
//   - Structs → DumpPretty (for a compact struct-field overview)
//   - Maps/Slices → DumpJSON
//   - Channels → DumpChan
//   - Everything else → DumpPretty
func Dump(item any) {
	v := reflect.ValueOf(item)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			fmt.Println("<nil>")
			return
		}
		v = v.Elem()
		item = v.Interface()
	}

	switch v.Kind() {
	case reflect.Struct:
		DumpPretty(item)
	case reflect.Map, reflect.Slice, reflect.Array:
		DumpJSON(item)
	case reflect.Chan:
		DumpChan(item)
	default:
		DumpPretty(item)
	}
}

// SdumpJSON returns the indented JSON representation of item.
func SdumpJSON(item any) string {
	s, err := json.MarshalIndent(item, "", "  ")
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(s)
}

// SdumpYAML returns the YAML representation of item.
func SdumpYAML(item any) string {
	s, err := yaml.Marshal(item)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(s)
}

// SdumpTOML returns the TOML representation of item.
func SdumpTOML(item any) string {
	s, err := toml.Marshal(item)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(s)
}

// SdumpPretty returns the item as a %+v formatted string.
func SdumpPretty(item any) string {
	return fmt.Sprintf("%+v", item)
}

// SdumpCompactJSON returns the single-line JSON representation of item.
func SdumpCompactJSON(item any) string {
	s, err := json.Marshal(item)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return string(s)
}

// LogJSON logs the indented JSON representation of item.
func LogJSON(item any) {
	log.Printf("DEBUG: %s", SdumpJSON(item))
}

// LogYAML logs the YAML representation of item.
func LogYAML(item any) {
	log.Printf("DEBUG: %s", SdumpYAML(item))
}

// LogTOML logs the TOML representation of item.
func LogTOML(item any) {
	log.Printf("DEBUG: %s", SdumpTOML(item))
}

// LogPretty logs the %+v representation of item.
func LogPretty(item any) {
	log.Printf("DEBUG: %+v", item)
}

// LogCompactJSON logs the single-line JSON representation of item.
func LogCompactJSON(item any) {
	log.Printf("DEBUG: %s", SdumpCompactJSON(item))
}

// LogType logs the concrete Go type of item.
func LogType(item any) {
	log.Printf("DEBUG type: %T", item)
}

// LogFields reflectively logs each exported field of a struct, one per log line.
func LogFields(item any) {
	v := reflect.ValueOf(item)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			log.Print("DEBUG fields: <nil>")
			return
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		log.Printf("DEBUG fields: <not a struct: %T>", item)
		return
	}
	t := v.Type()
	for i := range v.NumField() {
		fv := v.Field(i)
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}
		log.Printf("DEBUG field %s = %+v", ft.Name, fv.Interface())
	}
}

// SprintFields returns a single-line string of all exported struct fields in
// "name=value" form, joined by ", ".
func SprintFields(item any) string {
	v := reflect.ValueOf(item)
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "<nil>"
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fmt.Sprintf("<not a struct: %T>", item)
	}
	t := v.Type()
	var parts []string
	for i := range v.NumField() {
		fv := v.Field(i)
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%+v", ft.Name, fv.Interface()))
	}
	return strings.Join(parts, ", ")
}
