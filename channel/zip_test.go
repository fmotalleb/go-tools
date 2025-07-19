package channel_test

import (
	"sort"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/fmotalleb/go-tools/channel"
)

func TestZipChannel(t *testing.T) {
	t.Run("zip channel test", func(t *testing.T) {
		ch1 := make(chan int)
		ch2 := make(chan int)
		go func() {
			for i := 0; i < 5; i++ {
				ch1 <- i
			}
			close(ch1)
		}()
		go func() {
			for i := 0; i < 5; i++ {
				ch2 <- i
			}
			close(ch2)
		}()
		zipped := channel.Zip(ch1, ch2)
		ans := make([]int, 0)
		for val := range zipped {
			ans = append(ans, val)
		}
		sort.Ints(ans)
		assert.Equal(t, []int{0, 0, 1, 1, 2, 2, 3, 3, 4, 4}, ans)
	})
}
