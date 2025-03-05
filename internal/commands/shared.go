package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
)

var flagModel = &cli.StringFlag{
	Name:    "model",
	Aliases: []string{"m"},
	Usage:   "The model to use",
}

var flagJSON = &cli.BoolFlag{
	Name:  "json",
	Usage: "Output the models in JSON format",
}

var flagChatSession = &cli.StringFlag{
	Name:    "session",
	Aliases: []string{"s"},
	Usage:   "The chat session ID",
}

var flagChatInstant = &cli.BoolFlag{
	Name:  "instant",
	Usage: "Instantly send the message without storing it in the chat session",
}

var flagPersona = &cli.StringFlag{
	Name:    "persona",
	Aliases: []string{"p"},
	Usage:   "The persona to use",
	Value:   "default",
}

// LoadConfig is a helper function to load the configuration and attach it to the context.
func LoadConfig(c *cli.Context) error {
	conf, err := config.Load()
	if errors.Is(err, config.ErrConfigFileNotFound) {
		fmt.Fprintln(c.App.Writer, "Please run 'chat config init' to init the configuration file.")
	}
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Overwrite Model with command-line flag (`-m, --model`)
	if model := c.String("model"); model != "" {
		conf.Chat.Model = model
	}

	c.Context = config.WithConfig(c.Context, conf)
	return nil
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
