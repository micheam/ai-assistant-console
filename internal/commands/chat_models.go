package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/anthropic"
	"micheam.com/aico/internal/logging"
	"micheam.com/aico/internal/openai"
)

var ChatModels = &cli.Command{
	Name:  "models",
	Usage: "Show the available models",
	Flags: []cli.Flag{
		flagDebug,
		flagJSON,
		flagModel,
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		conf, err := LoadConfig(ctx, cmd)
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}

		// Setup logger
		logLevel := logging.LevelInfo
		if cmd.Bool("debug") {
			logLevel = logging.LevelDebug
		}
		logger, cleanup, err := setupLogger(conf.Logfile(), logLevel)
		if err != nil {
			return err
		}
		defer cleanup()

		models := []modelView{}

		// OpenAI Models
		for _, model := range openai.AvailableModels() {
			desc, _ := openai.DescribeModel(model)
			models = append(models, modelView{
				Name:        model,
				Selected:    model == conf.Chat.Model,
				Description: strings.ReplaceAll(desc, "\n", " "),
			})
		}

		// Anthropic Models
		for _, model := range anthropic.AvailableModels() {
			desc, _ := anthropic.DescribeModel(model)
			models = append(models, modelView{
				Name:        model,
				Selected:    model == conf.Chat.Model,
				Description: strings.ReplaceAll(desc, "\n", " "),
			})
		}

		// Print the models
		for _, model := range models {
			if cmd.Bool("json") {
				if err := json.NewEncoder(cmd.Root().Writer).Encode(model); err != nil {
					return err
				}
				continue
			}
			fmt.Fprintln(cmd.Root().Writer, model.String())
		}

		_ = logger
		return nil
	},
}

type modelView struct {
	Selected    bool   `json:"selected"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (m *modelView) String() string {
	if m.Selected {
		return fmt.Sprintf("%s *", m.Name)
	}
	return m.Name
}
