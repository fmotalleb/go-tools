package broadcast

// Subscribe to a broadcast and closes the broadcast one it finishes.
// Do not pass the channel to another goroutine since it will be closed if the callback return.
func Subscribe[T any](sub Subscription[T], callback func(<-chan T)) {
	index, ch := sub.Subscribe()
	defer func() {
		sub.Unsubscribe(index)
	}()
	callback(ch)
}
