package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

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
	},

	// default action: list models
	Action: listModels,
	Commands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "list available models",
			Action:  listModels,
		},
		{
			Name:      "info",
			Usage:     "show model information",
			ArgsUsage: "MODEL",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if cmd.Args().Len() == 0 {
					return fmt.Errorf("model name required")
				}
				return showModelInfo(ctx, cmd.Args().First())
			},
		},
		{
			Name:      "test",
			Usage:     "test model connection",
			ArgsUsage: "MODEL",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				if cmd.Args().Len() == 0 {
					return fmt.Errorf("model name required")
				}
				return testModel(ctx, cmd.Args().First())
			},
		},
	},
}

func listModels(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	models := []listItemView{}
	for _, model := range openai.AvailableModels() { // OpenAI Models
		desc, _ := openai.DescribeModel(model)
		models = append(models, listItemView{
			Name:        model,
			Selected:    model == conf.Chat.Model,
			Description: strings.ReplaceAll(desc, "\n", " "),
		})
	}
	for _, model := range anthropic.AvailableModels() { // Anthropic Models
		desc, _ := anthropic.DescribeModel(model)
		models = append(models, listItemView{
			Name:        model,
			Selected:    model == conf.Chat.Model,
			Description: strings.ReplaceAll(desc, "\n", " "),
		})
	}
	for _, model := range models {
		fmt.Fprintln(cmd.Root().Writer, model.String())
	}
	return nil
}

type listItemView struct {
	Selected    bool   `json:"selected"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (m *listItemView) String() string {
	if m.Selected {
		return fmt.Sprintf("%s *", m.Name)
	}
	return m.Name
}

// Show Model Info Command -------------------------------------------------------------------------

func showModelInfo(ctx context.Context, model string) error {
	fmt.Printf("Model: %s\n", model)
	fmt.Println("Provider: Anthropic")
	fmt.Println("Context window: 200k tokens")
	return nil
}

func testModel(ctx context.Context, model string) error {
	fmt.Printf("Testing connection to %s...\n", model)
	// TODO: 実装
	return nil
}
