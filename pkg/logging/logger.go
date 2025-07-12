package logging

import (
	"log/slog"
	"os"
	"strings"
)

// Component represents different parts of the system for scoped logging
type Component string

const (
	ComponentCLI       Component = "cli"
	ComponentServer    Component = "server"
	ComponentGRPC      Component = "grpc"
	ComponentHTTP      Component = "http"
	ComponentDatastore Component = "datastore"
)

// Config holds logging configuration
type Config struct {
	Level  slog.Level
	Format string // "json" or "text"
}

// DefaultConfig returns a default logging configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  slog.LevelInfo,
		Format: "text",
	}
}

// NewLogger creates a new slog.Logger with the given configuration
func NewLogger(config *Config) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: config.Level,
	}

	switch strings.ToLower(config.Format) {
	case "json":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// For gets a component-scoped logger from the default logger
func For(component Component) *slog.Logger {
	return slog.Default().With("component", string(component))
}

// ForComponent gets a component-scoped logger with a request ID
func ForComponent(component Component, requestID string) *slog.Logger {
	return slog.Default().With("component", string(component), "request_id", requestID)
}

// ParseLevel parses a string log level into slog.Level
func ParseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
