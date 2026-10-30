//go:build !unix

package main

import (
	"context"
	"errors"

	"github.com/urfave/cli/v3"
)

var CmdServe = &cli.Command{
	Name:  "serve",
	Usage: "run the completion daemon for editor clients (Unix only)",
	Action: func(context.Context, *cli.Command) error {
		return errors.New("aico serve is only available on Unix-like systems")
	},
}
