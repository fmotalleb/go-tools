//go:build unix

package reloader

import (
	"os"
	"syscall"
)

var DefaultSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGHUP,
}
