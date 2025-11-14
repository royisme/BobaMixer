// Package logger provides structured logging with automatic rotation and sensitive data masking.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	globalLogger *zap.Logger
	sugarLogger  *zap.SugaredLogger

	// Patterns for sensitive data detection
	apiKeyPattern    = regexp.MustCompile(`(?i)(x-api-key|authorization|api[_-]?key)['"]?\s*[:=]\s*['"]?([^'"\s,}]+)`)
	bearerPattern    = regexp.MustCompile(`(?i)Bearer\s+([A-Za-z0-9\-._~+/]+=*)`)
	secretPattern    = regexp.MustCompile(`(?i)(secret|password|token)['"]?\s*[:=]\s*['"]?([^'"\s,}]+)`)
	bodyPattern      = regexp.MustCompile(`(?i)"(messages|prompt|input)":\s*"([^"]{50,})"`)
	largeJSONPattern = regexp.MustCompile(`\{[^}]{200,}\}`)
)

// Init initializes the global logger with file rotation
func Init(home string) error {
	logDir := filepath.Join(home, "logs")
	if err := os.MkdirAll(logDir, 0700); err != nil {
		return fmt.Errorf("create log directory: %w", err)
	}

	// Daily log file with rotation
	logFile := filepath.Join(logDir, fmt.Sprintf("boba-%s.jsonl", time.Now().Format("20060102")))

	// Configure lumberjack for rotation
	writer := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // MB
		MaxBackups: 5,
		MaxAge:     0, // days (0 = no age-based deletion)
		Compress:   false,
		LocalTime:  true,
	}

	// Ensure log file has secure permissions
	if err := os.Chmod(logFile, 0600); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("set log file permissions: %w", err)
	}

	// Configure encoder for JSON output
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		zap.InfoLevel,
	)

	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugarLogger = globalLogger.Sugar()

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if globalLogger != nil {
		//nolint:errcheck,gosec // Best effort sync, error irrelevant at shutdown
		globalLogger.Sync()
	}
}

// Info logs an informational message
func Info(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		return
	}
	globalLogger.Info(Sanitize(msg), sanitizeFields(fields)...)
}

// String creates a zap field for a string value
func String(key, val string) zap.Field {
	return zap.String(key, Sanitize(val))
}

// Int creates a zap field for an int value
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Int64 creates a zap field for an int64 value
func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

// Bool creates a zap field for a bool value
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Error creates a zap field for an error value
func Err(err error) zap.Field {
	if err == nil {
		return zap.Skip()
	}
	return zap.String("error", Sanitize(err.Error()))
}

// Infof logs a formatted informational message
func Infof(format string, args ...interface{}) {
	if sugarLogger == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	sugarLogger.Info(Sanitize(msg))
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		return
	}
	globalLogger.Warn(Sanitize(msg), sanitizeFields(fields)...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	if sugarLogger == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	sugarLogger.Warn(Sanitize(msg))
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		return
	}
	globalLogger.Error(Sanitize(msg), sanitizeFields(fields)...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	if sugarLogger == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	sugarLogger.Error(Sanitize(msg))
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	if globalLogger == nil {
		return
	}
	globalLogger.Debug(Sanitize(msg), sanitizeFields(fields)...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	if sugarLogger == nil {
		return
	}
	msg := fmt.Sprintf(format, args...)
	sugarLogger.Debug(Sanitize(msg))
}

// Sanitize removes sensitive information from log messages
func Sanitize(msg string) string {
	// Mask API keys
	msg = apiKeyPattern.ReplaceAllString(msg, `$1: ***REDACTED***`)

	// Mask Bearer tokens
	msg = bearerPattern.ReplaceAllString(msg, `Bearer ***REDACTED***`)

	// Mask secrets/passwords
	msg = secretPattern.ReplaceAllString(msg, `$1: ***REDACTED***`)

	// Truncate large request/response bodies
	msg = bodyPattern.ReplaceAllString(msg, `"$1": "***TRUNCATED***"`)

	// Truncate large JSON payloads
	msg = largeJSONPattern.ReplaceAllString(msg, `{***TRUNCATED***}`)

	return msg
}

// sanitizeFields sanitizes zap fields
func sanitizeFields(fields []zap.Field) []zap.Field {
	sanitized := make([]zap.Field, len(fields))
	for i, field := range fields {
		// Check field key for sensitive names
		if isSensitiveKey(field.Key) {
			sanitized[i] = zap.String(field.Key, "***REDACTED***")
			continue
		}

		// For string fields, sanitize the value
		if field.Type == zapcore.StringType {
			// Access the string value and sanitize it
			sanitized[i] = zap.String(field.Key, Sanitize(field.String))
		} else {
			sanitized[i] = field
		}
	}
	return sanitized
}

// isSensitiveKey checks if a field key contains sensitive information
func isSensitiveKey(key string) bool {
	sensitive := []string{
		"api_key", "apikey", "api-key",
		"secret", "password", "token",
		"authorization", "auth",
		"x-api-key", "bearer",
		"payload", "request_body", "response_body",
	}

	lowerKey := regexp.MustCompile(`[A-Z]`).ReplaceAllStringFunc(key, func(s string) string {
		return "_" + s
	})
	lowerKey = regexp.MustCompile(`-`).ReplaceAllString(lowerKey, "_")

	for _, s := range sensitive {
		if regexp.MustCompile(`(?i)`+s).MatchString(lowerKey) {
			return true
		}
	}
	return false
}
