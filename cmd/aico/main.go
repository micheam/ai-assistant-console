package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	app := &cli.Command{
		Name:  "aico",
		Usage: "AI Assistant Console",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "model",
				Aliases: []string{"m"},
				Usage:   "model name to use",
				Sources: cli.EnvVars("AI_MODEL"),
			},
			&cli.StringFlag{
				Name:    "context-file",
				Aliases: []string{"c"},
				Usage:   "context file path",
			},
			&cli.StringFlag{
				Name:  "system",
				Usage: "system prompt",
			},
			&cli.BoolFlag{
				Name:  "no-stream",
				Usage: "disable streaming output",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args()
			if args.Len() == 0 {
				return cli.ShowAppHelp(cmd)
			}
			return runGenerate(ctx, cmd, args.Slice())
		},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "show version information",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					fmt.Println("aico version 0.1.0")
					return nil
				},
			},
			CmdEnv,
			CmdConfig,
			CmdModels,
		},
	}
	return app.Run(context.Background(), args)
}

// Common flags
var (
	flagDebug = &cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug logging",
	}
	flagModel = &cli.StringFlag{
		Name:    "model",
		Aliases: []string{"m"},
		Usage:   "The model to use",
	}
	flagJSON = &cli.BoolFlag{
		Name:  "json",
		Usage: "Output the models in JSON format",
	}
	flagChatSession = &cli.StringFlag{
		Name:    "session",
		Aliases: []string{"s"},
		Usage:   "The chat session ID",
	}
	flagChatInstant = &cli.BoolFlag{
		Name:  "instant",
		Usage: "Instantly send the message without storing it in the chat session",
	}
	flagPersona = &cli.StringFlag{
		Name:    "persona",
		Aliases: []string{"p"},
		Usage:   "The persona to use",
		Value:   "default",
	}
)

// Common errors
var (
	ErrConfigFileNotFound = errors.New("config file not found")
)
