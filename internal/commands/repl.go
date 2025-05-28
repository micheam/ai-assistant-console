package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

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

		// Load GenerativeModel
		m, err := setupGenerativeModel(conf.Chat.Model)
		if err != nil {
			return fmt.Errorf("create model: %w", err)
		}

		ctx = logging.ContextWith(ctx, logger.With("model", m.Name()))
		persona := cmd.String("persona")
		return repl.Init(conf, persona, m).Run(ctx)
	},
}
