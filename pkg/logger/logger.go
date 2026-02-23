// Package logger provides structured, high-performance logging.
// We use Zap (Uber's logger) instead of standard library for:
// - Structured JSON logging
// - High performance (zero-allocation in hot paths)
// - Log levels (debug, info, warn, error)
package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log is the global logger instance.
// Use this throughout the application.
var Log *zap.Logger

// Field is our own type — wraps zap.Field
// callers use logger.Field, never zap.Field
type Field = zap.Field

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


// Sync flushes buffered log entries.
// Call explicitly after shutdown — do not rely solely on defer.
// Safe to call multiple times.
func Sync() {
	if Log == nil {
		return
	}
	// error is intentionally ignored — sync errors on stderr are benign
	_ = Log.Sync()
}


// Helper functions for common log levels
// ── level helpers ─────────────────────────────────────────────────────
func Debug(msg string, fields ...Field) { Log.Debug(msg, fields...) }
func Info(msg string, fields ...Field)  { Log.Info(msg, fields...) }
func Warn(msg string, fields ...Field)  { Log.Warn(msg, fields...) }
func Error(msg string, fields ...Field) { Log.Error(msg, fields...) }

// Fatal logs at error level, flushes, then exits.
// Does NOT use zap.Fatal — zap.Fatal calls os.Exit before Sync() runs.
// This version guarantees the log is flushed before process exits.
func Fatal(msg string, fields ...Field) {
	Log.Error(msg, fields...)
	Sync() // flush before exit — critical
	os.Exit(1)
}

// With creates a child logger with additional fields.
func With(fields ...Field) *zap.Logger {
	return Log.With(fields...)
}

// ── field constructors ────────────────────────────────────────────────
func String(key, val string) Field                 { return zap.String(key, val) }
func Int(key string, val int) Field                { return zap.Int(key, val) }
func Int32(key string, val int32) Field            { return zap.Int32(key, val) }
func Int64(key string, val int64) Field            { return zap.Int64(key, val) }
func Bool(key string, val bool) Field              { return zap.Bool(key, val) }
func Duration(key string, val time.Duration) Field { return zap.Duration(key, val) }
func Err(err error) Field                          { return zap.Error(err) }
func Any(key string, val interface{}) Field        { return zap.Any(key, val) }
