package config

import (
	"bytes"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromReader_Complete(t *testing.T) {
	t.Run("parses [complete] profiles", func(t *testing.T) {
		src := `
[complete]
default_profile = "local"

[complete.profiles.local]
endpoint = "http://127.0.0.1:8012/infill"
model = "qwen2.5-coder-7b"
api_key_env = "LLAMA_API_KEY"
n_predict = 64
t_max_prompt_ms = 300
t_max_predict_ms = 800
prefix_lines = 128
suffix_lines = 32
candidates = 3
`
		conf, err := loadFromReader(strings.NewReader(src))
		require.NoError(t, err)
		assert.Equal(t, "local", conf.Complete.DefaultProfile)
		assert.Equal(t, CompleteProfile{
			Endpoint:      "http://127.0.0.1:8012/infill",
			Model:         "qwen2.5-coder-7b",
			APIKeyEnv:     "LLAMA_API_KEY",
			NPredict:      64,
			TMaxPromptMS:  300,
			TMaxPredictMS: 800,
			PrefixLines:   128,
			SuffixLines:   32,
			Candidates:    3,
		}, conf.Complete.Profiles["local"])
	})

	t.Run("no [complete]: leaves the section empty", func(t *testing.T) {
		conf, err := loadFromReader(strings.NewReader(`model = "claude-haiku-5-5"`))
		require.NoError(t, err)
		assert.Empty(t, conf.Complete.DefaultProfile)
		assert.Nil(t, conf.Complete.Profiles)
	})

	invalid := []struct {
		name    string
		src     string
		wantErr []string
	}{
		{
			name: "missing endpoint",
			src: `
[complete.profiles.local]
model = "m"
`,
			wantErr: []string{`profiles."local"`, "endpoint is required"},
		},
		{
			name: "endpoint without http scheme",
			src: `
[complete.profiles.local]
endpoint = "127.0.0.1:8012/infill"
`,
			wantErr: []string{`profiles."local"`, "not an http(s) URL"},
		},
		{
			name: "negative budget",
			src: `
[complete.profiles.local]
endpoint = "http://127.0.0.1:8012/infill"
t_max_predict_ms = -1
`,
			wantErr: []string{`profiles."local"`, "t_max_predict_ms must not be negative"},
		},
		{
			name: "default_profile not defined",
			src: `
[complete]
default_profile = "remote"

[complete.profiles.local]
endpoint = "http://127.0.0.1:8012/infill"
`,
			wantErr: []string{`default_profile "remote"`},
		},
	}
	for _, tt := range invalid {
		t.Run("rejects "+tt.name, func(t *testing.T) {
			_, err := loadFromReader(strings.NewReader(tt.src))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "complete:")
			for _, want := range tt.wantErr {
				assert.Contains(t, err.Error(), want)
			}
		})
	}
}

func TestDefaultConfig_EncodesWithoutComplete(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, toml.NewEncoder(&buf).Encode(DefaultConfig()))
	assert.NotContains(t, buf.String(), "complete")
}

func TestConfig_CompleteProfile(t *testing.T) {
	local := CompleteProfile{Endpoint: "http://127.0.0.1:8012/infill", Model: "local-model"}
	remote := CompleteProfile{Endpoint: "http://rack:8012/infill", NPredict: 32}

	tests := []struct {
		name     string
		cfg      CompleteConfig
		request  string
		wantName string
		wantErr  error
	}{
		{
			name:     "explicit name",
			cfg:      CompleteConfig{DefaultProfile: "local", Profiles: map[string]CompleteProfile{"local": local, "remote": remote}},
			request:  "remote",
			wantName: "remote",
		},
		{
			name:     "empty name uses default_profile",
			cfg:      CompleteConfig{DefaultProfile: "remote", Profiles: map[string]CompleteProfile{"local": local, "remote": remote}},
			wantName: "remote",
		},
		{
			name:     "empty name uses the only profile",
			cfg:      CompleteConfig{Profiles: map[string]CompleteProfile{"local": local}},
			wantName: "local",
		},
		{
			name:    "no profiles",
			cfg:     CompleteConfig{},
			wantErr: ErrNoCompleteProfile,
		},
		{
			name:    "several profiles without default_profile",
			cfg:     CompleteConfig{Profiles: map[string]CompleteProfile{"local": local, "remote": remote}},
			wantErr: ErrNoCompleteProfile,
		},
		{
			name:    "unknown name",
			cfg:     CompleteConfig{Profiles: map[string]CompleteProfile{"local": local}},
			request: "nope",
			wantErr: ErrUnknownCompleteProfile,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := &Config{Complete: tt.cfg}
			name, p, err := conf.CompleteProfile(tt.request)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, name)
			assert.Equal(t, tt.cfg.Profiles[tt.wantName].WithDefaults(), p)
		})
	}
}

func TestCompleteProfile_WithDefaults(t *testing.T) {
	t.Run("fills zero values", func(t *testing.T) {
		got := CompleteProfile{Endpoint: "http://h/infill"}.WithDefaults()
		assert.Equal(t, CompleteProfile{
			Endpoint:      "http://h/infill",
			NPredict:      DefaultCompleteNPredict,
			TMaxPromptMS:  DefaultCompleteTMaxPromptMS,
			TMaxPredictMS: DefaultCompleteTMaxPredictMS,
			PrefixLines:   DefaultCompletePrefixLines,
			SuffixLines:   DefaultCompleteSuffixLines,
			Candidates:    DefaultCompleteCandidates,
		}, got)
	})

	t.Run("keeps configured values", func(t *testing.T) {
		p := CompleteProfile{Endpoint: "http://h/infill", NPredict: 1, TMaxPromptMS: 2, TMaxPredictMS: 3, PrefixLines: 4, SuffixLines: 5, Candidates: 6}
		assert.Equal(t, p, p.WithDefaults())
	})
}

func TestCompleteProfile_APIKey(t *testing.T) {
	t.Setenv("AICO_TEST_COMPLETE_KEY", "secret")

	assert.Equal(t, "secret", CompleteProfile{APIKeyEnv: "AICO_TEST_COMPLETE_KEY"}.APIKey())
	assert.Empty(t, CompleteProfile{}.APIKey())
	assert.Empty(t, CompleteProfile{APIKeyEnv: "AICO_TEST_COMPLETE_KEY_UNSET"}.APIKey())
}
