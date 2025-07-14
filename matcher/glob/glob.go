package glob

import (
	"fmt"
	"path/filepath"
)

type Matcher struct {
	pattern string
}

func (m *Matcher) Match(s string) bool {
	result, err := filepath.Match(m.pattern, s)
	if err != nil {
		return false
	}
	return result
}

func Compile(exp string) (*Matcher, error) {
	_, err := filepath.Match(exp, "/")
	if err != nil {
		return nil, err
	}
	return &Matcher{exp}, nil
}

func MustCompile(exp string) *Matcher {
	return &Matcher{exp}
}

func (m *Matcher) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("glob:%s", m.pattern)), nil
}
