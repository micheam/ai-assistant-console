package commands

import (
	"encoding/json"
	"fmt"

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

		for _, model := range availableModels() {
			mv := modelView{
				Name:     model,
				Selected: model == conf.Chat.Model,
			}
			if c.Bool("json") {
				if err := json.NewEncoder(c.App.Writer).Encode(mv); err != nil {
					return err
				}
				continue
			}
			fmt.Fprintln(c.App.Writer, mv.String())
		}

		_ = logger
		return nil
	},
}

type modelView struct {
	Name     string `json:"name"`
	Selected bool   `json:"selected"`
}

func (m *modelView) String() string {
	if m.Selected {
		return fmt.Sprintf("%s *", m.Name)
	}
	return m.Name
}

func availableModels() []string {
	models := append([]string{}, openai.AvailableModels()...)
	models = append(models, anthropic.AvailableModels()...)
	return models
}
