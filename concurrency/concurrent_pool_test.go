package concurrency

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
)

func TestPool_NewPool(t *testing.T) {
	t.Run("Create pool with valid initial size", func(t *testing.T) {
		counter := int64(0)
		generator := func() int {
			return int(atomic.AddInt64(&counter, 1))
		}

		pool := NewPool(generator, 3)
		assert.Equal(t, 3, pool.GetSize())
		assert.Equal(t, 3, pool.GetCurrentCount())
		pool.Close()
	})

	t.Run("Create pool with zero initial size defaults to 1", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, 0)
		assert.Equal(t, 1, pool.GetSize())
		assert.Equal(t, 1, pool.GetCurrentCount())
		pool.Close()
	})

	t.Run("Create pool with negative initial size defaults to 1", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, -5)
		assert.Equal(t, 1, pool.GetSize())
		assert.Equal(t, 1, pool.GetCurrentCount())
		pool.Close()
	})
}

func TestPool_GetPut(t *testing.T) {
	t.Run("Get and Put basic operations", func(t *testing.T) {
		counter := int64(0)
		generator := func() int {
			return int(atomic.AddInt64(&counter, 1))
		}

		pool := NewPool(generator, 2)

		// Get items
		item1 := pool.Get()
		assert.Equal(t, 1, pool.GetCurrentCount())

		item2 := pool.Get()
		assert.Equal(t, 0, pool.GetCurrentCount())

		// Put items back
		pool.Put(item1)
		assert.Equal(t, 1, pool.GetCurrentCount())

		pool.Put(item2)
		assert.Equal(t, 2, pool.GetCurrentCount())

		pool.Close()
	})

	t.Run("Get when pool is empty is blocked", func(t *testing.T) {
		counter := int64(0)
		generator := func() int {
			return int(atomic.AddInt64(&counter, 1))
		}

		pool := NewPool(generator, 1)

		// Exhaust pool
		item1 := pool.Get()
		assert.Equal(t, 0, pool.GetCurrentCount())

		// Get should generate new item since we're under max size
		item2 := pool.Get()
		assert.NotEqual(t, item1, item2)

		pool.Close()
	})

	t.Run("Put to full pool discards item", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, 1)

		// Fill pool
		item := pool.Get()
		pool.Put(item)
		assert.Equal(t, 1, pool.GetCurrentCount())

		// Try to put another item - should be discarded
		pool.Put("extra")
		assert.Equal(t, 1, pool.GetCurrentCount())

		pool.Close()
	})
}

func TestPool_SetSize(t *testing.T) {
	t.Run("Increase pool size", func(t *testing.T) {
		counter := int64(0)
		generator := func() int {
			return int(atomic.AddInt64(&counter, 1))
		}

		pool := NewPool(generator, 2)
		originalCount := pool.GetCurrentCount()

		pool.SetSize(5)
		assert.Equal(t, 5, pool.GetSize())
		assert.Equal(t, 5, pool.GetCurrentCount())
		assert.True(t, pool.GetCurrentCount() > originalCount)

		pool.Close()
	})

	t.Run("Decrease pool size", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, 5)

		pool.SetSize(2)
		assert.Equal(t, 2, pool.GetSize())
		assert.True(t, pool.GetCurrentCount() <= 2)

		pool.Close()
	})

	t.Run("Set size to zero defaults to 1", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, 3)

		pool.SetSize(0)
		assert.Equal(t, 1, pool.GetSize())

		pool.Close()
	})

	t.Run("Set negative size defaults to 1", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, 3)

		pool.SetSize(-2)
		assert.Equal(t, 1, pool.GetSize())

		pool.Close()
	})
}

func TestPool_Concurrency(t *testing.T) {
	t.Run("Concurrent Get and Put operations", func(t *testing.T) {
		counter := int64(0)
		generator := func() int {
			return int(atomic.AddInt64(&counter, 1))
		}

		pool := NewPool(generator, 10)

		var wg sync.WaitGroup
		numGoroutines := 50

		// Launch multiple goroutines doing Get/Put operations
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					item := pool.Get()
					time.Sleep(time.Microsecond) // Simulate work
					pool.Put(item)
				}
			}()
		}

		wg.Wait()
		pool.Close()
	})

	t.Run("Concurrent SetSize operations", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, 5)

		var wg sync.WaitGroup

		// Launch goroutines changing size
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(size int) {
				defer wg.Done()
				pool.SetSize(size%5 + 1)
			}(i)
		}

		// Launch goroutines doing Get/Put
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 5; j++ {
					item := pool.Get()
					pool.Put(item)
				}
			}()
		}

		wg.Wait()
		pool.Close()
	})
}

func TestPool_GeneratorCalls(t *testing.T) {
	t.Run("Generator called correct number of times", func(t *testing.T) {
		callCount := int64(0)
		generator := func() string {
			atomic.AddInt64(&callCount, 1)
			return "generated"
		}

		pool := NewPool(generator, 3)
		assert.Equal(t, int64(3), atomic.LoadInt64(&callCount))

		// Get all items
		pool.Get()
		pool.Get()
		pool.Get()

		// Get one more - should call generator
		pool.Get()
		assert.Equal(t, int64(4), atomic.LoadInt64(&callCount))

		pool.Close()
	})
}

func TestPool_Close(t *testing.T) {
	t.Run("Close pool properly", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, 3)

		// Use some items
		item1 := pool.Get()
		item2 := pool.Get()
		pool.Put(item1)

		// Close should not panic
		assert.NotPanics(t, pool.Close)

		// Verify context is cancelled
		select {
		case <-pool.ctx.Done():
			// Expected
		case <-time.After(time.Millisecond * 100):
			t.Fatal("Context should be cancelled after Close()")
		}

		// Put after close should panic
		assert.Panics(t, func() { pool.Put(item2) })
	})
}

func TestPool_EdgeCases(t *testing.T) {
	t.Run("Get blocks when pool empty and at max capacity", func(t *testing.T) {
		generator := func() string { return "test" }
		pool := NewPool(generator, 1)
		pool.SetSize(1)
		// Get the only item
		item := pool.Get()
		assert.Equal(t, 0, pool.GetCurrentCount())

		// Start goroutine that will block on Get
		done := make(chan bool)
		go func() {
			l := pool.Get() // This should block
			println(l)
			done <- true
		}()

		// Verify it's blocking
		select {
		case <-done:
			t.Fatal("Get should have blocked")
		case <-time.After(time.Millisecond * 50):
			// Expected - Get is blocking
		}

		// Put item back to unblock
		pool.Put(item)

		// Now Get should complete
		select {
		case <-done:
			// Expected
		case <-time.After(time.Millisecond * 100):
			t.Fatal("Get should have completed after Put")
		}

		pool.Close()
	})
}
