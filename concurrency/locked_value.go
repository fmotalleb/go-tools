package concurrency

import (
	"sync"
)

type (
	Operator[T any]    func(T) T
	LockedValue[T any] struct {
		value T
		mx    *sync.RWMutex
	}
)

func NewLockedValue[T any](initial T) *LockedValue[T] {
	return &LockedValue[T]{
		value: initial,
		mx:    &sync.RWMutex{},
	}
}

func (lv *LockedValue[T]) Get() T {
	lv.mx.RLock()
	defer lv.mx.RUnlock()
	return lv.value
}

func (lv *LockedValue[T]) Set(newValue T) {
	lv.mx.Lock()
	defer lv.mx.Unlock()
	lv.value = newValue
}

// Operate and modify the value in locked state

func (lv *LockedValue[T]) Operate(operator Operator[T]) {
	lv.mx.Lock()
	defer lv.mx.Unlock()
	lv.value = operator(lv.value)
}
