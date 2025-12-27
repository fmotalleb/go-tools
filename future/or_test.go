package future_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/fmotalleb/go-tools/future"
)

func TestOr(t *testing.T) {
	t.Run("should return value from get function", func(t *testing.T) {
		result := future.Or(context.Background(), func() int {
			return 42
		}, 0)
		assert.Equal(t, 42, result)
	})

	t.Run("should return default value when context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		result := future.Or(ctx, func() int {
			time.Sleep(100 * time.Millisecond)
			return 42
		}, 100)
		assert.Equal(t, 100, result)
	})

	t.Run("should return default value when get function is slow and context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		result := future.Or(ctx, func() int {
			time.Sleep(100 * time.Millisecond)
			return 42
		}, 100)
		assert.Equal(t, 100, result)
	})

	t.Run("should return value when get function is fast enough", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		result := future.Or(ctx, func() int {
			time.Sleep(20 * time.Millisecond)
			return 42
		}, 100)
		assert.Equal(t, 42, result)
	})
}
