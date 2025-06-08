package sysctx

import (
	"context"
	"os"
	"os/signal"
)

// CancelWith receives a context and cancel it after receiving os signal filtered using [sig]
func CancelWith(ctx context.Context, sig ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	// Set up a channel to listen for OS signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, sig...)

	// Goroutine to cancel the context when a signal is received
	go func() {
		<-signalChan
		cancel()
	}()
	return ctx
}
