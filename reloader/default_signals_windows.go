//go:build windows

package reloader

import (
	"os"
)

var DefaultSignals = []os.Signal{
	os.Interrupt,
}
