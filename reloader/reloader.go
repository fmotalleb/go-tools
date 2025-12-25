package reloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
)

var (
	ErrParentContextCanceled = errors.New("parent context killed")
	ErrReloadTimeout         = errors.New("reload timeout exceeded")
	ErrReloadChannelClosed   = errors.New("reload channel closed")
)

func WithOsSignal(
	parent context.Context,
	task func(context.Context) error,
	timeout time.Duration,
	signals ...os.Signal,
) error {
	reloadSig := make(chan os.Signal, 1)
	signal.Notify(reloadSig, os.Interrupt, syscall.SIGHUP)
	defer signal.Stop(reloadSig)
	return WithReload(
		parent,
		reloadSig,
		task,
		timeout,
	)
}

func WithReload[T any](
	parent context.Context,
	reload <-chan T,
	task func(context.Context) error,
	timeout time.Duration,
) error {
	if err := parent.Err(); err != nil {
		return err
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		err := handleTask(parent, task, reload, timer, timeout)
		if err != nil {
			return err
		}
	}
}

func handleTask[T any](
	parent context.Context,
	task func(context.Context) error,
	reload <-chan T,
	timer *time.Timer,
	timeout time.Duration,
) error {
	logger := log.Of(parent).Named("reloader")
	ctx, cancel := context.WithCancel(parent)
	errCh := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("task panic: %v", r)
			}
		}()
		errCh <- task(ctx)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			logger.Warn("task finished with error", zap.Error(err))
		} else {
			logger.Debug("task finished cleanly")
		}
		cancel()
		return err

	case <-parent.Done():
		err := errors.Join(ErrParentContextCanceled, parent.Err())
		logger.Warn("parent context is dead", zap.Error(err))
		cancel()
		return err

	case r, ok := <-reload:
		cancel()
		rLog := logger.Named("with-signal").WithLazy(zap.Any("signal", r))
		if !ok {
			rLog.Error("reload signal channel is closed", zap.Error(ErrReloadChannelClosed))
			return ErrReloadChannelClosed
		}
		rLog.Debug("reload signal received")
		select {
		case <-errCh:
			rLog.Debug(
				"task finished",
			)
			resetTimer(timer, timeout)
		case <-timer.C:
			rLog.Warn(
				"task did't finish after given timeout",
				zap.Duration("timeout", timeout),
			)
			return ErrReloadTimeout
		}
	}
	return nil
}

func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}
