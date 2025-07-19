package writer

import (
	"errors"
	"io"
	"reflect"
	"strings"

	"github.com/FMotalleb/go-tools/log"
)

type Writer struct {
	writer
	definition any
}

type writer interface {
	io.Writer
}

func (w *Writer) Decode(from reflect.Type, val interface{}) (any, error) {
	var ty, path string
	params := make(map[string]any)
	switch from.Kind() {
	case reflect.String:
		strVal := val.(string)
		split := strings.SplitN(strVal, ",", 2)
		params["type"] = split[0]
		if len(split) == 2 {
			params["path"] = split[1]
		} else {
			params["path"] = ""
		}
	case reflect.Slice:
		arr := val.([]any)
		if len(arr) == 0 {
			break
		}
		params["type"] = arr[0]
		if len(arr) == 2 {
			params["path"] = arr[1]
		} else {
			params["path"] = ""
		}
	case reflect.Map:
		if from.Key().Kind() != reflect.String {
			return nil, errors.New("map key must be a string")
		}
		params = val.(map[string]interface{})
	default:
		return nil, errors.New("unsupported type for writer")
	}
	w.definition = val.(string)
	var ok bool
	ty, ok = params["type"].(string)
	if !ok {
		return nil, errors.New("`type` key is required in map, as string")
	}
	path, ok = params["path"].(string)
	if !ok {
		return nil, errors.New("`path` key is required in map, as string")
	}

	switch ty {
	case "stderr", "std", "":
		w.writer = NewStdErr()
	case "zap", "log":
		b := log.NewBuilder().FromEnv()
		if path != "" {
			b = b.Name(path)
		}
		l, err := b.Build()
		if err != nil {
			return nil, err
		}
		w.writer = NewZapWriter(l)
	case "rotate", "rotated", "file":
		w.writer = NewRotateWriter(RotateFileName(path))
	default:
		return nil, errors.New("unknown writer type: " + ty)
	}
	return w, nil
}
