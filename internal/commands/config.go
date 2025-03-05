package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/urfave/cli/v2"

	"micheam.com/aico/internal/config"
)

var Config = &cli.Command{
	Name:  "config",
	Usage: "Manage the configuration for the AI assistant",
	Subcommands: []*cli.Command{
		configPath,
		configInit,
		configEdit,
	},
}

var configPath = &cli.Command{
	Name:  "path",
	Usage: "Show the path to the configuration file",
	Action: func(c *cli.Context) error {
		path := config.ConfigFilePath()
		_, err := fmt.Fprintln(c.App.Writer, path)
		return err
	},
}

var configInit = &cli.Command{
	Name:  "init",
	Usage: "Initialize the configuration file",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "path",
			Aliases: []string{"c"},
			EnvVars: []string{config.EnvKeyConfigPath},
			Usage:   "The path to the configuration file",
		},
	},
	Action: func(c *cli.Context) error {
		if c.String("path") != "" {
			// TODO: Make it configurable by other means than environment variables
			os.Setenv(config.EnvKeyConfigPath, c.String("path"))
		}
		conf, err := config.InitAndLoad()
		if err != nil {
			return err
		}
		fmt.Fprintln(c.App.Writer, "Configuration file initialized")
		fmt.Fprintln(c.App.Writer, conf.Location())
		return nil
	},
}

var configEdit = &cli.Command{
	Name:   "edit",
	Usage:  "Edit the configuration file",
	Before: LoadConfig,
	Action: func(c *cli.Context) error {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
			if runtime.GOOS == "windows" {
				editor = "notepad.exe"
			}
		}
		conf := config.MustFromContext(c.Context)
		cmd := exec.Command(editor, conf.Location())
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	},
}
