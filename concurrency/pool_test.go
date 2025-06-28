package concurrency_test

import (
	"sync"
	"testing"

	"github.com/FMotalleb/go-tools/concurrency"
	"github.com/alecthomas/assert/v2"
)

func TestNewPool(t *testing.T) {
	pool := concurrency.NewPool(func() int { return 42 })

	val := pool.Get()
	assert.Equal(t, val, 42)
}

func TestGetPut(t *testing.T) {
	pool := concurrency.NewPool(func() string { return "default" })

	// Put custom value
	pool.Put("hello")

	// Retrieve it back
	val := pool.Get()
	assert.Equal(t, val, "hello")
}

func TestUsing(t *testing.T) {
	pool := concurrency.NewPool(func() int { return 1 })

	var mu sync.Mutex
	var calledWith int

	pool.Using(func(v int) {
		mu.Lock()
		defer mu.Unlock()
		calledWith = v + 1
	})

	mu.Lock()
	assert.Equal(t, calledWith, 2)
	mu.Unlock()
}
