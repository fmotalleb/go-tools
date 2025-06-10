package broadcast

import (
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

type Subscription[T any] interface {
	Subscribe(bufferSize ...int) (uint64, <-chan T)
	Unsubscribe(uint64) bool
}

type Broadcaster[T any] struct {
	logger      *zap.Logger
	lastItem    *atomic.Uint64
	lock        *sync.RWMutex
	subscribers map[uint64]chan T
}

func NewBroadcaster[T any](logger *zap.Logger) *Broadcaster[T] {
	initialMap := make(map[uint64]chan T, 0)
	return &Broadcaster[T]{
		logger:      logger,
		lastItem:    &atomic.Uint64{},
		lock:        new(sync.RWMutex),
		subscribers: initialMap,
	}
}

func (b *Broadcaster[T]) Subscribe(bufferSize ...int) (uint64, <-chan T) {
	size := 1
	if len(bufferSize) != 0 && bufferSize[0] != 0 {
		size = bufferSize[0]
	}
	l := b.logger.
		Named("Subscribe").
		With(zap.Int("buffer-size", size))
	ch := make(chan T, size)
	b.lock.Lock()
	defer b.lock.Unlock()
	index := b.lastItem.Add(1)
	b.subscribers[index] = ch
	l.Debug("subscription created", zap.Uint64("index", index))
	return index, ch
}

func (b *Broadcaster[T]) SubscriberCount() int {
	b.lock.RLock()
	defer b.lock.RUnlock()
	return len(b.subscribers)
}

func (b *Broadcaster[T]) Unsubscribe(index uint64) bool {
	l := b.logger.Named("Unsubscribe").With(zap.Uint64("index", index))
	b.lock.Lock()
	defer b.lock.Unlock()
	if ch, ok := b.subscribers[index]; ok {
		delete(b.subscribers, index)
		close(ch)
		l.Debug("unsubscribed successfully")
		return true
	}
	l.Debug("unsubscribe failed: index not found")
	return false
}

func (b *Broadcaster[T]) Broadcast(input <-chan T) {
	l := b.logger.Named("Broadcast")
	for val := range input {
		l.Debug("broadcasting value")
		b.Publish(val)
	}
	l.Debug("input channel closed, broadcast ended")
}

func (b *Broadcaster[T]) BindTo(input <-chan T) {
	l := b.logger.Named("BindTo")
	for val := range input {
		l.Debug("binding value to publish")
		b.Publish(val)
	}
	l.Debug("input channel closed, unsubscribing all")
	b.unsubscribeAll()
}

func (b *Broadcaster[T]) Publish(val T) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	for _, sub := range b.subscribers {
		// TODO: fix or clarify
		// Drain the subscriber channel if the listener released it without closing
		// A bad move but normally they will be deleted after this loop, the reason for this is
		// the unsubscribe also wants to lock but will fail if publish is in this loop
		// thus a channel may be waiting to be removed but that happens until after this loop
		// this is a Deadlock scenario and the fix has a Race condition scenario. currently the deadlock is ignored
		// since its hard to generate and its the developer's negligence if deadlock occurs.

		// Its possible to patch both issues if subs received an array instead of single item
		// this way we can drain sub, append new item to it then dispatch it again
		// _ = concurrency.Drain(sub)
		sub <- val
	}
}

func (b *Broadcaster[T]) unsubscribeAll() {
	l := b.logger.Named("unsubscribeAll")
	b.lock.Lock()
	defer b.lock.Unlock()
	count := len(b.subscribers)
	for _, ch := range b.subscribers {
		close(ch)
	}
	b.subscribers = make(map[uint64]chan T, 0)
	l.Debug("unsubscribed all", zap.Int("count", count))
}
