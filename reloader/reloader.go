package reloader

import (
	"context"
	"errors"
	"time"
)

var ErrParentContextCanceled = errors.New("parent context killed")
var ErrReloadTimeout = errors.New("reload timeout exceeded")

func WithReload(
	parent context.Context,
	reload <-chan any,
	task func(context.Context) error,
	timeout time.Duration,
) error {
	if err := parent.Err(); err != nil {
		return err
	}

	for {
		ctx, cancel := context.WithCancel(parent)
		errCh := make(chan error, 1)

		go func() {
			errCh <- task(ctx)
		}()

		select {
		case err := <-errCh:
			cancel()
			return err

		case <-parent.Done():
			cancel()
			return errors.Join(ErrParentContextCanceled, parent.Err())

		case <-reload:
			cancel()
			select {
			case <-errCh:
			case <-time.After(timeout):
				return ErrReloadTimeout
			}
			continue
		}
	}
}
