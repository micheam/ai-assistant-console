package config

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	want := &Config{
		Chat: Chat{
			Model: "gpt-3.5-turbo",
			Persona: map[string]Personality{
				"default": {
					Description: "Professional",
					Messages: []string{
						"Message1 for Professional",
						"Message2 for Professional",
						"Message3 for Professional",
						"Message4 for Professional",
					},
				},
			},
			Session: Session{
				Directory:    "./sessions",
				DirectoryRaw: "./sessions",
			},
		},
	}
	got, err := load("testdata/example.yaml")
	if assert.NoError(t, err) {
		if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(*want)); diff != "" {
			t.Errorf("LoadConfig() mismatch (-want +got):\n%s", diff)
		}
		assert.Contains(t, got.Logfile(), "com.micheam.aico/aico.log")
	}
}

func TestLoadConfig_Chat_ResolveEnv(t *testing.T) {
	t.Setenv("SOMEENVVAR", "/someotherdirecory")

	in := `chat:
  model: o1
  session:
    directory: "$SOMEENVVAR/sessions"
`

	r := bytes.NewReader([]byte(in))
	got, err := loadFromReader(r)
	require.NoError(t, err)
	want := &Config{
		Chat: Chat{
			Model: "o1",
			Session: Session{
				Directory:    "/someotherdirecory/sessions",
				DirectoryRaw: "$SOMEENVVAR/sessions",
			},
		},
	}
	if diff := cmp.Diff(want, got, cmpopts.IgnoreUnexported(*want)); diff != "" {
		t.Errorf("LoadConfig() mismatch (-want +got):\n%s", diff)
	}
}
