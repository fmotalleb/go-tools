package debouncer

import (
	"context"
	"sync"
	"time"

	"github.com/maniartech/signals"
)

type (
	// eventData holds the payload and the context of the last event
	// received during a debounce window.
	eventData[T any] struct {
		data T
		ctx  context.Context
	}

	// DebouncedSignal wraps a base signal and adds a combined
	// "immediate emit + trailing debounce" behavior.
	//
	// Rules:
	//   1. First event emits immediately.
	//   2. All events during the debounce window are collapsed into the latest.
	//   3. When the window expires, the latest event is sent.
	DebouncedSignal[T any] struct {
		// the underlying signal that will actually dispatch events
		next signals.Signal[T]

		// duration of the debounce window
		debounce time.Duration

		mu         sync.Mutex
		latest     *eventData[T] // last event received during the window
		cancelWait context.CancelFunc
	}
)

// NewDebouncedSignal creates a new wrapping signal that applies the debounce logic.
// If debounce <= 0, it panics.
func NewDebouncedSignal[T any](sig signals.Signal[T], debounce time.Duration) *DebouncedSignal[T] {
	if debounce <= 0 {
		panic("DebouncedSignal requires Debounce duration")
	}

	return &DebouncedSignal[T]{
		next:     sig,
		debounce: debounce,
	}
}

// AddListener forwards listener registration to the underlying signal.
func (s *DebouncedSignal[T]) AddListener(handler signals.SignalListener[T], key ...string) int {
	return s.next.AddListener(handler, key...)
}

// Emit performs debounced emission:
//
//   - If this is the first event after the previous window,
//     the event is emitted immediately and a new debounce
//     window is started.
//
//   - If a window is active, the event is stored as the "latest"
//     and will be emitted only after the window expires.
func (s *DebouncedSignal[T]) Emit(ctx context.Context, payload T) {
	s.mu.Lock()

	// Determine if the previous debounce window is over.
	if s.cancelWait == nil {
		// First emit in this cycle → immediate dispatch.
		s.next.Emit(ctx, payload)

		// Define the window end.
		s.latest = nil

		// Create cancellable wait context for trailing dispatch.
		waitCtx, cancel := context.WithCancel(context.Background())
		s.cancelWait = cancel

		// Start handler for trailing emit.
		go func() {
			s.waitAndDispatch(waitCtx)
			cancel()
			s.cancelWait = nil
		}()
		s.mu.Unlock()
		return
	}

	// Debounce window active → record last event.
	s.latest = &eventData[T]{
		data: payload,
		ctx:  ctx,
	}
	s.mu.Unlock()
}

// IsEmpty delegates to underlying signal.
func (s *DebouncedSignal[T]) IsEmpty() bool {
	return s.next.IsEmpty()
}

// Len delegates to underlying signal.
func (s *DebouncedSignal[T]) Len() int {
	return s.next.Len()
}

// RemoveListener delegates to underlying signal.
func (s *DebouncedSignal[T]) RemoveListener(key string) int {
	return s.next.RemoveListener(key)
}

// Reset clears all debounce state and cancels pending trailing dispatch.
func (s *DebouncedSignal[T]) Reset() {
	s.next.Reset()
	s.mu.Lock()
	if s.cancelWait != nil {
		s.cancelWait()
	}
	s.latest = nil
	s.mu.Unlock()
}

// waitAndDispatch waits the debounce duration. If the window expires,
// the latest event (if any) is dispatched (trailing-edge behavior).
func (s *DebouncedSignal[T]) waitAndDispatch(ctx context.Context) {
	for {
		select {
		case <-time.After(s.debounce):
			// Window expired.
			s.mu.Lock()
			latest := s.latest
			s.latest = nil

			// Set new cycle window start time.
			s.mu.Unlock()

			// Send trailing event if available.
			if latest != nil {
				s.next.Emit(latest.ctx, latest.data)
			} else {
				// exit cycle if nothing is waiting to be emitted, enabling the fast shot of the next emit
				return
			}
		case <-ctx.Done():
			// Reset or cancellation.
			return
		}
	}
}
