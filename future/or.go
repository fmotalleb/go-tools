package future

import (
	"context"
)

func Or[T any](ctx context.Context, get func() T, def T) T {
	ch := make(chan T, 1)

	go func() {
		ch <- get()
	}()

	select {
	case result := <-ch:
		return result
	case <-ctx.Done():
		return def
	}
}
