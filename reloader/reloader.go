package reloader

import (
	"context"
	"errors"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
)

var (
	ErrParentContextCanceled = errors.New("parent context killed")
	ErrReloadTimeout         = errors.New("reload timeout exceeded")
)

func WithReload[T any](
	parent context.Context,
	reload <-chan T,
	task func(context.Context) error,
	timeout time.Duration,
) error {
	if err := parent.Err(); err != nil {
		return err
	}
	logger := log.Of(parent).Named("reloader")
	for {
		ctx, cancel := context.WithCancel(parent)
		errCh := make(chan error, 1)

		go func() {
			errCh <- task(ctx)
		}()

		select {
		case err := <-errCh:
			logger.Warn("task finished with error", zap.Error(err))
			cancel()
			return err

		case <-parent.Done():
			err := errors.Join(ErrParentContextCanceled, parent.Err())
			logger.Warn("parent context is dead", zap.Error(err))
			cancel()
			return err

		case r := <-reload:
			logger = logger.Named("with-signal").WithLazy(zap.Any("signal", r))
			logger.Debug("reload signal received")
			cancel()
			select {
			case <-errCh:
				logger.Debug(
					"task finished",
				)
			case <-time.After(timeout):
				logger.Warn(
					"task did't finish after given timeout",
					zap.Duration("timeout", timeout),
				)
				return ErrReloadTimeout
			}
			continue
		}
	}
}
