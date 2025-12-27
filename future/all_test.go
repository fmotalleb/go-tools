package future_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/fmotalleb/go-tools/future"
)

func TestAll(t *testing.T) {
	t.Run("should return all results when all functions succeed", func(t *testing.T) {
		fn1 := func() (int, error) {
			time.Sleep(20 * time.Millisecond)
			return 1, nil
		}
		fn2 := func() (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 2, nil
		}
		fn3 := func() (int, error) {
			time.Sleep(30 * time.Millisecond)
			return 3, nil
		}

		results, err := future.All(context.Background(), fn1, fn2, fn3)
		assert.NoError(t, err)
		assert.Equal(t, []int{1, 2, 3}, results)
	})

	t.Run("should return error if one function fails", func(t *testing.T) {
		fn1 := func() (int, error) {
			return 1, nil
		}
		fn2 := func() (int, error) {
			return 0, errors.New("failed")
		}
		fn3 := func() (int, error) {
			return 3, nil
		}

		results, err := future.All(context.Background(), fn1, fn2, fn3)
		assert.Error(t, err)
		assert.True(t, results == nil)
	})

	t.Run("should return error when context is canceled before functions complete", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		fn1 := func() (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 1, nil
		}
		fn2 := func() (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 2, nil
		}

		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()

		results, err := future.All(ctx, fn1, fn2)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
		assert.True(t, results == nil)
	})

	t.Run("should return error when context times out", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		fn1 := func() (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 1, nil
		}
		fn2 := func() (int, error) {
			time.Sleep(20 * time.Millisecond)
			return 2, nil
		}

		results, err := future.All(ctx, fn1, fn2)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
		assert.True(t, results == nil)
	})

	t.Run("should handle empty slice of functions", func(t *testing.T) {
		results, err := future.All[int](context.Background())
		assert.NoError(t, err)
		assert.Equal(t, []int{}, results)
	})
}
