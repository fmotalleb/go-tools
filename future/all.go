package future

import (
	"context"
	"sync"
)

// All executes multiple functions concurrently.
// It returns a slice of results if all functions complete successfully.
// If any function returns an error or the context is canceled, it returns the first error encountered.
//
// ctx: The context to control the execution of all functions.
// fns: A slice of functions to execute concurrently.
//
// Returns a slice of results from all functions or the first error that occurred.
func All[T any](ctx context.Context, fns ...func() (T, error)) ([]T, error) {
	results := make([]T, len(fns))
	errors := make(chan error, len(fns)) // Buffered channel for errors
	var wg sync.WaitGroup

	for i, fn := range fns {
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
				results[i] = res
			case err := <-errSingleCh:
				errors <- err
			case <-ctx.Done():
				errors <- ctx.Err()
			}
		})
	}

	wg.Wait()
	close(errors)

	// Collect all errors, if any
	var firstErr error
	for err := range errors {
		if err != nil {
			if firstErr == nil {
				firstErr = err // Store the first error encountered
			}
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return results, nil
}
