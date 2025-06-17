package concurrency

import (
	"context"
	"sync"
	"sync/atomic"
)

type Pool[T any] struct {
	items        chan T
	generator    func() T
	maxSize      int64
	currentSize  int64
	itemsCreated int64 // Track total items created
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	closed       int64 // atomic flag for closed state
}

func NewPool[T any](generator func() T, initialSize int) *Pool[T] {
	if initialSize < 1 {
		initialSize = 1
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &Pool[T]{
		items:     make(chan T, initialSize),
		generator: generator,
		maxSize:   int64(initialSize),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Pre-populate the pool
	pool.populate()

	return pool
}

// Get retrieves an item from the pool
func (p *Pool[T]) Get() T {
	select {
	case item := <-p.items:
		atomic.AddInt64(&p.currentSize, -1)
		return item
	default:
		// Pool is empty, check if we can generate new item
		maxSize := atomic.LoadInt64(&p.maxSize)
		itemsCreated := atomic.LoadInt64(&p.itemsCreated)

		if itemsCreated < maxSize {
			// We can generate a new item
			atomic.AddInt64(&p.itemsCreated, 1)
			return p.generator()
		}
		// Pool is at capacity, must wait for an item to be returned
		select {
		case item := <-p.items:
			atomic.AddInt64(&p.currentSize, -1)
			return item
		case <-p.ctx.Done():
			// Pool is closed, return zero value
			var zero T
			return zero
		}
	}
}

// Put returns an item to the pool
func (p *Pool[T]) Put(item T) {
	// Check if pool is closed
	if atomic.LoadInt64(&p.closed) == 1 {
		panic("cannot put item to closed pool")
	}

	select {
	case p.items <- item:
		atomic.AddInt64(&p.currentSize, 1)
	default:
		// Pool is full, discard the item
		// Don't decrement itemsCreated as the item still "exists"
	}
}

// populate fills the pool to its maximum size
func (p *Pool[T]) populate() {
	maxSize := atomic.LoadInt64(&p.maxSize)
	currentSize := atomic.LoadInt64(&p.currentSize)

	for i := currentSize; i < maxSize; i++ {
		select {
		case p.items <- p.generator():
			atomic.AddInt64(&p.currentSize, 1)
			atomic.AddInt64(&p.itemsCreated, 1)
		default:
			return // Pool is full
		}
	}
}

// SetSize changes the pool size at runtime
func (p *Pool[T]) SetSize(newSize int) {
	if newSize < 1 {
		newSize = 1
	}

	// Check if pool is closed
	if atomic.LoadInt64(&p.closed) == 1 {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	oldSize := int(atomic.LoadInt64(&p.maxSize))
	atomic.StoreInt64(&p.maxSize, int64(newSize))

	if newSize > oldSize {
		// Increase pool size - create new channel and populate
		newItems := make(chan T, newSize)

		// Transfer existing items
		close(p.items)

		for item := range p.items {
			select {
			case newItems <- item:
			default:
				break
			}
		}

		p.items = newItems
		p.populate()
	} else if newSize < oldSize {
		// Decrease pool size - create smaller channel
		newItems := make(chan T, newSize)

		// Transfer limited items
		transferred := 0
		close(p.items)
		for item := range p.items {
			if transferred < newSize {
				select {
				case newItems <- item:
					transferred++
				default:
					break
				}
			}
		}

		p.items = newItems
		atomic.StoreInt64(&p.currentSize, int64(transferred))
		// Reset itemsCreated to match new reality
		atomic.StoreInt64(&p.itemsCreated, int64(transferred))
	}
}

// GetSize returns current pool size limit
func (p *Pool[T]) GetSize() int {
	return int(atomic.LoadInt64(&p.maxSize))
}

// GetCurrentCount returns current number of items in pool
func (p *Pool[T]) GetCurrentCount() int {
	return int(atomic.LoadInt64(&p.currentSize))
}

// GetItemsCreated returns total number of items created
func (p *Pool[T]) GetItemsCreated() int {
	return int(atomic.LoadInt64(&p.itemsCreated))
}

// Close shuts down the pool and releases resources
func (p *Pool[T]) Close() {
	// Set closed flag atomically
	if !atomic.CompareAndSwapInt64(&p.closed, 0, 1) {
		return // Already closed
	}

	p.cancel()
	p.mu.Lock()
	defer p.mu.Unlock()

	close(p.items)
	// Drain remaining items
	for range p.items {
		// Just drain, items will be garbage collected
	}
}
