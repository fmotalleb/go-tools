package ctxtool

import (
	"context"
	"fmt"
	"reflect"
)

type ContextKey string

func Put[T any](ctx context.Context, item T) context.Context {
	name := getTypename(item)

	return context.WithValue(ctx, ctxKey("typed", name), item)
}

func Get[T any](ctx context.Context) T {
	var zero T // Default zero value for type T
	name := reflect.TypeOf(zero).String()
	value := ctx.Value(ctxKey("typed", name))
	if value == nil {
		return zero
	}

	// Type assertion to ensure the value is of type T
	castedValue, ok := value.(T)
	if !ok {
		return zero
	}
	return castedValue
}

func getTypename[T any](item T) string {
	return reflect.TypeOf(item).String()
}

func ctxKey(prefix string, key string) ContextKey {
	return ContextKey(fmt.Sprintf("%s:%s", prefix, key))
}
