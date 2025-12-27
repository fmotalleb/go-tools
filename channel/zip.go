// Package channel provides utility functions for channels.
package channel

import "sync"

func Zip[T interface{}](buffer int, channels ...<-chan T) <-chan T {
	output := make(chan T, buffer)
	wg := new(sync.WaitGroup)
	for _, ch := range channels {
		wg.Go(func() {
			for i := range ch {
				output <- i
			}
		})
	}
	go func() {
		wg.Wait()
		close(output)
	}()
	return output
}
