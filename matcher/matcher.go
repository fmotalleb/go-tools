package matcher

import (
	"errors"
	"reflect"
	"strings"

	"github.com/FMotalleb/go-tools/matcher/glob"
)

type Matcher struct {
	matcher
}

type matcher interface {
	Match(string) bool
}

func (m *Matcher) Decode(from, _ reflect.Type, val interface{}) (any, error) {
	var ty, pat string
	switch from.Kind() {
	case reflect.String:
		str := val.(string)
		split := strings.Split(str, ":")
		switch len(split) {
		case 1:
			ty = "glob"
			pat = str
		default:
			ty = split[0]
			pat = strings.Join(split[1:], ":")
		}
	}
	switch ty {
	case "glob":
		var err error
		var mat matcher
		if mat, err = glob.Compile(pat); err != nil {
			return nil, err

		}
		m.matcher = mat
		return m, nil
	}

	return errors.New("failed to find matcher variant"), nil
}
