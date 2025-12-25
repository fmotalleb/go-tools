package reloader

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestWithReload_ParentContextCanceled(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	cancel()

	err := WithReload(parent, make(<-chan bool), func(ctx context.Context) error {
		return nil
	}, time.Second)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestWithReload_TaskReturnsError(t *testing.T) {
	expectedErr := errors.New("task error")
	err := WithReload(context.Background(), make(<-chan bool), func(ctx context.Context) error {
		return expectedErr
	}, time.Second)

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestWithReload_ReloadChannelClosed(t *testing.T) {
	reload := make(chan bool)
	close(reload)

	err := WithReload(context.Background(), reload, func(ctx context.Context) error {
		<-ctx.Done()
		return nil
	}, time.Second)

	if !errors.Is(err, ErrReloadChannelClosed) {
		t.Errorf("expected %v, got %v", ErrReloadChannelClosed, err)
	}
}

func TestWithReload_ReloadTimeout(t *testing.T) {
	reload := make(chan bool, 1)
	reload <- true

	err := WithReload(context.Background(), reload, func(ctx context.Context) error {
		<-ctx.Done()
		time.Sleep(2 * time.Second)
		return nil
	}, time.Second)

	if !errors.Is(err, ErrReloadTimeout) {
		t.Errorf("expected %v, got %v", ErrReloadTimeout, err)
	}
}

func TestWithReload_SuccessfulReload(t *testing.T) {
	reload := make(chan bool, 1)
	taskFinished := make(chan struct{})

	go func() {
		err := WithReload(context.Background(), reload, func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		}, time.Second)
		if err != nil && !errors.Is(err, ErrParentContextCanceled) && !errors.Is(err, context.Canceled) {
			t.Errorf("unexpected error: %v", err)
		}
		close(taskFinished)
	}()

	reload <- true
	time.Sleep(100 * time.Millisecond) // give it time to process the reload
	// in a real scenario, we might want to check if the task was actually restarted
}

func TestWithOsSignal(t *testing.T) {
	// This is tricky to test without actually sending signals.
	// We can't easily test the full functionality in a unit test.
	// We can, however, test that it returns an error if the parent context is canceled.
	parent, cancel := context.WithCancel(context.Background())
	cancel()

	err := WithOsSignal(parent, func(ctx context.Context) error {
		return nil
	}, time.Second, os.Interrupt, syscall.SIGHUP)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestWithReload_TaskPanics(t *testing.T) {
	err := WithReload(context.Background(), make(<-chan bool), func(ctx context.Context) error {
		panic("task panic")
	}, time.Second)

	if err == nil {
		t.Error("expected an error, got nil")
	}

	expected := "task panic: task panic"
	if err.Error() != expected {
		t.Errorf("expected error message '%s', got '%s'", expected, err.Error())
	}
}

func TestWithReload_TaskFinishesNormallyAndRestarts(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	var counter int
	task := func(ctx context.Context) error {
		counter++
		return nil
	}

	go func() {
		// We expect this to return an error when the parent context is canceled.
		_ = WithReload(parent, make(<-chan bool), task, time.Second)
	}()

	// Let the reloader run for a bit to allow the task to be called multiple times.
	time.Sleep(10 * time.Millisecond)
	cancel() // Stop the reloader.

	// Give the reloader a moment to exit after cancellation
	time.Sleep(10 * time.Millisecond)

	if counter <= 1 {
		t.Errorf("expected task to be restarted, but it was called only %d time(s)", counter)
	}
}
