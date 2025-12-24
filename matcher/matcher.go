package matcher

import (
	"errors"
	"reflect"
	"strings"

	"github.com/fmotalleb/go-tools/matcher/glob"
	"github.com/fmotalleb/go-tools/matcher/regexp"
	"github.com/fmotalleb/go-tools/matcher/wildcard"
)

type Matcher struct {
	matcher
}

type matcher interface {
	Match(string) bool
	MarshalText() ([]byte, error)
}

func (m *Matcher) Decode(from reflect.Type, val interface{}) (any, error) {
	var ty, pat string
	if val, ok := val.(string); ok {
		ty, pat = fromStr(val)
	}
	switch ty {
	case "wildcard", "domain", "wc":
		var err error
		var mat matcher
		if mat, err = wildcard.Compile(pat); err != nil {
			return nil, err
		}
		m.matcher = mat
		return m, nil
	case "glob", "file", "files":
		var err error
		var mat matcher
		if mat, err = glob.Compile(pat); err != nil {
			return nil, err
		}
		m.matcher = mat
		return m, nil
	case "regex", "regxp", "grep":
		var err error
		var mat matcher
		if mat, err = regexp.Compile(pat); err != nil {
			return nil, err
		}
		m.matcher = mat
		return m, nil
	}

	return nil, errors.New("failed to find matcher variant")
}

func fromStr(val interface{}) (string, string) {
	var ty, pat string
	str := val.(string)
	split := strings.Split(str, ":")
	switch len(split) {
	case 1:
		ty = "wildcard"
		pat = str
	default:
		ty = split[0]
		pat = strings.Join(split[1:], ":")
	}
	return ty, pat
}
