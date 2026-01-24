//go:build unix

package reloader

import (
	"os"
	"syscall"
)

var DefaultSignals = []os.Signal{
	syscall.SIGHUP,
}
