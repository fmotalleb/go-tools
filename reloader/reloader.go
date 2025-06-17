package reloader

import (
	"context"
	"sync"
	"time"
)

type resetSignal struct{}

type Reloader[T any] struct {
	reset     chan resetSignal
	generator func(context.Context) <-chan T
	handler   func(T)
	mu        sync.Mutex
	running   bool
}

func New[T any](
	generator func(context.Context) <-chan T,
	handler func(T),
) *Reloader[T] {
	reset := make(chan resetSignal)
	return &Reloader[T]{
		reset:     reset,
		generator: generator,
		handler:   handler,
	}
}

// Reset worker this is blocking by design if you want to ask worker to reset use ResetTimeout.
func (h *Reloader[T]) Reset() {
	h.reset <- resetSignal{}
}

// ResetTimeout receives a [time.Duration] d and waits for given time trying to reset the worker.
// default value is a Millisecond
func (h *Reloader[T]) ResetTimeout(d ...time.Duration) bool {
	var waitTime time.Duration = time.Millisecond
	if len(d) != 0 {
		waitTime = d[0]
	}
	select {
	case h.reset <- resetSignal{}:
		return true
	case <-time.After(waitTime):
		return false
	}
}

// Start worker (has exclusive locking) and must be run in a goroutine.
func (h *Reloader[T]) Start(ctx context.Context) {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return
	}
	h.running = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		h.running = false
		h.mu.Unlock()
	}()

	for {
		// Create cancellable context for current generator
		genCtx, cancel := context.WithCancel(ctx)
		dataChan := h.generator(genCtx)

		// Process data until reset or context cancellation
	loop:
		for {
			select {
			case <-ctx.Done():
				cancel()
				return
			case <-h.reset:
				cancel()
				break loop // Break inner loop to restart generator
			case data, ok := <-dataChan:
				if !ok {
					cancel()
					return // Generator closed, exit completely
				}
				go h.handler(data)
			}
		}
	}
}

// Stop the handler.
func (h *Reloader[T]) Stop() {
	close(h.reset)
}
