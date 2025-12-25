# Reloader Package

The `reloader` package provides a robust mechanism for running tasks that can be gracefully reloaded without terminating the main process. It is ideal for long-running services that need to reload configuration but is also suitable for single-shot tasks.

## Core Concepts

The package's behavior is centered on how your task executes:

- **Reloading Long-Running Tasks**: If you provide a long-running task (e.g., a web server or a worker that blocks on `<-ctx.Done()`), the reloader will keep it running. When a reload signal is received (e.g., `SIGHUP`), the reloader cancels the task's context, waits for it to shut down gracefully, and then restarts it.

- **Single-Shot Execution**: If the task you provide finishes on its own (i.e., it returns without being canceled), the reloader will treat it as a completed single-shot task and exit cleanly.

- **Graceful Shutdown & Timeout**: In a reload scenario, the task's context is canceled. Your task should listen for this (`<-ctx.Done()`) to shut down cleanly. A timeout is enforced to prevent a faulty task from blocking a reload indefinitely.

## Usage Example (Long-Running Task)

The most common use case is to run a service that listens for OS signals. The `WithOsSignal` function provides a simple way to do this. This example demonstrates a long-running worker that will be restarted on a reload signal.

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/fmotalleb/go-tools/reloader"
)

// myWorker represents a long-running task.
// It must respect context cancellation to support graceful shutdown.
func myWorker(ctx context.Context) error {
	log.Println("Worker starting...")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Context was canceled (reload or parent shutdown).
			log.Println("Worker stopping gracefully...")
			time.Sleep(500 * time.Millisecond) // Simulate cleanup.
			log.Println("Worker finished cleanup.")
			return nil
		case t := <-ticker.C:
			log.Printf("Worker is running at %v\n", t)
		}
	}
}

func main() {
	ctx := context.Background()

	// Set a 5-second timeout for graceful shutdown during a reload.
	shutdownTimeout := 5 * time.Second

	log.Println("Starting reloader. Send SIGHUP to reload, or Ctrl+C to exit.")

	// WithOsSignal blocks until the task exits on its own, an error occurs,
	// or the parent context is canceled.
	err := reloader.WithOsSignal(ctx, myWorker, shutdownTimeout)
	if err != nil && err != context.Canceled {
		log.Fatalf("Reloader finished with an error: %v", err)
	}
	log.Println("Reloader shut down.")
}

```

## API

### `func WithOsSignal(...) error`

This is the primary entrypoint for most applications. It wraps `WithReload`, pre-configured to listen for OS signals (`SIGHUP`, `Interrupt` by default).

### `func WithReload[T any](...) error`

This is the core of the package. It is more generic and allows you to supply your own reload channel, which can be triggered by any event in your application.

## Error Handling

The reloader functions return an error to indicate a terminal condition. If a task finishes normally without an error, `nil` is returned.

- `ErrParentContextCanceled`: Returned if the main context is canceled.
- `ErrReloadTimeout`: Returned if a task fails to shut down within the specified timeout after a reload signal.
- `ErrReloadChannelClosed`: Returned if the custom reload channel provided to `WithReload` is closed.
