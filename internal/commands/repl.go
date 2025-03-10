package commands

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
	"micheam.com/aico/internal/repl"
)

var ChatRepl = &cli.Command{
	Name:  "repl",
	Usage: "Start a chat session in a REPL",
	Flags: []cli.Flag{
		flagDebug,
		flagModel,
		flagPersona,
	},
	Before: LoadConfig,
	Action: func(c *cli.Context) error {
		// Load configuration
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

		// Load GenerativeModel
		m, err := setupGenerativeModel(conf.Chat.Model)
		if err != nil {
			return fmt.Errorf("create model: %w", err)
		}

		ctx := logging.ContextWith(c.Context, logger.With("model", m.Name()))
		return repl.Init(conf, c.String("persona"), m).Run(ctx)
	},
}
