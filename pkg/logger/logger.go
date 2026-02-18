// Package logger provides structured, high-performance logging.
// We use Zap (Uber's logger) instead of standard library for:
// - Structured JSON logging
// - High performance (zero-allocation in hot paths)
// - Log levels (debug, info, warn, error)
package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global logger instance.
// Use this throughout the application.
var Log *zap.Logger

// Init initializes the logger with configuration.
// Call this once at application startup.
func Init(level string, format string) error {
	// Parse log level from string
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Build configuration
	var cfg zap.Config
	if format == "json" {
		// JSON format for production (machine readable)
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		// Console format for development (human readable with colors)
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.Level = zap.NewAtomicLevelAt(zapLevel)

	// Build logger
	var err error
	Log, err = cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return err
	}

	return nil
}

// Sync flushes any buffered log entries.
// Call this before application exit.
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// With creates a child logger with additional fields.
// Use this to add context to logs (request_id, user_id, etc.)
func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// Helper functions for common log levels
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// Field constructors - wrap zap functions for convenience
func String(key, val string) zap.Field                 { return zap.String(key, val) }
func Int(key string, val int) zap.Field                { return zap.Int(key, val) }
func Int32(key string, val int32) zap.Field            { return zap.Int32(key, val) }
func Int64(key string, val int64) zap.Field            { return zap.Int64(key, val) }
func Bool(key string, val bool) zap.Field              { return zap.Bool(key, val) }
func Duration(key string, val time.Duration) zap.Field { return zap.Duration(key, val) }
func ErrorField(err error) zap.Field                   { return zap.Error(err) }
func Any(key string, val interface{}) zap.Field        { return zap.Any(key, val) }
