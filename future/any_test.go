package future_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/fmotalleb/go-tools/future"
)

func TestAny(t *testing.T) {
	t.Run("should return the first successful result", func(t *testing.T) {
		fn1 := func() (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 1, errors.New("failed")
		}
		fn2 := func() (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 2, nil
		}
		fn3 := func() (int, error) {
			time.Sleep(20 * time.Millisecond)
			return 3, nil
		}

		result, err := future.Any(context.Background(), fn1, fn2, fn3)
		assert.NoError(t, err)
		assert.Equal(t, 2, result)
	})

	t.Run("should return error if all functions fail", func(t *testing.T) {
		fn1 := func() (int, error) {
			time.Sleep(10 * time.Millisecond)
			return 0, errors.New("failed 1")
		}
		fn2 := func() (int, error) {
			time.Sleep(20 * time.Millisecond)
			return 0, errors.New("failed 2")
		}

		_, err := future.Any(context.Background(), fn1, fn2)
		assert.Error(t, err)
	})

	t.Run("should return context canceled error if context is canceled before any success", func(t *testing.T) {
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

		_, err := future.Any(ctx, fn1, fn2)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.Canceled))
	})

	t.Run("should return context deadline exceeded error if context times out before any success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		fn1 := func() (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 1, nil
		}
		fn2 := func() (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 2, nil
		}

		_, err := future.Any(ctx, fn1, fn2)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, context.DeadlineExceeded))
	})

	t.Run("should handle empty slice of functions", func(t *testing.T) {
		result, err := future.Any[int](context.Background())
		assert.NoError(t, err)
		assert.Equal(t, 0, result) // Zero value for int
	})
}
