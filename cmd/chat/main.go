package main

import (
	"context"
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
		},
		Before: func(c *cli.Context) error {
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
				fmt.Println(theme.Info("Debug mode is on")) // TODO: promote to Logger or Presenter
				fmt.Printf(theme.Info("You can find logs in %q\n"), lfile)
				fmt.Println()
			}

			// Override model if specified with flag
			model := c.String("model")
			cfg.Chat.Model = model

			// Setup context
			ctx = WithConfig(ctx, cfg)
			ctx = WithLogger(ctx, logger)

			c.Context = ctx
			return nil
		},
		Commands: []*cli.Command{

			{
				Name:        "config",
				Usage:       "Show config file path",
				Description: "Show config file path",
				Action: func(c *cli.Context) error {
					path := config.ConfigFilePath(c.Context)
					fmt.Println(path)
					return nil
				},
			},

			{
				Name:        "tui",
				Usage:       "Chat with AI in TUI",
				Description: "Start TUI application to chat with AI",
				Action: func(c *cli.Context) error {
					ctx := c.Context
					cfg := ConfigFrom(ctx)
					logger := LoggerFrom(ctx)
					logger.SetPrefix("[CHAT][TUI] ")

					handler := tui.New(cfg, logger)
					return handler.Run(ctx)
				},
			},

			SendMessageCommand,
		},
	}
}

// --------------------------------------------------------------------
// Handle context
// --------------------------------------------------------------------

type contextKey int

const (
	contextKeyConfig contextKey = iota
	contextKeyLogger
)

func WithConfig(ctx context.Context, cfg *config.Config) context.Context {
	return context.WithValue(ctx, contextKeyConfig, cfg)
}

func ConfigFrom(ctx context.Context) *config.Config {
	return ctx.Value(contextKeyConfig).(*config.Config)
}

func WithLogger(ctx context.Context, logger *log.Logger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}

func LoggerFrom(ctx context.Context) *log.Logger {
	return ctx.Value(contextKeyLogger).(*log.Logger)
}
