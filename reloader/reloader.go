package reloader

import "context"

type Handler[T any] struct {
	reset     chan any
	generator func(context.Context) <-chan T
	handler   func(T)
}

func NewHandler[T any]() *Handler[T] {

	return &Handler[T]{}
}
