package broadcast_test

import (
	"testing"
	"time"

	"github.com/FMotalleb/go-tools/broadcast"
	"github.com/FMotalleb/go-tools/log"
	"github.com/alecthomas/assert/v2"
	"go.uber.org/zap"
)

type payload = int

func broadCastGenerator(log *zap.Logger) (*broadcast.Broadcaster[payload], chan payload, <-chan payload, <-chan payload, <-chan payload) {
	b := broadcast.NewBroadcaster[payload](log)
	input := make(chan payload)
	_, sub1 := b.Subscribe()
	_, sub2 := b.Subscribe(5)
	_, sub3 := b.Subscribe(0)
	return b, input, sub1, sub2, sub3
}

func TestBroadcaster(t *testing.T) {
	log := log.NewBuilder().Silent().MustBuild()

	assertReceived := func(t *testing.T, ch <-chan payload, expected int, label string) {
		t.Helper()
		select {
		case val := <-ch:
			assert.Equal(t, expected, val, label)
		default:
			t.Errorf("%s did not receive broadcast", label)
		}
	}

	t.Run("MultipleSubscribersReceiveBroadcast", func(t *testing.T) {
		b, input, sub1, sub2, sub3 := broadCastGenerator(log)
		go b.Broadcast(input)

		input <- 42
		time.Sleep(50 * time.Millisecond)

		assertReceived(t, sub1, 42, "sub1")
		assertReceived(t, sub2, 42, "sub2")
		assertReceived(t, sub3, 42, "sub3")

		close(input)
	})

	t.Run("MultipleSubscribersReceiveBindClose", func(t *testing.T) {
		b, input, sub1, sub2, sub3 := broadCastGenerator(log)
		go b.BindTo(input)

		input <- 42
		time.Sleep(50 * time.Millisecond)

		assertReceived(t, sub1, 42, "sub1")
		assertReceived(t, sub2, 42, "sub2")
		assertReceived(t, sub3, 42, "sub3")

		close(input)
		<-sub1 // ensure sub1 closes
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		b := broadcast.NewBroadcaster[payload](log)
		input := make(chan payload)
		indx, sub := b.Subscribe()
		go b.BindTo(input)

		input <- 42
		time.Sleep(50 * time.Millisecond)

		assertReceived(t, sub, 42, "sub")

		assert.True(t, b.Unsubscribe(indx), "first unsubscribe should succeed")
		assert.False(t, b.Unsubscribe(indx), "second unsubscribe should fail")

		<-sub // ensure closed
	})

	t.Run("helpers.Subscribe", func(t *testing.T) {
		b := broadcast.NewBroadcaster[payload](log)
		input := make(chan payload)
		go b.BindTo(input)
		var received []int
		waitChan := make(chan struct{})

		go func() {
			waitChan <- struct{}{}
			broadcast.Subscribe(b, func(ch <-chan payload) {
				val := <-ch
				received = append(received, val)
				close(waitChan)
			})
		}()

		<-waitChan

		input <- 42

		<-waitChan

		assert.Equal(t, []int{42}, received)

		// No subscribers should remain
		assert.Equal(t, 0, b.SubscriberCount(), "should have 0 subscribers after Subscribe exits")

		close(input)
	})
}
