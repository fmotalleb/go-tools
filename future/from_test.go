package future_test

import (
	"context"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/fmotalleb/go-tools/future"
)

func TestFrom(t *testing.T) {
	t.Run("should return value before context is canceled", func(t *testing.T) {
		result, err := future.From(context.Background(), func() int {
			return 42
		})
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("should return error when context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := future.From(ctx, func() int {
			time.Sleep(100 * time.Millisecond)
			return 42
		})
		assert.Error(t, err)
	})

	t.Run("should return error when timeout is reached", func(t *testing.T) {
		_, err := future.From(context.Background(), func() int {
			time.Sleep(200 * time.Millisecond)
			return 42
		}, 100*time.Millisecond)
		assert.Error(t, err)
	})

	t.Run("should return value before timeout is reached", func(t *testing.T) {
		result, err := future.From(context.Background(), func() int {
			time.Sleep(50 * time.Millisecond)
			return 42
		}, 100*time.Millisecond)
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})
}
