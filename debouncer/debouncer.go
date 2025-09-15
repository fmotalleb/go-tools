package debouncer

// from: https://github.com/bep/debounce
import (
	"sync"
	"time"
)

func New() func(time.Duration, func()) {
	d := &callbackDebouncer{}
	return d.add
}

func NewStatic(after time.Duration) func(f func()) {
	add := New()
	return func(f func()) {
		add(after, f)
	}
}

type callbackDebouncer struct {
	mu    sync.Mutex
	timer *time.Timer
}

func (d *callbackDebouncer) add(after time.Duration, f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(after, f)
}
