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

// datadir returns default data directory
//
// We determin data directory by the rules below:
// 1. If AICO_DATA_DIR environment variable is set, use it
// 2. If XDG_DATA_HOME environment variable is set, use it
// 3. otherwise, use $HOME/.local/share
//
// TODO: use internal/config instead
func datadir() string {
	if os.Getenv("AICO_DATA_DIR") != "" {
		return os.Getenv("AICO_DATA_DIR")
	}

	if os.Getenv("XDG_DATA_HOME") != "" {
		return os.Getenv("XDG_DATA_HOME")
	}

	return fmt.Sprintf("%s/.local/share", os.Getenv("HOME"))
}

// logfile returns logfile with location based on datadir.
//
// TODO: make log file configurable
// TODO: use internal/config to detect log file location
func logfile() *os.File {
	logfile, err := os.OpenFile(
		fmt.Sprintf("%s/chatgpt.log", datadir()),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
	return logfile
}
