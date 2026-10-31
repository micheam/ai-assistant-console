//go:build unix

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/complete"
	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
	"micheam.com/aico/internal/providers/llamacpp"
	"micheam.com/aico/internal/serve"
)

var flagSocket = &cli.StringFlag{
	Name:  "socket",
	Usage: "Unix socket `PATH` to listen on (default: see `aico serve path`)",
}

var CmdServe = &cli.Command{
	Name:  "serve",
	Usage: "run the completion daemon for editor clients",
	Description: `Listens on a Unix domain socket and serves code completion requests
from editors (vim-aico) over JSON-RPC. Editors start it on demand; one
daemon is shared by every editor instance of the same user.

The daemon writes nothing to stdout or stderr; see the log file.`,
	Flags:  []cli.Flag{flagSocket},
	Action: runServe,
	Commands: []*cli.Command{
		{
			Name:   "path",
			Usage:  "print the socket path the daemon would listen on",
			Flags:  []cli.Flag{flagSocket},
			Action: runServePath,
		},
	},
}

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func socketPath(cmd *cli.Command) string {
	if p := cmd.String(flagSocket.Name); p != "" {
		return p
	}
	return serve.DefaultSocketPath()
}

func runServePath(ctx context.Context, cmd *cli.Command) error {
	_, err := fmt.Fprintln(cmd.Root().Writer, socketPath(cmd))
	return err
}

func runServe(ctx context.Context, cmd *cli.Command) error {
	logger, cleanup, err := initializeLogger(ctx, cmd)
	if err != nil {
		return err
	}
	defer cleanup()
	ctx = logging.ContextWith(ctx, logger)

	conf, err := config.LoadOrDefault()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	path := socketPath(cmd)
	ln, err := serve.Listen(path)
	if errors.Is(err, serve.ErrAlreadyRunning) {
		// An editor started us while another instance was already up:
		// not an error, the editor will connect to that one.
		logger.Info("another daemon is already running", "socket", path)
		return nil
	}
	if err != nil {
		return err
	}
	defer ln.Close()

	srv := serve.New(serve.Options{
		Config:       conf,
		ReloadConfig: config.LoadOrDefault,
		Version:      version,
		NewCompleter: newCompleter,
		Logger:       logger,
	})

	// Editors spawn the daemon as a child job; SIGHUP from their terminal
	// closing must not take it down. SIGINT/SIGTERM shut it down cleanly.
	signal.Ignore(syscall.SIGHUP)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case sig := <-sigs:
			logger.Info("signal received, shutting down", "signal", sig.String())
			srv.Shutdown()
		case <-srv.Done():
		}
	}()

	logger.Info("listening", "socket", path, "version", version)
	defer logger.Info("stopped")
	return srv.Serve(ln)
}

func newCompleter(p config.CompleteProfile) complete.Completer {
	return llamacpp.New(llamacpp.Options{
		Endpoint:      p.Endpoint,
		Model:         p.Model,
		APIKey:        p.APIKey(),
		NPredict:      p.NPredict,
		TMaxPromptMS:  p.TMaxPromptMS,
		TMaxPredictMS: p.TMaxPredictMS,
	})
}
