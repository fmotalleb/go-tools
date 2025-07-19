package writer

import (
	"io"
	"os"
)

type StrErr struct {
	io.Writer
}

func NewStdErr() io.Writer {
	return os.Stderr
}
