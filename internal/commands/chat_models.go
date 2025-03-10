package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/anthropic"
	"micheam.com/aico/internal/config"
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
	Before: LoadConfig,
	Action: func(c *cli.Context) error {
		conf := config.MustFromContext(c.Context)

		// Setup logger
		logLevel := logging.LevelInfo
		if c.Bool("debug") {
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
			if c.Bool("json") {
				if err := json.NewEncoder(c.App.Writer).Encode(model); err != nil {
					return err
				}
				continue
			}
			fmt.Fprintln(c.App.Writer, model.String())
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
