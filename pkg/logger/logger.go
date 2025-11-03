package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

// Logger wraps slog for structured logging
type Logger struct {
	*slog.Logger
}

// Config holds logger configuration
type Config struct {
	Level      string
	Format     string // json or text
	AddSource  bool
	Service    string
	Version    string
	Environment string
}

// contextKey is the type for context keys
type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

// New creates a new structured logger
func New(cfg Config) *Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Wrap with default attributes
	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", cfg.Service),
		slog.String("version", cfg.Version),
		slog.String("environment", cfg.Environment),
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewDefault creates a default logger
func NewDefault() *Logger {
	return New(Config{
		Level:       "info",
		Format:      "json",
		AddSource:   true,
		Service:     "aureo-vpn",
		Version:     "1.0.0",
		Environment: getEnv("ENVIRONMENT", "development"),
	})
}

// WithRequestID adds a request ID to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{
		Logger: l.With(slog.String("request_id", requestID)),
	}
}

// WithUserID adds a user ID to the logger
func (l *Logger) WithUserID(userID uuid.UUID) *Logger {
	return &Logger{
		Logger: l.With(slog.String("user_id", userID.String())),
	}
}

// WithContext extracts values from context and adds them to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	attrs := []slog.Attr{}

	if requestID := ctx.Value(requestIDKey); requestID != nil {
		if id, ok := requestID.(string); ok {
			attrs = append(attrs, slog.String("request_id", id))
		}
	}

	if userID := ctx.Value(userIDKey); userID != nil {
		if id, ok := userID.(uuid.UUID); ok {
			attrs = append(attrs, slog.String("user_id", id.String()))
		}
	}

	if len(attrs) > 0 {
		return &Logger{
			Logger: l.Logger.With(slog.Group("context", attrs)),
		}
	}

	return l
}

// WithError adds an error to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.With(slog.String("error", err.Error())),
	}
}

// WithField adds a custom field to the logger
func (l *Logger) WithField(key string, value any) *Logger {
	return &Logger{
		Logger: l.With(slog.Any(key, value)),
	}
}

// WithFields adds multiple custom fields to the logger
func (l *Logger) WithFields(fields map[string]any) *Logger {
	attrs := make([]any, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}
	return &Logger{
		Logger: l.With(attrs...),
	}
}

// LogRequest logs an HTTP request
func (l *Logger) LogRequest(method, path, ip string, statusCode int, duration time.Duration) {
	l.Info("http_request",
		slog.String("method", method),
		slog.String("path", path),
		slog.String("ip", ip),
		slog.Int("status", statusCode),
		slog.Duration("duration", duration),
	)
}

// LogError logs an error with stack trace
func (l *Logger) LogError(msg string, err error, fields ...any) {
	attrs := append([]any{slog.String("error", err.Error())}, fields...)
	l.Error(msg, attrs...)
}

// LogPanic logs a panic and recovers
func (l *Logger) LogPanic(r any) {
	l.Error("panic_recovered",
		slog.Any("panic", r),
	)
}

// LogDBQuery logs a database query
func (l *Logger) LogDBQuery(query string, duration time.Duration, err error) {
	attrs := []any{
		slog.String("query", query),
		slog.Duration("duration", duration),
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		l.Error("database_query_failed", attrs...)
	} else {
		l.Debug("database_query", attrs...)
	}
}

// LogAuth logs an authentication event
func (l *Logger) LogAuth(event, userID, ip string, success bool) {
	l.Info("auth_event",
		slog.String("event", event),
		slog.String("user_id", userID),
		slog.String("ip", ip),
		slog.Bool("success", success),
	)
}

// LogVPN logs a VPN event
func (l *Logger) LogVPN(event string, sessionID, userID, nodeID uuid.UUID, protocol string) {
	l.Info("vpn_event",
		slog.String("event", event),
		slog.String("session_id", sessionID.String()),
		slog.String("user_id", userID.String()),
		slog.String("node_id", nodeID.String()),
		slog.String("protocol", protocol),
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Global logger instance
var global *Logger

func init() {
	global = NewDefault()
}

// Global returns the global logger instance
func Global() *Logger {
	return global
}

// SetGlobal sets the global logger instance
func SetGlobal(l *Logger) {
	global = l
}

// Helper functions for global logger
func Debug(msg string, args ...any) {
	global.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	global.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	global.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	global.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	global.Error(msg, args...)
	os.Exit(1)
}
