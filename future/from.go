package future

import (
	"context"
	"time"
)

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
