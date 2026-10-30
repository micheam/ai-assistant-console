package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"slices"
)

// CompleteConfig is the [complete] section: named profiles describing where and how
// to request code completions (fill-in-the-middle style).
type CompleteConfig struct {
	// DefaultProfile names the profile used when a client does not ask for
	// a specific one. It may be omitted when exactly one profile exists.
	DefaultProfile string `toml:"default_profile,omitempty"`

	// Profiles maps a profile name to its settings ([complete.profiles.<name>]).
	Profiles map[string]CompleteProfile `toml:"profiles,omitempty"`
}

// CompleteProfile describes one completion backend and the budgets used with it.
//
// Zero numeric values mean "use the default"; see [CompleteProfile.WithDefaults].
type CompleteProfile struct {
	// Endpoint is the full URL of the completion endpoint, including its
	// path (e.g. "http://127.0.0.1:8012/infill").
	Endpoint string `toml:"endpoint"`

	// Model is sent as the request's model field. Empty omits it.
	Model string `toml:"model,omitempty"`

	// APIKeyEnv names the environment variable holding the API key, so the
	// key itself does not have to be written in the config file.
	APIKeyEnv string `toml:"api_key_env,omitempty"`

	// NPredict is the maximum number of tokens to generate per completion.
	NPredict int `toml:"n_predict,omitempty"`

	// TMaxPromptMS and TMaxPredictMS bound the time the backend may spend
	// on prompt processing and on generation, in milliseconds.
	TMaxPromptMS  int `toml:"t_max_prompt_ms,omitempty"`
	TMaxPredictMS int `toml:"t_max_predict_ms,omitempty"`

	// PrefixLines and SuffixLines are how many lines before and after the
	// cursor an editor client should send as context.
	PrefixLines int `toml:"prefix_lines,omitempty"`
	SuffixLines int `toml:"suffix_lines,omitempty"`

	// Candidates is how many suggestions one request generates when the
	// client does not ask for a specific number.
	Candidates int `toml:"candidates,omitempty"`
}

// Defaults applied by [CompleteProfile.WithDefaults]. They match llama.vim's
// defaults, so a profile behaves like llama.vim unless configured otherwise.
const (
	DefaultCompleteNPredict      = 128
	DefaultCompleteTMaxPromptMS  = 500
	DefaultCompleteTMaxPredictMS = 1000
	DefaultCompletePrefixLines   = 256
	DefaultCompleteSuffixLines   = 64
	DefaultCompleteCandidates    = 1
)

var (
	// ErrNoCompleteProfile is returned when a completion profile is requested but none
	// is configured, or several are configured without a default_profile.
	ErrNoCompleteProfile = errors.New("no completion profile selected")

	// ErrUnknownCompleteProfile is returned when the requested profile name is
	// not defined under [complete.profiles].
	ErrUnknownCompleteProfile = errors.New("unknown completion profile")
)

// WithDefaults returns a copy of p with every zero numeric field replaced by
// its default value.
func (p CompleteProfile) WithDefaults() CompleteProfile {
	setDefault := func(v *int, def int) {
		if *v == 0 {
			*v = def
		}
	}
	setDefault(&p.NPredict, DefaultCompleteNPredict)
	setDefault(&p.TMaxPromptMS, DefaultCompleteTMaxPromptMS)
	setDefault(&p.TMaxPredictMS, DefaultCompleteTMaxPredictMS)
	setDefault(&p.PrefixLines, DefaultCompletePrefixLines)
	setDefault(&p.SuffixLines, DefaultCompleteSuffixLines)
	setDefault(&p.Candidates, DefaultCompleteCandidates)
	return p
}

// APIKey returns the value of the environment variable named by APIKeyEnv,
// or "" when APIKeyEnv is empty or the variable is unset.
func (p CompleteProfile) APIKey() string {
	if p.APIKeyEnv == "" {
		return ""
	}
	return os.Getenv(p.APIKeyEnv)
}

func (p CompleteProfile) validate() error {
	if p.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	u, err := url.Parse(p.Endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("endpoint %q is not an http(s) URL", p.Endpoint)
	}
	for name, v := range map[string]int{
		"n_predict":        p.NPredict,
		"t_max_prompt_ms":  p.TMaxPromptMS,
		"t_max_predict_ms": p.TMaxPredictMS,
		"prefix_lines":     p.PrefixLines,
		"suffix_lines":     p.SuffixLines,
		"candidates":       p.Candidates,
	} {
		if v < 0 {
			return fmt.Errorf("%s must not be negative: %d", name, v)
		}
	}
	return nil
}

func (f CompleteConfig) validate() error {
	for name, p := range f.Profiles {
		if err := p.validate(); err != nil {
			return fmt.Errorf("profiles.%q: %w", name, err)
		}
	}
	if f.DefaultProfile != "" {
		if _, ok := f.Profiles[f.DefaultProfile]; !ok {
			return fmt.Errorf("default_profile %q is not defined under [complete.profiles]", f.DefaultProfile)
		}
	}
	return nil
}

// DefaultCompleteProfileName returns the profile used when none is requested:
// default_profile if set, otherwise the only profile if exactly one exists,
// otherwise "".
func (c *Config) DefaultCompleteProfileName() string {
	if c.Complete.DefaultProfile != "" {
		return c.Complete.DefaultProfile
	}
	if len(c.Complete.Profiles) == 1 {
		for name := range c.Complete.Profiles {
			return name
		}
	}
	return ""
}

// CompleteProfileNames returns the configured profile names in sorted order.
func (c *Config) CompleteProfileNames() []string {
	names := make([]string, 0, len(c.Complete.Profiles))
	for name := range c.Complete.Profiles {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// CompleteProfile resolves the named profile, or the default profile when name
// is empty, and returns it with defaults applied along with its name.
//
// errors:
//
//   - [ErrNoCompleteProfile]: name is empty and no default can be determined.
//   - [ErrUnknownCompleteProfile]: the profile is not defined.
func (c *Config) CompleteProfile(name string) (string, CompleteProfile, error) {
	if name == "" {
		name = c.DefaultCompleteProfileName()
	}
	if name == "" {
		return "", CompleteProfile{}, fmt.Errorf("%w: set complete.default_profile or define exactly one profile (%d defined)",
			ErrNoCompleteProfile, len(c.Complete.Profiles))
	}
	p, ok := c.Complete.Profiles[name]
	if !ok {
		return "", CompleteProfile{}, fmt.Errorf("%w: %q", ErrUnknownCompleteProfile, name)
	}
	return name, p.WithDefaults(), nil
}
