package future

func Channel(get func()) <-chan *struct{} {
	ch := make(chan *struct{}, 1)
	go func() {
		get()
		ch <- new(struct{})
	}()
	return ch
}

func ChannelValue[T any](get func() T) <-chan T {
	ch := make(chan T, 1)
	go func() {
		ch <- get()
	}()
	return ch
}
