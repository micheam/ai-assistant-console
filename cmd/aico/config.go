package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
)

var CmdConfig = &cli.Command{
	Name:  "config",
	Usage: "Manage the configuration for the AI assistant",
	Commands: []*cli.Command{
		{
			Name:   "path",
			Usage:  "Show the path to the configuration file",
			Action: runShowConfigPath,
		},
		{
			Name:   "init",
			Usage:  "Initialize the configuration file",
			Action: runInitConfig,
		},
		{
			Name:   "edit",
			Usage:  "Edit the configuration file",
			Action: runEditConfig,
		},
	},
}

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runShowConfigPath(ctx context.Context, cmd *cli.Command) error {
	path := config.ConfigFilePath()
	_, err := fmt.Fprintln(cmd.Root().Writer, path)
	return err
}

func runInitConfig(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.InitAndLoad()
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.Root().Writer, "Configuration file initialized")
	fmt.Fprintln(cmd.Root().Writer, conf.Location())

	logger, cleanup, err := initializeLogger(ctx, cmd)
	if err != nil {
		return err
	}
	defer cleanup()
	logger.Info("Configuration file initialized", "path", conf.Location())

	return nil
}

func runEditConfig(ctx context.Context, cmd *cli.Command) error {
	editor, ok := os.LookupEnv("EDITOR")
	if !ok {
		editor = "vim"
		if runtime.GOOS == "windows" {
			editor = "notepad.exe"
		}
	}
	conf, err := loadConfig(ctx, cmd)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	ctx = config.WithConfig(ctx, conf)

	logger, cleanup, err := initializeLogger(ctx, cmd)
	if err != nil {
		return err
	}
	defer cleanup()

	cmdExec := exec.Command(editor, conf.Location())
	cmdExec.Stdin = os.Stdin
	cmdExec.Stdout = os.Stdout
	cmdExec.Stderr = os.Stderr
	if err := cmdExec.Run(); err != nil {
		return fmt.Errorf("edit configuration: %w", err)
	}

	logger.Debug("Configuration file edited")
	return nil
}
