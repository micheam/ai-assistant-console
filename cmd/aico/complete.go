package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/config"
)

var CmdComplete = &cli.Command{
	Name:  "complete",
	Usage: "manage code completion settings",
	Commands: []*cli.Command{
		{
			Name:   "profiles",
			Usage:  "list completion profiles defined under [complete.profiles]",
			Action: runListCompleteProfiles,
		},
	},
}

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runListCompleteProfiles(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.LoadOrDefault()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	defaultName := conf.DefaultCompleteProfileName()
	profiles := []completeProfileListItemView{}
	for _, name := range conf.CompleteProfileNames() {
		p := conf.Complete.Profiles[name]
		profiles = append(profiles, completeProfileListItemView{
			Name:     name,
			Endpoint: p.Endpoint,
			Model:    p.Model,
			Default:  name == defaultName,
		})
	}

	w := cmd.Root().Writer
	if cmd.Bool(flagJSON.Name) {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(profiles)
	}
	for _, p := range profiles {
		fmt.Fprintln(w, p.String())
	}
	return nil
}

type completeProfileListItemView struct {
	Name     string `json:"name"`
	Endpoint string `json:"endpoint"`
	Model    string `json:"model"`
	Default  bool   `json:"default"`
}

func (p *completeProfileListItemView) String() string {
	s := p.Name
	if p.Model != "" {
		s += " (" + p.Model + ")"
	}
	s += " " + p.Endpoint
	if p.Default {
		s += " *"
	}
	return s
}
