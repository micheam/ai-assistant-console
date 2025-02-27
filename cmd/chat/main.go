package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	assistantconsole "micheam.com/aico/internal"
	"micheam.com/aico/internal/commands"
)

func main() {
	err := newApp().Run(os.Args)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}

func newApp() *cli.App {
	return &cli.App{
		Name:           "chat",
		Usage:          "Chat with an AI assistant",
		Version:        assistantconsole.Version,
		Description:    "",
		DefaultCommand: "",
		Commands: []*cli.Command{
			commands.ChatSend,
			commands.Config,
		},
		Flags:                []cli.Flag{},
		EnableBashCompletion: true,
		Authors:              []*cli.Author{},
		Suggest:              true,
	}
}
