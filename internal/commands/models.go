package commands

import (
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/openai"
)

var ChatModels = &cli.Command{
	Name:  "models",
	Usage: "Show the available models",
	Flags: []cli.Flag{
		flagJSON,
		flagModel,
	},
	Before: loadConfig,
	Action: func(c *cli.Context) error {
		conf := config.MustFromContext(c.Context)
		models := openai.AvailableModels()
		for _, model := range models {
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
