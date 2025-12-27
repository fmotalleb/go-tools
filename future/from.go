package future

import (
	"context"
	"time"
)

// From executes a function in a goroutine and returns its result.
// It can be canceled by the provided context or an optional timeout.
//
// ctx: The context to control the execution.
// get: The function to execute.
// timeout: An optional duration to set a timeout. If multiple timeouts are provided, only the first one is used.
//
// Returns the result of the `get` function or an error if the context is canceled or the timeout is reached.
func From[T any](ctx context.Context, get func() T, timeout ...time.Duration) (T, error) {
	internalContext := ctx
	if len(timeout) != 0 {
		timedCtx, cancel := context.WithTimeout(ctx, timeout[0])
		defer cancel()
		internalContext = timedCtx
	}

	ch := make(chan T, 1)

	go func() {
		ch <- get()
	}()

	select {
	case result := <-ch:
		return result, nil
	case <-internalContext.Done():
		var def T
		return def, internalContext.Err()
	}
}

// From executes a function in a goroutine and returns its result.
// It can be canceled by the provided context or an optional timeout.
//
// ctx: The context to control the execution.
// get: The function to execute.
// timeout: An optional duration to set a timeout. If multiple timeouts are provided, only the first one is used.
//
// Returns the result or error of the `get` function or an error if the context is canceled or the timeout is reached.
func FromErr[T any](ctx context.Context, get func() (T, error), timeout ...time.Duration) (T, error) {
	internalContext := ctx
	if len(timeout) != 0 {
		timedCtx, cancel := context.WithTimeout(ctx, timeout[0])
		defer cancel()
		internalContext = timedCtx
	}

	ch := make(chan T, 1)
	errCh := make(chan error, 1)

	go func() {
		res, err := get()
		if err != nil {
			errCh <- err
			return
		}
		ch <- res
	}()

	var def T
	select {
	case result := <-ch:
		return result, nil
	case err := <-errCh:
		return def, err
	case <-internalContext.Done():
		return def, internalContext.Err()
	}
}
