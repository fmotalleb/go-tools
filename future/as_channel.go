package future

func Channel(get func()) <-chan *struct{} {
	ch := make(chan *struct{}, 1)
	go func() {
		get()
		ch <- new(struct{})
	}()
	return ch
}
