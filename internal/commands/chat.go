package commands

import (
	"github.com/urfave/cli/v2"
)

var ChatSend = &cli.Command{
	Name:      "send",
	Usage:     "Send a message to the AI assistant",
	Action:    func(c *cli.Context) error { return nil },
	ArgsUsage: "<message>",
}
