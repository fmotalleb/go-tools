package reloader

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
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
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	reload := make(chan bool, 1)
	var execCount int32

	task := func(ctx context.Context) error {
		atomic.AddInt32(&execCount, 1)
		<-ctx.Done()
		return nil
	}

	go func() {
		// This will run until the parent context is canceled.
		_ = WithReload(parent, reload, task, time.Second)
	}()

	// Trigger one reload.
	reload <- true

	// Give it a moment to process the reload and start the new task.
	time.Sleep(100 * time.Millisecond)

	// Stop the main loop.
	cancel()

	// Give the reloader a moment to exit.
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&execCount) <= 1 {
		t.Errorf("expected task to be executed more than once on reload, but got %d executions", execCount)
	}
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

func TestWithReload_TaskFinishesNormallyAsSingleShot(t *testing.T) {
	var counter int
	task := func(ctx context.Context) error {
		counter++
		return nil
	}

	// With the new implementation, WithReload should exit cleanly (return nil)
	// when the task finishes on its own.
	err := WithReload(context.Background(), make(<-chan bool), task, time.Second)

	if err != nil {
		t.Fatalf("Expected WithReload to exit with nil error, but got: %v", err)
	}

	if counter != 1 {
		t.Errorf("Expected task to be called exactly once, but it was called %d time(s)", counter)
	}
}

func TestResetTimer_WhenTimerExpired(t *testing.T) {
	// Create a timer with a very short duration.
	timer := time.NewTimer(1 * time.Millisecond)

	// Wait for the timer to expire to ensure t.Stop() will return false.
	<-timer.C

	// Now call resetTimer on the expired timer.
	// This should execute the `!t.Stop()` branch.
	newDuration := 5 * time.Millisecond
	resetTimer(timer, newDuration)

	// Check if the timer fires again after the new duration.
	select {
	case <-timer.C:
		// All good, the timer was reset correctly.
	case <-time.After(10 * time.Millisecond):
		t.Error("timer was not reset correctly after expiring")
	}
}
