package commands

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
)

// loadConfig is a helper function to load the configuration and attach it to the context.
func loadConfig(c *cli.Context) error {
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
