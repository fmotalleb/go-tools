package future_test

import (
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/fmotalleb/go-tools/future"
)

func TestChannel(t *testing.T) {
	t.Run("it should call the get function", func(t *testing.T) {
		isCalled := false
		get := func() {
			isCalled = true
		}
		ch := future.Channel(get)
		<-ch
		assert.True(t, isCalled)
	})
	t.Run("it should be a non-blocking call", func(t *testing.T) {
		ch := future.Channel(func() {
			time.Sleep(100 * time.Millisecond)
		})
		assert.True(t, ch != nil)
	})

	t.Run("it should allow waiting on a waitgroup using a future", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(1)

		// Worker: simulates doing work and then calling Done.
		workerGet := func() {
			time.Sleep(50 * time.Millisecond)
			wg.Done()
		}
		future.Channel(workerGet) // We don't need the channel for the worker, just that it runs.

		// Assert that the future.Channel unblocks, meaning wg.Wait() completed.
		select {
		case <-future.Channel(wg.Wait):
			// Success: The future wrapping wg.Wait() completed.
		case <-time.After(100 * time.Millisecond):
			t.Fatal("The future wrapping wg.Wait() timed out.")
		}
	})
}
