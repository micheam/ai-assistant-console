package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
)

// LoadConfig is a helper function to load the configuration and attach it to the context.
//
// errors:
//
// - [ErrConfigFileNotFound]: The config file was not found.
func LoadConfig(_ context.Context, cmd *cli.Command) (*config.Config, error) {
	conf, err := config.Load()
	if errors.Is(err, config.ErrConfigFileNotFound) {
		return nil, ErrConfigFileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	// Overwrite Model with command-line flag (`-m, --model`)
	if model := cmd.String("model"); model != "" {
		conf.Chat.Model = model
	}
	return conf, nil
}

// readLines reads lines from the given reader.
func readLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan lines: %w", err)
	}
	return lines, nil
}

// setupLogger initializes and returns a logger based on configuration and log level.
func setupLogger(filename string, level slog.Level) (*logging.Logger, func(), error) {
	cleanup := func() {}
	if filename == "" {
		return nil, cleanup, fmt.Errorf("empty filename")
	}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("open logfile: %w", err)
	}
	opt := &logging.Options{Level: level}
	cleanup = func() { f.Close() }
	return logging.New(f, opt), cleanup, nil
}
