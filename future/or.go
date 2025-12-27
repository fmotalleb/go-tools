package future

import (
	"context"
)

// Or executes a function in a goroutine and returns its result.
// If the context is canceled before the function completes, it returns a default value.
//
// ctx: The context to control the execution.
// get: The function to execute.
// def: The default value to return if the context is canceled.
//
// Returns the result of the `get` function or the default value.
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
