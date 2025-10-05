package ptrcmp

import "cmp"

func Or[T comparable](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return cmp.Or(*p, fallback)
}
