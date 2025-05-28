package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
)

var Config = &cli.Command{
	Name:  "config",
	Usage: "Manage the configuration for the AI assistant",
	Commands: []*cli.Command{
		configPath,
		configInit,
		configEdit,
	},
}

var configPath = &cli.Command{
	Name:  "path",
	Usage: "Show the path to the configuration file",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		path := config.ConfigFilePath()
		_, err := fmt.Fprintln(cmd.Root().Writer, path)
		return err
	},
}

var configInit = &cli.Command{
	Name:  "init",
	Usage: "Initialize the configuration file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"cmd"},
			Sources: cli.EnvVars(config.EnvKeyConfigPath),
			Usage:   "The path to the configuration file",
		},
	},
	Action: func(ctx context.Context, cmd *cli.Command) error {
		if cmd.String("path") != "" {
			// TODO: Make it configurable by other means than environment variables
			os.Setenv(config.EnvKeyConfigPath, cmd.String("path"))
		}
		conf, err := config.InitAndLoad()
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.Root().Writer, "Configuration file initialized")
		fmt.Fprintln(cmd.Root().Writer, conf.Location())
		return nil
	},
}

var configEdit = &cli.Command{
	Name:  "edit",
	Usage: "Edit the configuration file",
	Action: func(ctx context.Context, cmd *cli.Command) error {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
			if runtime.GOOS == "windows" {
				editor = "notepad.exe"
			}
		}
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

		cmdExec := exec.Command(editor, conf.Location())
		cmdExec.Stdin = os.Stdin
		cmdExec.Stdout = os.Stdout
		cmdExec.Stderr = os.Stderr
		if err := cmdExec.Run(); err != nil {
			return fmt.Errorf("edit configuration: %w", err)
		}

		logger.Debug("Configuration file edited")
		return nil
	},
}
