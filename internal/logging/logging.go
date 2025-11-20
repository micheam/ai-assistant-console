package logging

import (
	"context"
	"io"
	"log/slog"
)

type Options struct {
	AddSource bool
	Level     Level
	// ReplaceAttr func(groups []string, a Attr) Attr
}

type Level = slog.Level

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Logger = slog.Logger

func New(out io.Writer, opts *Options) *Logger {
	if opts == nil {
		opts = &Options{}
	}
	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		AddSource: opts.AddSource,
		Level:     opts.Level,
	})
	return slog.New(handler)
}

// ----------------------------------------------------------------------------
// Mng Logger with context
// ----------------------------------------------------------------------------

type contextKey int

const contextKeyLogger contextKey = iota

// ContextWith sets the logger in the context.
func ContextWith(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

// LoggerFrom retrieves the logger from the context.
func LoggerFrom(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(contextKeyLogger).(*Logger); ok {
		return logger
	}
	return New(io.Discard, nil)
}
