package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/logging"
)

var (
	version   = "devel"   // will be injected at build time via ldflags
	buildTime = "unknown" // will be injected at build time via ldflags
	appname   = "aico"
)

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "\nERROR: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	app := &cli.Command{
		Name:    appname,
		Usage:   "AI Assistant Console",
		Version: fmt.Sprintf("%s (built at %s)", version, buildTime),
		// --context is a StringSliceFlag. urfave/cli v3 splits each value on
		// ',' by default, which breaks inline context text that happens to
		// contain a comma (e.g. `--context "$(go doc ./pkg)"`). Disable that
		// splitting so a --context value is taken verbatim.
		//
		// Note: urfave/cli tracks this as process-wide mutable state that is
		// reset by cmd.setupDefaults() every time a (sub)command in the
		// dispatch chain runs, so this must also be set on every subcommand
		// that accepts --context (see CmdSession and its "resume" child).
		DisableSliceFlagSeparator: true,
		EnableShellCompletion:     true,
		Flags: []cli.Flag{
			flagDebug,
			flagJSON,
			flagModel,

			flagSessionID,
			flagLast,
			flagNoStream,
			flagPersona,
			flagSystemPrompt,
			flagSource,
			flagContext,

			flagAPIKeyAnthropic,
			flagAPIKeyOpenAI,
			flagAPIKeyGroq,
			flagAPIKeyCerebras,
		},
		Action:         runGenerate,
		ExitErrHandler: handleExitError,
		Commands: []*cli.Command{
			CmdEnv,
			CmdConfig,
			CmdModels,
			CmdPersona,
			CmdSession,
		},
	}
	return app.Run(context.Background(), args)
}

// Common flags
var (
	flagSource = &cli.StringFlag{
		Name:    "source",
		Aliases: []string{"s"},
		Usage: "the ONE primary subject to act on (what the output is about, " +
			"and may someday be written back to) - a string, @file, @- for " +
			"stdin, or 'label:@file'/'label:@-' to name an unlabeled source " +
			"(e.g. --source @code.go, --source @-, --source 'buffer.go:@-'); " +
			"for read-only reference material instead, see --context",
		// There is exactly one source per prompt by design (see --context
		// for multiple, read-only inputs), so reject a second --source
		// instead of silently letting it win.
		OnlyOnce: true,
	}
	flagContext = &cli.StringSliceFlag{
		Name:    "context",
		Aliases: []string{"c"},
		Usage: "read-only background material for the prompt, never the " +
			"subject acted on - a string, @file, @- for stdin, or " +
			"'label:@file'/'label:@-' to name it; repeatable " +
			"(e.g. --context 'text', --context @file.txt, " +
			"--context \"$(go doc ./pkg)\")",
	}
	flagDebug = &cli.BoolFlag{
		Name:  "debug",
		Usage: "Enable debug logging",
	}
	flagJSON = &cli.BoolFlag{
		Name:  "json",
		Usage: "Output in JSON format",
	}
	flagModel = &cli.StringFlag{
		Name:    "model",
		Aliases: []string{"m"},
		Usage:   "Model to use (e.g., 'gpt-4o' or 'openai:gpt-4o' for explicit provider)",
	}
	flagNoStream = &cli.BoolFlag{
		Name:  "no-stream",
		Usage: "disable streaming output",
	}
	flagPersona = &cli.StringFlag{
		Name:    "persona",
		Aliases: []string{"p"},
		Usage:   "The persona to use",
		Value:   "default",
	}
	flagSystemPrompt = &cli.StringFlag{
		Name:  "system",
		Usage: "system prompt",
	}
	flagSessionID = &cli.StringFlag{
		Name:  "session",
		Usage: "session `ID` for conversation history",
	}
	flagLast = &cli.BoolFlag{
		Name:  "last",
		Usage: "resume the most recent session",
	}

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
	flagAPIKeyGroq = &cli.StringFlag{
		Name:  "groq-api-key",
		Usage: "Groq API Key",
		Sources: cli.NewValueSourceChain(
			cli.EnvVar(envKeyWithPrefix(appname, "groq_api_key")),
		),
	}
	flagAPIKeyCerebras = &cli.StringFlag{
		Name:  "cerebras-api-key",
		Usage: "Cerebras API Key",
		Sources: cli.NewValueSourceChain(
			cli.EnvVar(envKeyWithPrefix(appname, "cerebras_api_key")),
		),
	}
)

// Common errors
var (
	ErrConfigFileNotFound = errors.New("config file not found")
)

func handleExitError(ctx context.Context, cmd *cli.Command, err error) {
	logging.LoggerFrom(ctx).
		With("cmd", cmd.Name).
		Error("exiting with error", "error", err)
}
