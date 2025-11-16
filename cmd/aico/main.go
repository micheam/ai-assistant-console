package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

var version = "devel"

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	app := &cli.Command{
		Name:                  "aico",
		Usage:                 "AI Assistant Console",
		Version:               version,
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			flagContextFile,
			flagDebug,
			flagJSON,
			flagModel,
			flagNoStream,
			flagPersona,
			flagSystemPrompt,
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args()
			if args.Len() == 0 {
				return cli.ShowAppHelp(cmd)
			}
			return runGenerate(ctx, cmd, args.Slice())
		},
		Commands: []*cli.Command{
			CmdEnv,
			CmdConfig,
			CmdModels,
		},
	}
	return app.Run(context.Background(), args)
}

// Common flags
var (
	flagContextFile = &cli.StringFlag{
		Name:    "context-file",
		Aliases: []string{"c"},
		Usage:   "context file path"}

	flagDebug = &cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug logging"}

	flagJSON = &cli.BoolFlag{
		Name:  "json",
		Usage: "Output the models in JSON format"}

	flagModel = &cli.StringFlag{
		Name:    "model",
		Aliases: []string{"m"},
		Usage:   "The model to use"}

	flagNoStream = &cli.BoolFlag{
		Name:  "no-stream",
		Usage: "disable streaming output"}

	flagPersona = &cli.StringFlag{
		Name:    "persona",
		Aliases: []string{"p"},
		Usage:   "The persona to use",
		Value:   "default"}

	flagSystemPrompt = &cli.StringFlag{
		Name:  "system",
		Usage: "system prompt"}
)

// Common errors
var (
	ErrConfigFileNotFound = errors.New("config file not found")
)
