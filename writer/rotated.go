package writer

import (
	"io"

	"github.com/natefinch/lumberjack"
)

type RotateOpt = func(l *lumberjack.Logger) *lumberjack.Logger

// Rotated is a custom writer that handles log rotation.
type Rotated struct {
	io.Writer
}

func NewRotateWriter(opts ...RotateOpt) io.Writer {
	lumberjackLogger := &lumberjack.Logger{}
	for _, opt := range opts {
		lumberjackLogger = opt(lumberjackLogger)
	}
	return &Rotated{
		lumberjackLogger,
	}
}

func RotateMaxSize(sizeMb int) RotateOpt {
	return func(l *lumberjack.Logger) *lumberjack.Logger {
		l.MaxSize = sizeMb
		return l
	}
}

func RotateMaxAge(days int) RotateOpt {
	return func(l *lumberjack.Logger) *lumberjack.Logger {
		l.MaxAge = days
		return l
	}
}

func RotateCompress(compress bool) RotateOpt {
	return func(l *lumberjack.Logger) *lumberjack.Logger {
		l.Compress = compress
		return l
	}
}

func RotateLocalTime(localTime bool) RotateOpt {
	return func(l *lumberjack.Logger) *lumberjack.Logger {
		l.LocalTime = localTime
		return l
	}
}

func RotateFileName(name string) RotateOpt {
	return func(l *lumberjack.Logger) *lumberjack.Logger {
		l.Filename = name
		return l
	}
}
