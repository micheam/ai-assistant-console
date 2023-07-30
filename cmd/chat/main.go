package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/theme"
	tui "micheam.com/aico/internal/tui/chat"
)

const authEnvKey = "OPENAI_API_KEY"

//go:embed version.txt
var version string

func main() {
	if err := app().Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func app() *cli.App {
	return &cli.App{
		Name:    "chat",
		Usage:   "Chat with AI",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug mode",
				EnvVars: []string{"AICO_DEBUG"},
			},
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Usage:   "GPT model to use",
				Value:   config.DefaultModel,
			},
			&cli.BoolFlag{
				Name:  "tui",
				Usage: "Enable TUI mode",
			},
		},
		Action: func(c *cli.Context) error {
			ctx := c.Context

			// Load Config
			cfg, err := config.InitAndLoad(ctx)
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			// Setup logger
			logger := log.New(io.Discard, "", log.LstdFlags|log.LUTC)
			if c.Bool("debug") {
				lfile := cfg.Logfile()
				f, err := os.OpenFile(
					lfile,
					os.O_APPEND|os.O_CREATE|os.O_WRONLY,
					0644,
				)
				if os.IsNotExist(err) {
					if f, err = os.Create(lfile); err != nil {
						return fmt.Errorf("prepare log file(%q): %w", lfile, err)
					}
				}
				defer f.Close()

				logger.SetOutput(f)
				logger.SetPrefix("[CHAT] ")

				fmt.Println(theme.Info("Debug mode is on")) // TODO: promote to Logger or Presenter
				fmt.Printf(theme.Info("You can find logs in %q\n"), lfile)
				fmt.Println()
			}

			// Override model if specified with flag
			model := c.String("model")
			cfg.Chat.Model = model

			if c.Bool("tui") {
				logger.SetPrefix("[CHAT][TUI] ")
				handler := tui.New(cfg, logger)
				return handler.Run(ctx)
			}
			return fmt.Errorf("not implemented")
		},
	}
}
