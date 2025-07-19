package builder

import (
	"fmt"
	"strings"

	"github.com/fmotalleb/go-tools/clone"
)

type Nested struct {
	Data map[string]any
	Sep  string
}

func NewNested(sep ...string) *Nested {
	separator := "."
	if len(sep) > 0 && sep[0] != "" {
		separator = sep[0]
	}
	return &Nested{
		Data: make(map[string]any),
		Sep:  separator,
	}
}

func (b *Nested) Clone() *Nested {
	return &Nested{
		Data: clone.Map(b.Data),
		Sep:  b.Sep,
	}
}

func (b *Nested) Set(key string, value any, sep ...string) *Nested {
	bb, _ := b.TrySet(key, value, sep...)
	return bb
}

func (b *Nested) TrySet(key string, value any, sep ...string) (*Nested, error) {
	separator := b.Sep
	if len(sep) > 0 && sep[0] != "" {
		separator = sep[0]
	}
	parts := strings.Split(key, separator)
	if err := makeDeepMap(parts, value, b.Data); err != nil {
		return nil, err
	}
	return b, nil
}

// makeDeepMap inserts value into nested maps at path keys.
func makeDeepMap(path []string, value any, src map[string]any) error {
	current := src
	for i, key := range path {
		if i == len(path)-1 {
			current[key] = value
			return nil
		}
		if next, exists := current[key]; !exists {
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		} else {
			m, ok := next.(map[string]any)
			if !ok {
				return fmt.Errorf("cannot descend into non-map at key %q in path %v", key, path)
			}
			current = m
		}
	}
	return nil
}
