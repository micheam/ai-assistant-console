package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"
)

const (
	version = "devel"
	appname = "aico"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	app := &cli.Command{
		Name:                  appname,
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

			flagAPIKeyAnthropic,
			flagAPIKeyOpenAI,
		},
		Action: runGenerate,
		Commands: []*cli.Command{
			CmdEnv,
			CmdConfig,
			CmdModels,
			CmdPersona,
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

	//
	// API Keys For Providers
	//

	flagAPIKeyOpenAI = &cli.StringFlag{
		Name:  "openai-api-key",
		Usage: "OpenAI API Key",
		Sources: cli.NewValueSourceChain(
			cli.EnvVar(envKeyWithPrefix(appname, "openai_api_key")),
		),
	}
	flagAPIKeyAnthropic = &cli.StringFlag{
		Name:  "anthropic-api-key",
		Usage: "Anthropic API Key",
		Sources: cli.NewValueSourceChain(
			cli.EnvVar(envKeyWithPrefix(appname, "anthropic_api_key")),
		),
	}
)

// Common errors
var (
	ErrConfigFileNotFound = errors.New("config file not found")
)
