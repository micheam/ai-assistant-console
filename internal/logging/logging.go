package logging

import (
	"context"
	"io"
	"log"
	"os"
)

type contextKey int

const contextKeyLogger contextKey = iota

// NewLogger initializes a new logger with the specified output and flags.
func NewLogger(output io.Writer, prefix string, flags int) *log.Logger {
	return log.New(output, prefix, flags)
}

// SetupLogger sets up the logger based on the debug flag and configuration.
func SetupLogger(debug bool, logFile string) (*log.Logger, error) {
	logger := NewLogger(io.Discard, "", log.LstdFlags|log.LUTC)
	if debug {
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			if os.IsNotExist(err) {
				if f, err = os.Create(logFile); err != nil {
					return nil, err
				}
			} else {
				return nil, err
			}
		}
		logger.SetOutput(f)
	}
	return logger, nil
}

// WithLogger sets the logger in the context.
func WithLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

// LoggerFrom retrieves the logger from the context.
func LoggerFrom(ctx context.Context) *log.Logger {
	return ctx.Value(contextKeyLogger).(*log.Logger)
}
