# Zap Logger Builder Package

A fluent, builder-pattern wrapper for Uber's Zap logger with context integration, designed to make logger configuration simple and readable.

## Features

- 🚀 **Fluent Builder Pattern**: Chain methods for clean, readable configuration
- 🔧 **Context Integration**: Seamlessly attach loggers to Go contexts
- ⚡ **High Performance**: Built on top of Uber's Zap logger
- 🎯 **Type Safe**: Full type safety with compile-time checks
- 📦 **Zero Dependencies**: Only depends on `go.uber.org/zap`
- 🛠️ **Flexible Configuration**: Support for most Zap configuration options

## Installation

```bash
go get -u go.uber.org/zap
go get -u github.com/FMotalleb/log
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "github.com/FMotalleb/log"
)

func main() {
    // Create a simple logger
    logger := log.NewBuilder().
        Level("info").
        ServiceName("my-service").
        MustBuild()

    logger.Info("Hello, World!")

    // Create logger with context
    ctx := context.Background()
    ctx = log.WithNewLoggerForced(ctx,
        func(b *log.Builder) *log.Builder {
            return b.Level("debug").ServiceName("api-service")
        },
    )

    // Use logger from context
    log.FromContext(ctx).Info("Processing request")
}
```

### Development vs Production

```go
// Development logger (human-readable, colored output)
devLogger := log.NewBuilder().
    Development(true).
    Level("debug").
    ServiceName("my-app").
    MustBuild()

// Production logger (JSON, structured)
prodLogger := log.NewBuilder().
    Level("info").
    JSONEncoding().
    ServiceName("my-app").
    Version("1.0.0").
    Environment("production").
    MustBuild()
```

## Builder Methods

### Core Configuration

| Method | Description | Default |
|--------|-------------|---------|
| `Level(string)` | Set log level ("debug", "info", "warn", "error") | "info" |
| `Development(bool)` | Enable development mode (console + colors) | false |
| `ServiceName(string)` | Add service name to all logs | - |
| `Version(string)` | Add version to all logs | - |
| `Environment(string)` | Add environment to all logs | - |

### Encoding Options

| Method | Description |
|--------|-------------|
| `JSONEncoding()` | Use JSON format (production) |
| `ConsoleEncoding()` | Use console format (development) |
| `ColorLevel()` | Enable colored log levels |
| `LowercaseLevel()` | Use lowercase level names |
| `CapitalLevel()` | Use capital level names |

### Time Formatting

| Method | Description |
|--------|-------------|
| `ISO8601Time()` | ISO8601 timestamp format |
| `RFC3339Time()` | RFC3339 timestamp format |
| `EpochTime()` | Unix epoch timestamp |
| `CustomTime(layout)` | Custom time format |

### Output Configuration

| Method | Description |
|--------|-------------|
| `OutputPaths(paths...)` | Set output destinations |
| `AddOutputPath(path)` | Add output destination |
| `ErrorOutputPaths(paths...)` | Set error output destinations |
| `DisableCaller(bool)` | Disable caller information |
| `DisableStacktrace(bool)` | Disable stack traces |

### Advanced Options

| Method | Description |
|--------|-------------|
| `Sampling(initial, thereafter)` | Configure log sampling |
| `NoSampling()` | Disable log sampling |
| `AddHook(func)` | Add custom hook function |
| `Name(string)` | Set logger name |

## Context Integration

### Adding Logger to Context

```go
// Method 1: Build and attach logger
ctx, err := log.WithNewLogger(ctx,
    func(b *log.Builder) *log.Builder {
        return b.Level("info").ServiceName("api")
    },
)

// Method 2: Attach existing logger
logger := log.NewBuilder().Level("debug").MustBuild()
ctx = log.WithLogger(ctx, logger)

// Method 3: Build and attach (panic on error)
ctx = log.WithNewLoggerForced(ctx,
    func(b *log.Builder) *log.Builder {
        return b.Development(true).ServiceName("debug-service")
    },
)
```

### Using Logger from Context

```go
func processRequest(ctx context.Context, userID string) {
    logger := log.FromContext(ctx)
    
    logger.Info("Processing request",
        zap.String("user_id", userID),
        zap.String("action", "process"),
    )
    
    // Logger automatically includes service name, version, etc.
    // from the builder configuration
}
```

## Common Patterns

### HTTP Server with Request Logging

```go
func httpHandler(w http.ResponseWriter, r *http.Request) {
    // Create request-scoped logger
    ctx := log.WithNewLoggerForced(r.Context(),
        func(b *log.Builder) *log.Builder {
            return b.
                Level("info").
                ServiceName("api-server").
                AddInitialField("request_id", generateRequestID()).
                AddInitialField("method", r.Method).
                AddInitialField("path", r.URL.Path)
        },
    )
    
    logger := log.FromContext(ctx)
    logger.Info("Request started")
    
    // Process request with context
    processRequest(ctx)
    
    logger.Info("Request completed")
}
```

### File Logging

```go
fileLogger := log.NewBuilder().
    Level("warn").
    JSONEncoding().
    OutputPaths("/var/log/app.log").
    ErrorOutputPaths("/var/log/app-error.log").
    ServiceName("file-processor").
    MustBuild()
```

### Multiple Output Destinations

```go
multiLogger := log.NewBuilder().
    Level("info").
    OutputPaths("stdout", "/var/log/app.log", "syslog").
    ErrorOutputPaths("stderr", "/var/log/error.log").
    ServiceName("multi-output-service").
    MustBuild()
```

### Custom Hook for Monitoring

```go
monitoringLogger := log.NewBuilder().
    Level("info").
    AddHook(func(entry zapcore.Entry) error {
        if entry.Level >= zapcore.ErrorLevel {
            // Send to monitoring system (Sentry, DataDog, etc.)
            sendToMonitoring(entry)
        }
        return nil
    }).
    ServiceName("monitored-service").
    MustBuild()
```

## Configuration Examples

### Minimal Logger

```go
logger := log.NewBuilder().MustBuild()
```

### Development Logger

```go
logger := log.NewBuilder().
    Development(true).
    Level("debug").
    ServiceName("my-service").
    MustBuild()
```

### Production Logger

```go
logger := log.NewBuilder().
    Level("info").
    JSONEncoding().
    ISO8601Time().
    ServiceName("my-service").
    Version("1.2.3").
    Environment("production").
    Sampling(100, 100).
    MustBuild()
```

### Custom Field Keys

```go
logger := log.NewBuilder().
    TimeKey("ts").
    LevelKey("severity").
    MessageKey("msg").
    CallerKey("source").
    ServiceName("custom-service").
    MustBuild()
```

### Zap logger from environment vars

- **Namespace Prefix**: All variables use `ZAPLOG_` prefix to avoid conflicts
- **Simplified Configuration**: Focuses on core logging configuration options
- **Type Safety**: Boolean and integer values are properly parsed with fallbacks
- **Comma-Separated Lists**: Output paths support multiple destinations via comma separation

| Variable | Type | Description | Default | Valid Values |
|----------|------|-------------|---------|--------------|
| `ZAPLOG_LEVEL` | string | Logging level | "info" | "debug", "info", "warn", "error", "dpanic", "panic", "fatal" |
| `ZAPLOG_DEVELOPMENT` | bool | Development mode (console + colors) | false | "true", "false", "1", "0" |
| `ZAPLOG_TIME_FORMAT` | string | Timestamp format | "iso8601" | "iso8601", "rfc3339", "epoch", custom layout |
| `ZAPLOG_LEVEL_FORMAT` | string | Log level format | "lowercase" | "lowercase", "capital", "color" |
| `ZAPLOG_OUTPUT_PATHS` | []string | Output destinations (comma-separated) | "stdout" | "stdout", "/path/to/file", "syslog" |
| `ZAPLOG_ERROR_PATHS` | []string | Error output destinations (comma-separated) | "stderr" | "stderr", "/path/to/error.log" |
| `ZAPLOG_DISABLE_CALLER` | bool | Disable caller information | false | "true", "false", "1", "0" |
| `ZAPLOG_DISABLE_STACKTRACE` | bool | Disable stack traces | false | "true", "false", "1", "0" |
| `ZAPLOG_ENABLE_SAMPLING` | bool | Enable log sampling | false | "true", "false", "1", "0" |
| `ZAPLOG_SAMPLING_INITIAL` | int | Initial sampling rate | 100 | Any positive integer |
| `ZAPLOG_SAMPLING_THEREAFTER` | int | Subsequent sampling rate | 100 | Any positive integer |

### Configuration Examples

#### Development Environment

```bash
export ZAPLOG_LEVEL=debug
export ZAPLOG_DEVELOPMENT=true
export ZAPLOG_TIME_FORMAT=iso8601
export ZAPLOG_LEVEL_FORMAT=color
export ZAPLOG_OUTPUT_PATHS=stdout
export ZAPLOG_DISABLE_CALLER=false
export ZAPLOG_DISABLE_STACKTRACE=false
```

#### Production Environment

```bash
export ZAPLOG_LEVEL=info
export ZAPLOG_DEVELOPMENT=false
export ZAPLOG_TIME_FORMAT=iso8601
export ZAPLOG_LEVEL_FORMAT=lowercase
export ZAPLOG_OUTPUT_PATHS=stdout,/var/log/app.log
export ZAPLOG_ERROR_PATHS=stderr,/var/log/app-error.log
export ZAPLOG_ENABLE_SAMPLING=true
export ZAPLOG_SAMPLING_INITIAL=100
export ZAPLOG_SAMPLING_THEREAFTER=100
export ZAPLOG_DISABLE_CALLER=false
export ZAPLOG_DISABLE_STACKTRACE=false
```

#### Docker Compose Example

```yaml
version: '3.8'
services:
  app:
    image: my-app:latest
    environment:
      - ZAPLOG_LEVEL=info
      - ZAPLOG_DEVELOPMENT=false
      - ZAPLOG_TIME_FORMAT=iso8601
      - ZAPLOG_LEVEL_FORMAT=lowercase
      - ZAPLOG_OUTPUT_PATHS=stdout,/var/log/app.log
      - ZAPLOG_ERROR_PATHS=stderr,/var/log/error.log
      - ZAPLOG_ENABLE_SAMPLING=true
      - ZAPLOG_SAMPLING_INITIAL=100
      - ZAPLOG_SAMPLING_THEREAFTER=100
    volumes:
      - ./logs:/var/log
```

#### Usage in Go Code

```go
package main

import (
    "context"
    "os"
    "your-project/log"
)

func main() {
    // Set environment variables programmatically (for testing)
    os.Setenv("ZAPLOG_LEVEL", "debug")
    os.Setenv("ZAPLOG_DEVELOPMENT", "true")
    os.Setenv("ZAPLOG_TIME_FORMAT", "rfc3339")
    os.Setenv("ZAPLOG_LEVEL_FORMAT", "color")
    os.Setenv("ZAPLOG_OUTPUT_PATHS", "stdout,/tmp/app.log")
    
    // Create logger from environment
    logger := log.NewBuilder().FromEnv().MustBuild()
    logger.Info("Logger configured from ZAPLOG_ environment variables")
    
    // Create context with environment-configured logger
    ctx := context.Background()
    ctx = log.WithNewEnvLoggerForced(ctx)
    
    log.FromContext(ctx).Debug("Debug message with environment configuration")
}
```

### Variable Details

#### `ZAPLOG_TIME_FORMAT`

- **"iso8601"**: `2006-01-02T15:04:05.000Z0700` format
- **"rfc3339"**: `2006-01-02T15:04:05Z07:00` format  
- **"epoch"**: Unix timestamp (seconds since epoch)
- **Custom layout**: Any Go time layout string (e.g., `"2006-01-02 15:04:05"`)

#### `ZAPLOG_OUTPUT_PATHS` / `ZAPLOG_ERROR_PATHS`

- **"stdout"** / **"stderr"**: Standard output/error
- **File paths**: `/var/log/app.log`, `/tmp/debug.log`
- **"syslog"**: System log (Unix systems)
- **Multiple paths**: Comma-separated list: `"stdout,/var/log/app.log"`

#### `ZAPLOG_SAMPLING_*`

- Sampling reduces log volume in high-traffic scenarios
- `INITIAL`: Log first N messages per second
- `THEREAFTER`: Log every Nth message after initial quota
- Example: `INITIAL=100, THEREAFTER=100` logs first 100/sec, then every 100th message

## Error Handling

The builder provides two build methods:

- `Build()` - Returns `(*zap.Logger, error)` for explicit error handling
- `MustBuild()` - Returns `*zap.Logger` and panics on error (suitable for initialization)

```go
// Explicit error handling
logger, err := log.NewBuilder().Level("info").Build()
if err != nil {
    return fmt.Errorf("failed to create logger: %w", err)
}

// Panic on error (initialization time)
logger := log.NewBuilder().Level("info").MustBuild()
```

## Best Practices

1. **Use Context Integration**: Always pass loggers through context for request-scoped logging
2. **Set Service Metadata**: Include service name, version, and environment in production
3. **Choose Appropriate Levels**: Use debug for development, info for production
4. **Structure Your Logs**: Prefer structured fields over string formatting
5. **Handle Errors Gracefully**: Use `Build()` for runtime configuration, `MustBuild()` for initialization

## Performance Notes

- The builder creates a new instance for each method call (immutable pattern)
- Context operations are lightweight and safe for concurrent use
- `FromContext()` returns a no-op logger if none is found (safe to use anywhere)
- Zap's underlying performance characteristics are preserved

## License

This package wraps Uber's Zap logger. Please refer to Zap's license for usage terms.
