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

// TestWithReload_ReloadTimeoutUnreached verifies that WithReload correctly handles
// a reload signal arriving during the reload timeout grace period without returning prematurely.
//
// Context:
// - reload timeout: a grace period before forcibly restarting a task that hasn't acknowledged cancellation.
// - This test simulates a task that is running when the reload timeout starts.
//
// Test Steps:
// 1. Start WithReload in a goroutine with a test context.
// 2. Wait for the first task start via startSig channel.
// 3. Sleep 5ms to simulate elapsed time, allowing the reload timeout to start counting.
// 4. Send a reload signal via the reload channel.
// 5. Wait for the second task start to confirm the reload actually restarted the task.
// 6. Check that WithReload has NOT returned unexpectedly by observing errCh.
//
// Channels:
// - reload: triggers task restart
// - errCh: signals unexpected return of WithReload
// - startSig: signals task start for liveness observation
//
// Notes:
//   - time.Sleep is necessary to simulate real elapsed time so that the reload arrives
//     after the grace period starts, reproducing the edge case.
//   - All select statements have bounded timeouts to prevent test hangs.
//   - The test ensures liveness, correct task restart, and proper reload timeout handling.
func TestWithReload_ReloadTimeoutUnreached(t *testing.T) {
	reload := make(chan bool, 2)
	errCh := make(chan error, 2)
	startSig := make(chan struct{}, 2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		errCh <- WithReload(
			ctx,
			reload,
			func(ctx context.Context) error {
				startSig <- struct{}{} // signal task start
				<-ctx.Done()
				return nil
			},
			time.Millisecond,
		)
	}()
	// Wait for first task start
	select {
	case <-startSig:
	case <-time.After(20 * time.Millisecond):
		t.Fatal("task did not start initially")
	}
	time.Sleep(5 * time.Millisecond)
	// Trigger reload
	reload <- true

	// Wait for second task start (reload worked)
	select {
	case err := <-errCh:
		t.Fatalf("WithReload returned unexpectedly: %v", err)
	case <-startSig:
		// success: task restarted
	case <-time.After(20 * time.Millisecond):
		t.Fatal("task was not restarted on reload")
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

func TestWithOsSignal_NoExplicitSignals(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskRun := make(chan struct{}, 1)
	task := func(ctx context.Context) error {
		select {
		case taskRun <- struct{}{}:
		default:
		}
		<-ctx.Done()
		return nil
	}

	// This go routine will run WithOsSignal in the background.
	// It should terminate when os.Interrupt is sent.
	errCh := make(chan error, 1)
	go func() {
		errCh <- WithOsSignal(parent, task, time.Second)
	}()

	// Wait for the task to start
	<-taskRun

	// Give a small moment for the signal.Notify to set up
	time.Sleep(100 * time.Millisecond)

	// Send an interrupt signal to the process.
	// This should be handled by WithOsSignal's signal.Notify,
	// which is listening to DefaultSignals.
	// os.Interrupt is part of DefaultSignals on both Unix and Windows.
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("Failed to find process: %v", err)
	}
	err = p.Signal(os.Interrupt)
	if err != nil {
		t.Fatalf("Failed to send signal: %v", err)
	}
	// After the interrupt, the reloader will restart the task.
	// We need to cancel the parent context to make the reloader exit.
	cancel()

	// Wait for WithOsSignal to return. It should return a `context.Canceled` error
	// because the parent context was canceled.
	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("WithOsSignal returned an unexpected error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WithOsSignal did not return in time after interrupt signal")
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
