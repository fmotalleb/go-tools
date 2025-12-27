package future

import (
	"context"
	"errors"
	"sync"
)

// Any executes multiple functions concurrently and returns the result of the first one to complete successfully.
// If all functions return an error or the context is canceled, it returns an error.
//
// ctx: The context to control the execution of all functions.
// fns: A slice of functions to execute concurrently.
//
// Returns the result of the first function that completes successfully, or an error if all fail.
func Any[T any](ctx context.Context, fns ...func() (T, error)) (T, error) {
	if len(fns) == 0 {
		var zero T
		return zero, nil
	}

	results := make(chan T, 1) // Only need the first result
	errCh := make(chan error, len(fns))
	var wg sync.WaitGroup
	var once sync.Once

	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // Cancel all other goroutines once one has completed

	for _, fn := range fns {
		wg.Add(1)
		go func(fn func() (T, error)) {
			defer wg.Done()

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
				once.Do(func() {
					results <- res
				})
			case err := <-errSingleCh:
				errCh <- err
			case <-ctx.Done():
				// Don't add ctx.Err() to errors, to avoid race condition
			}
		}(fn)
	}

	go func() {
		wg.Wait()
		close(errCh)
		close(results)
	}()

	select {
	case res, ok := <-results:
		if ok {
			return res, nil
		}
		// If results is closed and empty, it means all goroutines failed.
		// Fall through to check for errors.
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}

	// If we are here, it means all functions resulted in an error.
	var firstErr error
	for err := range errCh {
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	if firstErr == nil {
		// This can happen if the context is canceled after all goroutines finished with errors
		if ctx.Err() != nil {
			firstErr = ctx.Err()
		} else {
			firstErr = errors.New("all functions failed")
		}
	}

	var zero T
	return zero, firstErr
}
