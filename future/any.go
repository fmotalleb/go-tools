package future

import (
	"context"
	"sync"
)

func Any[T any](ctx context.Context, fns ...func() (T, error)) (T, error) {
	if len(fns) == 0 {
		var zero T
		return zero, nil
	}

	results := make(chan T, len(fns))
	errors := make(chan error, len(fns))
	var wg sync.WaitGroup

	for _, fn := range fns {
		wg.Go(func() {
			resultSingleCh := make(chan T, 1)
			errSingleCh := make(chan error, 1)

			go func() {
				res, err := fn()
				if err != nil {
					errSingleCh <- err
					return
				}
				resultSingleCh <- res
			}()

			select {
			case res := <-resultSingleCh:
				results <- res
			case err := <-errSingleCh:
				errors <- err
			case <-ctx.Done():
				errors <- ctx.Err()
			}
		})
	}

	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	select {
	case res := <-results:
		return res, nil
	case err := <-errors:
		// Drain other potential errors from the channel
		go func() {
			for range errors {
			}
		}()
		var zero T
		return zero, err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}
