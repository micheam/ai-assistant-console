package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/repl"
)

var ChatRepl = &cli.Command{
	Name:  "repl",
	Usage: "Start a chat session in a REPL",
	Flags: []cli.Flag{
		flagModel,
		flagPersona,
	},
	Before: LoadConfig,
	Action: func(c *cli.Context) error {
		// Load configuration
		conf := config.MustFromContext(c.Context)

		// Load GenerativeModel
		m, err := setupGenerativeModel(*conf)
		if err != nil {
			return fmt.Errorf("create model: %w", err)
		}

		return repl.Init(conf, c.String("persona"), m).Run(c.Context)
	},
}
