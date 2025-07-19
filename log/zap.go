package log

import (
	"fmt"
	"strings"
	"time"

	"github.com/fmotalleb/go-tools/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type BuilderFunc = func(*Builder) *Builder

var (
	DefaultLevel                           = zapcore.InfoLevel
	DefaultFormat                          = zapcore.LowercaseLevelEncoder
	DefaultDevelopment                     = false
	DefaultSampling    *zap.SamplingConfig = nil
)

func SetDebugDefaults() {
	DefaultLevel = zapcore.DebugLevel
	DefaultFormat = zapcore.CapitalColorLevelEncoder
	DefaultDevelopment = true
	DefaultSampling = &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	}
}

// Builder provides a fluent interface for configuring zap logger
type Builder struct {
	level             zapcore.Level
	development       bool
	disableCaller     bool
	disableStacktrace bool
	sampling          *zap.SamplingConfig
	encoding          string
	encoderConfig     zapcore.EncoderConfig
	outputPaths       []string
	errorOutputPaths  []string
	initialFields     map[string]interface{}
	hooks             []func(zapcore.Entry) error
	name              string
}

// NewBuilder creates a new LoggerBuilder with default values
func NewBuilder() *Builder {
	return &Builder{
		level:             DefaultLevel,
		development:       DefaultDevelopment,
		disableCaller:     false,
		disableStacktrace: false,
		sampling:          DefaultSampling,
		encoding:          "json",
		encoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    DefaultFormat,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		outputPaths:      []string{"stdout"},
		errorOutputPaths: []string{"stderr"},
		initialFields:    make(map[string]interface{}),
		hooks:            make([]func(zapcore.Entry) error, 0),
	}
}

// Level sets the logging level (Default info)
func (b *Builder) Level(level string) *Builder {
	builder := *b
	switch level {
	case "debug":
		builder.level = zapcore.DebugLevel
	case "info":
		builder.level = zapcore.InfoLevel
	case "warn":
		builder.level = zapcore.WarnLevel
	case "error":
		builder.level = zapcore.ErrorLevel
	case "dpanic":
		builder.level = zapcore.DPanicLevel
	case "panic":
		builder.level = zapcore.PanicLevel
	case "fatal":
		builder.level = zapcore.FatalLevel
	default:
		builder.level = zapcore.InfoLevel
	}
	return &builder
}

// LevelValue sets the logging level using zapcore.Level
func (b *Builder) LevelValue(level zapcore.Level) *Builder {
	builder := *b
	builder.level = level
	return &builder
}

// Development enables/disables development mode (console logger)
func (b *Builder) Development(dev bool) *Builder {
	builder := *b
	builder.development = dev
	if dev {
		builder.encoding = "console"
		builder.encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return &builder
}

// DisableCaller enables/disables caller information
func (b *Builder) DisableCaller(disable bool) *Builder {
	builder := *b
	builder.disableCaller = disable
	return &builder
}

// DisableStacktrace enables/disables stacktrace
func (b *Builder) DisableStacktrace(disable bool) *Builder {
	builder := *b
	builder.disableStacktrace = disable
	return &builder
}

// Sampling sets the sampling configuration
func (b *Builder) Sampling(initial int, thereafter int) *Builder {
	builder := *b
	builder.sampling = &zap.SamplingConfig{
		Initial:    initial,
		Thereafter: thereafter,
	}
	return &builder
}

// NoSampling disables sampling
func (b *Builder) NoSampling() *Builder {
	builder := *b
	builder.sampling = nil
	return &builder
}

// Encoding sets the log encoding format
func (b *Builder) Encoding(encoding string) *Builder {
	builder := *b
	builder.encoding = encoding
	return &builder
}

// JSONEncoding sets JSON encoding
func (b *Builder) JSONEncoding() *Builder {
	return b.Encoding("json")
}

// ConsoleEncoding sets console encoding
func (b *Builder) ConsoleEncoding() *Builder {
	return b.Encoding("console")
}

// TimeKey sets the time field key
func (b *Builder) TimeKey(key string) *Builder {
	builder := *b
	builder.encoderConfig.TimeKey = key
	return &builder
}

// LevelKey sets the level field key
func (b *Builder) LevelKey(key string) *Builder {
	builder := *b
	builder.encoderConfig.LevelKey = key
	return &builder
}

// NameKey sets the name field key
func (b *Builder) NameKey(key string) *Builder {
	builder := *b
	builder.encoderConfig.NameKey = key
	return &builder
}

// CallerKey sets the caller field key
func (b *Builder) CallerKey(key string) *Builder {
	builder := *b
	builder.encoderConfig.CallerKey = key
	return &builder
}

// MessageKey sets the message field key
func (b *Builder) MessageKey(key string) *Builder {
	builder := *b
	builder.encoderConfig.MessageKey = key
	return &builder
}

// StacktraceKey sets the stacktrace field key
func (b *Builder) StacktraceKey(key string) *Builder {
	builder := *b
	builder.encoderConfig.StacktraceKey = key
	return &builder
}

// LineEnding sets the line ending
func (b *Builder) LineEnding(ending string) *Builder {
	builder := *b
	builder.encoderConfig.LineEnding = ending
	return &builder
}

// EncodeLevel sets the level encoder
func (b *Builder) EncodeLevel(encoder zapcore.LevelEncoder) *Builder {
	builder := *b
	builder.encoderConfig.EncodeLevel = encoder
	return &builder
}

// LowercaseLevel sets lowercase level encoding
func (b *Builder) LowercaseLevel() *Builder {
	return b.EncodeLevel(zapcore.LowercaseLevelEncoder)
}

// CapitalLevel sets capital level encoding
func (b *Builder) CapitalLevel() *Builder {
	return b.EncodeLevel(zapcore.CapitalLevelEncoder)
}

// ColorLevel sets colored level encoding
func (b *Builder) ColorLevel() *Builder {
	return b.EncodeLevel(zapcore.CapitalColorLevelEncoder)
}

// EncodeTime sets the time encoder
func (b *Builder) EncodeTime(encoder zapcore.TimeEncoder) *Builder {
	builder := *b
	builder.encoderConfig.EncodeTime = encoder
	return &builder
}

// ISO8601Time sets ISO8601 time encoding
func (b *Builder) ISO8601Time() *Builder {
	return b.EncodeTime(zapcore.ISO8601TimeEncoder)
}

// RFC3339Time sets RFC3339 time encoding
func (b *Builder) RFC3339Time() *Builder {
	return b.EncodeTime(zapcore.RFC3339TimeEncoder)
}

// EpochTime sets epoch time encoding
func (b *Builder) EpochTime() *Builder {
	return b.EncodeTime(zapcore.EpochTimeEncoder)
}

// CustomTime sets custom time encoding
func (b *Builder) CustomTime(layout string) *Builder {
	return b.EncodeTime(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(layout))
	})
}

// EncodeDuration sets the duration encoder
func (b *Builder) EncodeDuration(encoder zapcore.DurationEncoder) *Builder {
	builder := *b
	builder.encoderConfig.EncodeDuration = encoder
	return &builder
}

// SecondsDuration sets seconds duration encoding
func (b *Builder) SecondsDuration() *Builder {
	return b.EncodeDuration(zapcore.SecondsDurationEncoder)
}

// NanosDuration sets nanoseconds duration encoding
func (b *Builder) NanosDuration() *Builder {
	return b.EncodeDuration(zapcore.NanosDurationEncoder)
}

// StringDuration sets string duration encoding
func (b *Builder) StringDuration() *Builder {
	return b.EncodeDuration(zapcore.StringDurationEncoder)
}

// EncodeCaller sets the caller encoder
func (b *Builder) EncodeCaller(encoder zapcore.CallerEncoder) *Builder {
	builder := *b
	builder.encoderConfig.EncodeCaller = encoder
	return &builder
}

// ShortCaller sets short caller encoding
func (b *Builder) ShortCaller() *Builder {
	return b.EncodeCaller(zapcore.ShortCallerEncoder)
}

// FullCaller sets full caller encoding
func (b *Builder) FullCaller() *Builder {
	return b.EncodeCaller(zapcore.FullCallerEncoder)
}

// OutputPaths sets the output paths
func (b *Builder) OutputPaths(paths ...string) *Builder {
	builder := *b
	builder.outputPaths = make([]string, len(paths))
	copy(builder.outputPaths, paths)
	return &builder
}

// AddOutputPath adds an output path
func (b *Builder) AddOutputPath(path string) *Builder {
	builder := *b
	builder.outputPaths = append(builder.outputPaths, path)
	return &builder
}

// ErrorOutputPaths sets the error output paths
func (b *Builder) ErrorOutputPaths(paths ...string) *Builder {
	builder := *b
	builder.errorOutputPaths = make([]string, len(paths))
	copy(builder.errorOutputPaths, paths)
	return &builder
}

// Silent log messages (mostly used for testing)
func (b *Builder) Silent() *Builder {
	builder := *b
	builder.outputPaths = make([]string, 0)
	builder.errorOutputPaths = make([]string, 0)
	return &builder
}

// AddErrorOutputPath adds an error output path
func (b *Builder) AddErrorOutputPath(path string) *Builder {
	builder := *b
	builder.errorOutputPaths = append(builder.errorOutputPaths, path)
	return &builder
}

// InitialFields sets initial fields
func (b *Builder) InitialFields(fields map[string]interface{}) *Builder {
	builder := *b
	builder.initialFields = make(map[string]interface{})
	for k, v := range fields {
		builder.initialFields[k] = v
	}
	return &builder
}

// AddInitialField adds an initial field
func (b *Builder) AddInitialField(key string, value interface{}) *Builder {
	builder := *b
	if builder.initialFields == nil {
		builder.initialFields = make(map[string]interface{})
	}
	builder.initialFields[key] = value
	return &builder
}

// ServiceName adds service name to initial fields
func (b *Builder) ServiceName(name string) *Builder {
	return b.AddInitialField("service", name)
}

// Version adds version to initial fields
func (b *Builder) Version(version string) *Builder {
	return b.AddInitialField("version", version)
}

// Environment adds environment to initial fields
func (b *Builder) Environment(env string) *Builder {
	return b.AddInitialField("environment", env)
}

// AddHook adds a hook function
func (b *Builder) AddHook(hook func(zapcore.Entry) error) *Builder {
	builder := *b
	builder.hooks = append(builder.hooks, hook)
	return &builder
}

// Name sets the logger name
func (b *Builder) Name(name string) *Builder {
	builder := *b
	builder.name = name
	return &builder
}

// Build creates the zap logger with the configured options
func (b *Builder) Build() (*zap.Logger, error) {
	config := zap.Config{
		Level:             zap.NewAtomicLevelAt(b.level),
		Development:       b.development,
		DisableCaller:     b.disableCaller,
		DisableStacktrace: b.disableStacktrace,
		Sampling:          b.sampling,
		Encoding:          b.encoding,
		EncoderConfig:     b.encoderConfig,
		OutputPaths:       b.outputPaths,
		ErrorOutputPaths:  b.errorOutputPaths,
		InitialFields:     b.initialFields,
	}

	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	// Add hooks if any
	if len(b.hooks) > 0 {
		logger = logger.WithOptions(zap.Hooks(b.hooks...))
	}

	// Set name if provided
	if b.name != "" {
		logger = logger.Named(b.name)
	}

	return logger, nil
}

// MustBuild creates the logger and panics on error
func (b *Builder) MustBuild() *zap.Logger {
	logger, err := b.Build()
	if err != nil {
		panic(err)
	}
	return logger
}

// FromEnv configures the builder using environment variables with sane defaults
func (b *Builder) FromEnv() *Builder {
	builder := *b

	// Core configuration
	level := env.Or("ZAPLOG_LEVEL", DefaultLevel.String())
	builder = *builder.Level(level)

	// Development mode
	isDev := env.BoolOr("ZAPLOG_DEVELOPMENT", DefaultDevelopment)
	if isDev {
		builder = *builder.Development(true)
	}

	// Time format
	timeFormat := env.Or("ZAPLOG_TIME_FORMAT", "iso8601")
	switch strings.ToLower(timeFormat) {
	case "iso8601":
		builder = *builder.ISO8601Time()
	case "rfc3339":
		builder = *builder.RFC3339Time()
	case "epoch":
		builder = *builder.EpochTime()
	default:
		if timeFormat != "" {
			builder = *builder.CustomTime(timeFormat)
		}
	}

	// Level encoding
	levelFormat := env.Or("ZAPLOG_LEVEL_FORMAT")
	switch strings.ToLower(levelFormat) {
	case "lowercase":
		builder = *builder.LowercaseLevel()
	case "capital":
		builder = *builder.CapitalLevel()
	case "color":
		builder = *builder.ColorLevel()
	}

	// Output paths
	if outputPaths := env.SliceOr("ZAPLOG_OUTPUT_PATHS", []string{"stdout"}); len(outputPaths) > 0 {
		builder = *builder.OutputPaths(outputPaths...)
	}

	if errorPaths := env.SliceOr("ZAPLOG_ERROR_PATHS", []string{"stderr"}); len(errorPaths) > 0 {
		builder = *builder.ErrorOutputPaths(errorPaths...)
	}

	// Caller and stacktrace
	if env.BoolOr("ZAPLOG_DISABLE_CALLER", false) {
		builder = *builder.DisableCaller(true)
	}

	if env.BoolOr("ZAPLOG_DISABLE_STACKTRACE", false) {
		builder = *builder.DisableStacktrace(true)
	}

	// Sampling configuration
	if env.BoolOr("ZAPLOG_ENABLE_SAMPLING", false) {
		initial := env.IntOr("ZAPLOG_SAMPLING_INITIAL", 100)
		thereafter := env.IntOr("ZAPLOG_SAMPLING_THEREAFTER", 100)
		builder = *builder.Sampling(initial, thereafter)
	}

	return &builder
}
