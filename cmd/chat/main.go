package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	assistantconsole "micheam.com/aico/internal"
	"micheam.com/aico/internal/commands"
)

func main() {
	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var app = &cli.App{
	Name:           "chat",
	Usage:          "Chat with an AI assistant",
	Version:        assistantconsole.Version,
	Description:    "",
	DefaultCommand: "",
	Commands: []*cli.Command{
		commands.ChatSend,
		commands.ChatSession,
		commands.ChatModels,
		commands.Config,
	},
	Flags:                []cli.Flag{},
	EnableBashCompletion: true,
	Authors:              []*cli.Author{},
	Suggest:              true,
}
