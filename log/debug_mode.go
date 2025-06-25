//go:build debug

package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	DefaultLevel = zapcore.DebugLevel
	DefaultFormat = zapcore.CapitalColorLevelEncoder
	DefaultDevelopment = true
	DefaultSampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
}
