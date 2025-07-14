package regexp

import (
	"fmt"
	"regexp"
)

type Matcher struct {
	*regexp.Regexp
}

func (m *Matcher) Match(s string) bool {
	return m.MatchString(s)
}

func Compile(exp string) (*Matcher, error) {
	m, err := regexp.Compile(exp)
	if err != nil {
		return nil, err
	}
	return &Matcher{m}, nil
}

func MustCompile(exp string) *Matcher {
	m := regexp.MustCompile(exp)
	return &Matcher{m}
}

func (m *Matcher) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("regex:%s", m.String())), nil
}
