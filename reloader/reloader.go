package reloader

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/fmotalleb/go-tools/log"
	"go.uber.org/zap"
)

var (
	// ErrParentContextCanceled is returned when the main context provided to the reloader is canceled.
	ErrParentContextCanceled = errors.New("parent context killed")
	// ErrReloadTimeout is returned when a reload signal is received, but the running task
	// does not finish within the specified timeout duration.
	ErrReloadTimeout = errors.New("reload timeout exceeded")
	// ErrReloadChannelClosed is returned if the reload signal channel is closed,
	// which terminates the reloader.
	ErrReloadChannelClosed = errors.New("reload channel closed")
)

// WithOsSignal is a convenience wrapper around WithReload that listens for specific OS signals
// to trigger a reload.
//
// Parameters:
//   - parent: The main context for the entire operation. If this context is canceled,
//     the function will terminate.
//   - task: The function to execute. For long-running services, this function should block until
//     its context is canceled. If the task returns on its own, WithReload will exit.
//   - timeout: The grace period for the task to shut down after a reload signal.
//   - signals: An optional list of OS signals to listen for. If not provided, it defaults
//     to os.Interrupt and syscall.SIGHUP.
func WithOsSignal(
	parent context.Context,
	task func(context.Context) error,
	timeout time.Duration,
	signals ...os.Signal,
) error {
	reloadSig := make(chan os.Signal, 1)
	signal.Notify(reloadSig, signals...)
	defer signal.Stop(reloadSig)
	return WithReload(
		parent,
		reloadSig,
		task,
		timeout,
	)
}

// WithReload runs a task and handles reloading it based on signals from a channel.
// It is designed for long-running tasks that block until their context is canceled.
//
// The function's behavior depends on how the task finishes:
//  1. If the task finishes on its own (returns nil or an error), WithReload will exit.
//     This is the "single-shot" behavior.
//  2. If a reload signal is received, WithReload will cancel the task's context and wait
//     for it to finish. If it finishes gracefully, the loop continues, and the task is restarted.
//
// The function terminates if:
// 1. the parent context is canceled
// 2. the reload channel is closed,
// 3. the task returns an error
// 4. a reload times out.
//
// Parameters:
//   - parent: The main context. Cancellation will terminate the reloader.
//   - reload: A channel that, when it receives a value, triggers a graceful shutdown
//     and restart of the task.
//   - task: The function to execute.
//   - timeout: The grace period for the task to finish after a reload signal.
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
		err, shouldExit := handleTask(parent, task, reload, timer, timeout)
		if err != nil {
			return err
		}
		if shouldExit {
			return nil
		}
	}
}

// handleTask manages a single lifecycle of the task and returns a boolean indicating
// whether the WithReload loop should exit.
// It returns `(error, true)` for terminal conditions, and `(nil, false)` for a successful reload.
func handleTask[T any](
	parent context.Context,
	task func(context.Context) error,
	reload <-chan T,
	timer *time.Timer,
	timeout time.Duration,
) (error, bool) {
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
		// The task finished on its own. This is a terminal event.
		// Return the task's error and signal that the loop should exit.
		if err != nil {
			logger.Warn("task finished with error", zap.Error(err))
		} else {
			logger.Debug("task finished cleanly")
		}
		cancel()
		return err, true

	case <-parent.Done():
		// The parent context was canceled. This is a hard stop.
		err := errors.Join(ErrParentContextCanceled, parent.Err())
		logger.Warn("parent context is dead", zap.Error(err))
		cancel()
		return err, true

	case r, ok := <-reload:
		// A reload signal was received.
		cancel() // Signal the task to shut down.
		rLog := logger.Named("with-signal").WithLazy(zap.Any("signal", r))
		if !ok {
			rLog.Error("reload signal channel is closed", zap.Error(ErrReloadChannelClosed))
			return ErrReloadChannelClosed, true
		}
		rLog.Debug("reload signal received")
		// Wait for the task to finish, but only for the timeout duration.
		select {
		case <-errCh:
			// Task finished gracefully. Return (nil, false) to continue the loop.
			rLog.Debug(
				"task finished in grace window",
			)
			resetTimer(timer, timeout)
			return nil, false
		case <-timer.C:
			// The task did not finish in time. This is a terminal event.
			rLog.Warn(
				"task did't finish after given timeout",
				zap.Duration("timeout", timeout),
			)
			return ErrReloadTimeout, true
		}
	}
}

// resetTimer safely resets a timer. It stops the timer and drains the channel
// if the timer has already fired.
func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
	t.Reset(d)
}
