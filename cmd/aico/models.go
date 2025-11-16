package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/providers/anthropic"
	"micheam.com/aico/internal/providers/openai"
)

var CmdModels = &cli.Command{
	Name:  "models",
	Usage: "manage AI models",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "debug",
			Usage: "Enable debug logging",
		},
		&cli.BoolFlag{
			Name:  "json",
			Usage: "output in JSON format",
		},
	},

	// default action: list models
	Action: runListModels,
	Commands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "list available models",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "json",
					Usage: "output in JSON format",
				},
			},
			Action: runListModels,
		},
		{
			Name:      "show",
			Usage:     "show model information",
			ArgsUsage: "MODEL",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "json",
					Usage: "output in JSON format",
				},
			},
			Action: runShowModelInfo,
		},
	},
}

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runListModels(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	models := []listItemView{}
	for _, model := range allAvailableModels() {
		models = append(models, listItemView{
			Name:        model.Name(),
			Provider:    model.Provider(),
			Description: model.Description(),
			Selected:    model.Name() == conf.Chat.Model,
		})
	}

	if cmd.Bool("json") {
		encoder := json.NewEncoder(cmd.Root().Writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(models)
	}

	for _, model := range models {
		fmt.Fprintln(cmd.Root().Writer, model.String())
	}
	return nil
}

func allAvailableModels() []assistant.ModelDescriptor {
	return append(anthropic.AvailableModels(), openai.AvailableModels()...)
}

type listItemView struct {
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
}

func (m *listItemView) String() string {
	if m.Selected {
		return fmt.Sprintf("%s *", m.Name)
	}
	return m.Name
}

func runShowModelInfo(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 1 {
		return fmt.Errorf("model name is required")
	}
	modelName := cmd.Args().Get(0)
	for _, model := range allAvailableModels() {
		if model.Name() == modelName {
			if cmd.Bool("json") {
				info := map[string]string{
					"name":        model.Name(),
					"provider":    model.Provider(),
					"description": model.Description(),
				}
				encoder := json.NewEncoder(cmd.Root().Writer)
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			}

			fmt.Fprintf(cmd.Root().Writer, "Model: %s\n", model.Name())
			fmt.Fprintf(cmd.Root().Writer, "Provider: %s\n", model.Provider())
			fmt.Fprintf(cmd.Root().Writer, "Description: %s\n", model.Description())
			return nil
		}
	}
	return fmt.Errorf("model not found: %s", modelName)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

type model struct {
	Name        string
	provider    string
	Description string
}
