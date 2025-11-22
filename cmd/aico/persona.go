package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
)

var CmdPersona = &cli.Command{
	Name:  "persona",
	Usage: "manage personas",

	// default action: list personas
	Action: runListPersonas,
	Commands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "list available personas",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "json",
					Usage: "output in JSON format",
				},
			},
			Action: runListPersonas,
		},
	},
}

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runListPersonas(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	currentPersona := cmd.String(flagPersona.Name)

	personas := []personaListItemView{}
	for name, persona := range conf.PersonaMap {
		personas = append(personas, personaListItemView{
			Name:        name,
			Description: persona.Description,
			Selected:    name == currentPersona,
		})
	}

	// Sort by name for consistent output
	sort.Slice(personas, func(i, j int) bool {
		return personas[i].Name < personas[j].Name
	})

	if cmd.Bool(flagJSON.Name) {
		encoder := json.NewEncoder(cmd.Root().Writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(personas)
	}

	for _, p := range personas {
		fmt.Fprintln(cmd.Root().Writer, p.String())
	}
	return nil
}

type personaListItemView struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
}

func (p *personaListItemView) String() string {
	if p.Selected {
		return fmt.Sprintf("%s *", p.Name)
	}
	return p.Name
}
