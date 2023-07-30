package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	want := &Config{
		Chat: Chat{
			Model:       "gpt-3.5-turbo",
			Temperature: 0.5,
			Persona: Personality{
				Description: "Professional",
				Messages: []string{
					"Message1 for Professional",
					"Message2 for Professional",
					"Message3 for Professional",
					"Message4 for Professional",
				},
			},
		},
	}
	got, err := load("testdata/example.yaml")
	if assert.NoError(t, err) {
		if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(*want)); diff != "" {
			t.Errorf("LoadConfig() mismatch (-want +got):\n%s", diff)
		}
		assert.EqualValues(t,
			"/Users/micheam/Library/Application Support/com.micheam.aico/aico.log",
			got.Logfile())
	}
}
