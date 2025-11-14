// Package logging provides structured logging with rotation, sampling, and sanitization.
package logging

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Field aliases zap.Field so callers don't depend on zap directly.
type Field = zap.Field

// Logger exposes structured logging operations with contextual fields.
type Logger interface {
	With(fields ...Field) Logger
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Sync() error
}

// Config defines how a logger should write and sample structured events.
type Config struct {
	// Home is the root configuration directory (e.g. ~/.boba).
	Home string
	// Path overrides the log file path. If empty, Home/logs/boba-YYYYMMDD.jsonl is used.
	Path string
	// Level controls the minimum severity (debug, info, warn, error).
	Level string
	// MaxSizeMB controls log rotation size. Defaults to 10MB.
	MaxSizeMB int
	// MaxBackups configures how many rotated files to keep. Defaults to 5.
	MaxBackups int
	// MaxAgeDays controls how many days to keep logs (0 = unlimited).
	MaxAgeDays int
	// SamplingInitial configures zap sampling initial count. Defaults to 100.
	SamplingInitial int
	// SamplingThereafter configures zap sampling thereafter count. Defaults to 100.
	SamplingThereafter int
}

var (
	defaultLogger Logger = &noopLogger{}

	apiKeyPattern    = regexp.MustCompile(`(?i)(x-api-key|authorization|api[_-]?key)['"]?\s*[:=]\s*['"]?([^'"\s,}]+)`)
	bearerPattern    = regexp.MustCompile(`(?i)Bearer\s+([A-Za-z0-9\-._~+/]+=*)`)
	secretPattern    = regexp.MustCompile(`(?i)(secret|password|token)['"]?\s*[:=]\s*['"]?([^'"\s,}]+)`)
	bodyPattern      = regexp.MustCompile(`(?i)"(messages|prompt|input)":\s*"([^"]{50,})"`)
	largeJSONPattern = regexp.MustCompile(`\{[^}]{200,}\}`)
	upperPattern     = regexp.MustCompile(`[A-Z]`)
	dashPattern      = regexp.MustCompile(`-`)
)

// New builds a structured logger based on cfg.
func New(cfg Config) (Logger, error) {
	path, err := resolvePath(cfg)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	writer := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    pickInt(cfg.MaxSizeMB, 10),
		MaxBackups: pickInt(cfg.MaxBackups, 5),
		MaxAge:     pickInt(cfg.MaxAgeDays, 0),
		LocalTime:  true,
	}
	if err := os.Chmod(path, 0o600); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("set log file permissions: %w", err)
	}

	level := zap.InfoLevel
	if err := level.UnmarshalText([]byte(strings.ToLower(strings.TrimSpace(cfg.Level)))); err != nil && cfg.Level != "" {
		return nil, fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), zapcore.AddSync(writer), level)
	if cfg.SamplingInitial != 0 || cfg.SamplingThereafter != 0 {
		core = zapcore.NewSamplerWithOptions(core, time.Second, pickInt(cfg.SamplingInitial, 100), pickInt(cfg.SamplingThereafter, 100))
	}

	zl := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &zapLogger{logger: zl}, nil
}

// Configure replaces the package-level default logger.
func Configure(cfg Config) (Logger, error) {
	log, err := New(cfg)
	if err != nil {
		return nil, err
	}
	defaultLogger = log
	return log, nil
}

// Init configures the default logger using the provided home directory.
func Init(home string) error {
	_, err := Configure(Config{Home: home})
	return err
}

// Sync flushes buffered logs on the default logger.
func Sync() error {
	return defaultLogger.Sync()
}

// With adds structured fields to the default logger.
func With(fields ...Field) Logger {
	return defaultLogger.With(fields...)
}

// Info logs an informational message using the default logger.
func Info(msg string, fields ...Field) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs a warning using the default logger.
func Warn(msg string, fields ...Field) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs an error using the default logger.
func Error(msg string, fields ...Field) {
	defaultLogger.Error(msg, fields...)
}

type zapLogger struct {
	logger *zap.Logger
}

func (z *zapLogger) With(fields ...Field) Logger {
	if len(fields) == 0 {
		return z
	}
	return &zapLogger{logger: z.logger.With(sanitizeFields(fields)...)}
}

func (z *zapLogger) Info(msg string, fields ...Field) {
	z.logger.Info(Sanitize(msg), sanitizeFields(fields)...)
}

func (z *zapLogger) Warn(msg string, fields ...Field) {
	z.logger.Warn(Sanitize(msg), sanitizeFields(fields)...)
}

func (z *zapLogger) Error(msg string, fields ...Field) {
	z.logger.Error(Sanitize(msg), sanitizeFields(fields)...)
}

func (z *zapLogger) Sync() error {
	if z.logger == nil {
		return nil
	}
	return z.logger.Sync()
}

type noopLogger struct{}

func (n *noopLogger) With(_ ...Field) Logger { return n }
func (n *noopLogger) Info(string, ...Field)  {}
func (n *noopLogger) Warn(string, ...Field)  {}
func (n *noopLogger) Error(string, ...Field) {}
func (n *noopLogger) Sync() error            { return nil }

// String creates a string field with sanitization.
func String(key, val string) Field {
	return zap.String(key, Sanitize(val))
}

// Int creates an integer field.
func Int(key string, val int) Field {
	return zap.Int(key, val)
}

// Int64 creates an int64 field.
func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

// Bool creates a bool field.
func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

// Err serializes an error value while redacting sensitive details.
func Err(err error) Field {
	if err == nil {
		return zap.Skip()
	}
	return zap.String("error", Sanitize(err.Error()))
}

// Sanitize removes sensitive markers from a message before logging.
func Sanitize(msg string) string {
	msg = apiKeyPattern.ReplaceAllString(msg, `$1: ***REDACTED***`)
	msg = bearerPattern.ReplaceAllString(msg, `Bearer ***REDACTED***`)
	msg = secretPattern.ReplaceAllString(msg, `$1: ***REDACTED***`)
	msg = bodyPattern.ReplaceAllString(msg, `"$1": "***TRUNCATED***"`)
	msg = largeJSONPattern.ReplaceAllString(msg, `{***TRUNCATED***}`)
	return msg
}

func sanitizeFields(fields []Field) []Field {
	sanitized := make([]Field, len(fields))
	for i, field := range fields {
		if isSensitiveKey(field.Key) {
			sanitized[i] = zap.String(field.Key, "***REDACTED***")
			continue
		}
		if field.Type == zapcore.StringType {
			sanitized[i] = zap.String(field.Key, Sanitize(field.String))
		} else {
			sanitized[i] = field
		}
	}
	return sanitized
}

func isSensitiveKey(key string) bool {
	sensitive := []string{"api_key", "apikey", "api-key", "secret", "password", "token", "authorization", "auth", "x-api-key", "bearer", "payload", "request_body", "response_body"}
	normalized := upperPattern.ReplaceAllStringFunc(key, func(s string) string {
		return "_" + strings.ToLower(s)
	})
	normalized = strings.ToLower(dashPattern.ReplaceAllString(normalized, "_"))
	for _, pattern := range sensitive {
		needle := strings.ToLower(pattern)
		if needle == "token" && strings.Contains(normalized, "tokens") {
			continue
		}
		if strings.Contains(normalized, needle) {
			return true
		}
	}
	return false
}

func resolvePath(cfg Config) (string, error) {
	if cfg.Path != "" {
		return cfg.Path, nil
	}
	home := cfg.Home
	if home == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		home = filepath.Join(dir, ".boba")
	}
	filename := fmt.Sprintf("boba-%s.jsonl", time.Now().Format("20060102"))
	return filepath.Join(home, "logs", filename), nil
}

func pickInt(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
