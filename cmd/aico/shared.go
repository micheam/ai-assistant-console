package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
)

// loadConfig is a helper function to load the configuration and attach it to the context.
//
// errors:
//
// - [ErrConfigFileNotFound]: The config file was not found.
func loadConfig(ctx context.Context, cmd *cli.Command) (*config.Config, error) {
	conf, err := config.Load()
	if errors.Is(err, config.ErrConfigFileNotFound) {
		return nil, ErrConfigFileNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if model := cmd.String(flagModel.Name); model != "" {
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

// initializeLogger initializes a logger from command-line flags and returns it with a cleanup function.
func initializeLogger(ctx context.Context, cmd *cli.Command) (*logging.Logger, func(), error) {
	conf, err := loadConfig(ctx, cmd)
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}
	logLevel := logging.LevelInfo
	if cmd.Bool(flagDebug.Name) {
		logLevel = logging.LevelDebug
	}
	f, err := conf.OpenLogfile()
	if err != nil {
		return nil, nil, fmt.Errorf("open logfile: %w", err)
	}
	return logging.New(f, &logging.Options{Level: logLevel}),
		func() { f.Close() },
		nil
}
