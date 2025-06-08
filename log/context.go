package log

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerKey contextKey = "zap-logger"

// WithNewLogger builds logger and attach it to given context
func WithNewLogger(
	ctx context.Context,
	builders ...BuilderFunc,
) (context.Context, error) {
	b := NewBuilder()
	for _, builder := range builders {
		b = builder(b)
	}
	l, err := b.Build()
	if err != nil {
		return ctx, err
	}
	return WithLogger(ctx, l), nil
}

// WithNewLoggerForced does what [WithNewLogger] does but panics if fails
func WithNewLoggerForced(
	ctx context.Context,
	builders ...BuilderFunc,
) context.Context {
	b := NewBuilder()
	for _, builder := range builders {
		b = builder(b)
	}
	l := b.MustBuild()

	return WithLogger(ctx, l)
}

// WithNewEnvLogger builds logger using env variables and attach it to given context
func WithNewEnvLogger(
	ctx context.Context,
	builders ...BuilderFunc,
) (context.Context, error) {
	envBuilders := append([]BuilderFunc{
		func(b *Builder) *Builder {
			return b.FromEnv()
		},
	},
		builders...,
	)
	return WithNewLogger(
		ctx,
		envBuilders...,
	)
}

// WithNewLoggerForced does what [WithNewLogger] does but panics if fails
func WithNewEnvLoggerForced(
	ctx context.Context,
	builders ...BuilderFunc,
) context.Context {
	envBuilders := append([]BuilderFunc{
		func(b *Builder) *Builder {
			return b.FromEnv()
		},
	},
		builders...,
	)
	return WithNewLoggerForced(
		ctx,
		envBuilders...,
	)
}

// WithLogger adds logger to context
func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

// FromContext extracts logger from context, or returns a nop logger if it was not found
func FromContext(ctx context.Context) *zap.Logger {
	if logger, ok := ctx.Value(loggerKey).(*zap.Logger); ok {
		return logger
	}
	return zap.NewNop()
}
