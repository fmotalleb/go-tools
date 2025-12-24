package concurrency

import (
	"sync"
)

type Pool[T any] struct {
	innerPool *sync.Pool
}

func NewPool[T any](generator func() T) *Pool[T] {
	p := &Pool[T]{
		innerPool: &sync.Pool{
			New: func() any {
				return generator()
			},
		},
	}
	return p
}

// Get retrieves an item from the pool if present.
func (p *Pool[T]) Get() T {
	item := p.innerPool.Get()
	t := item.(T)
	return t
}

// Put returns an item to the pool and signals availability.
func (p *Pool[T]) Put(item T) {
	p.innerPool.Put(item)
}

// Using item from the pool then return it back to the pool after [job] is done.
func (p *Pool[T]) Using(job func(T)) {
	item := p.Get()
	defer p.Put(item)
	job(item)
}
