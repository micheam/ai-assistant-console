package main

import (
	"context"
	"fmt"
	"net/mail"
	"os"

	"github.com/urfave/cli/v3"

	commands "micheam.com/aico/internal/commands"
)

var version = "dev" // set by ldflags during build

var cmd = &cli.Command{
	Name:    "chat",
	Usage:   "AI Assistant Console",
	Version: version,
	Authors: []any{
		mail.Address{Name: "Michito Maeda", Address: "michito.maeda@gmail.com"},
	},
	DefaultCommand: "help",
	Commands: []*cli.Command{

		commands.ChatModels,
		commands.ChatRepl,
		commands.ChatSend,
		commands.ChatSession,

		commands.Config,
	},
	Flags: []cli.Flag{
		cli.GenerateShellCompletionFlag,
	},
	Suggest: true,
}

func main() {
	ctx := context.Background()
	err := cmd.Run(ctx, os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
