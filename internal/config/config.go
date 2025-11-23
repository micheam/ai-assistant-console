package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"micheam.com/aico/internal/providers/anthropic"
)

type contextKey int

const (
	contextKeyConfig contextKey = iota
)

func WithConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextKeyConfig, cfg)
}

func FromContext(ctx context.Context) (*Config, bool) {
	cfg, ok := ctx.Value(contextKeyConfig).(*Config)
	return cfg, ok
}

func MustFromContext(ctx context.Context) *Config {
	cfg, ok := FromContext(ctx)
	if !ok {
		panic("config not found in context")
	}
	return cfg
}

// Config is the configuration for the application
type Config struct {
	location string `toml:"-"`

	// Logfile is the path to the logfile
	logfile string `toml:"logfile"`

	// Model is the model to use for text generation
	//
	// If omitted, the default model for the application will be used.
	Model string `toml:"model"`

	// PersonaMap is the persona to use for text generation
	PersonaMap map[string]Personality `toml:"persona"`
}

// Location returns the location of the configuration file
func (c *Config) Location() string {
	return c.location
}

// Personality is the personality to use for text generation
type Personality struct {
	// Description is the description of the personality
	Description string `toml:"description"`

	// Message is the system message to use for the personality
	Message string `toml:"message"`
}

var ErrConfigFileNotFound = errors.New("config file not found")

func (c *Config) Logfile() string {
	if c.logfile == "" {
		return defaultLogfilePath()
	}
	if filepath.IsAbs(c.logfile) {
		return c.logfile
	}
	return filepath.Join(filepath.Dir(c.location), c.logfile)
}

// OpenLogfile opens the logfile for writing
//
// This will create the logfile if it does not exist.
// It will also create the directory if it does not exist.
//
// Make sure to close the returned file when done.
func (c *Config) OpenLogfile() (*os.File, error) {
	logfilePath := c.Logfile()
	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(logfilePath), 0755); err != nil {
		return nil, fmt.Errorf("mkdir all: %w", err)
	}
	f, err := os.OpenFile(logfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("open logfile: %w", err)
	}
	return f, nil
}

// load loads the configuration from the given path
//
// This may return an error if the file cannot be read or parsed.
func load(path string) (*Config, error) {
	// Check if the file exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, ErrConfigFileNotFound
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	config, err := loadFromReader(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("load from reader: %w", err)
	}
	config.location = path
	return config, nil
}

func loadFromReader(r io.Reader) (*Config, error) {
	var config Config
	if _, err := toml.NewDecoder(r).Decode(&config); err != nil {
		return nil, fmt.Errorf("decode toml: %w", err)
	}
	return &config, nil
}

const EnvKeyConfigPath = "AI_ASSISTANT_CONFIG_PATH"

// ConfigFilePath returns the path to the config file
func ConfigFilePath() string {
	if os.Getenv(EnvKeyConfigPath) != "" {
		return os.Getenv(EnvKeyConfigPath)
	}
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		return filepath.Join(os.Getenv("XDG_CONFIG_HOME"), ApplicationFQN, ConfigFileName)
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, ApplicationFQN, ConfigFileName)
}

func defaultLogfilePath() string {
	// Same directory as the config file
	return filepath.Join(filepath.Dir(ConfigFilePath()), LogFileName)
}

const (
	// DefaultModel is the default model to use
	DefaultModel = anthropic.ModelNameClaudeHaiku4_5

	// ApplicationFQN is the fully qualified name of the application
	ApplicationFQN = "com.micheam.aico"

	// ConfigFileName is the name of the config file
	ConfigFileName = "config.toml"

	// LogFileName is the name of the log file
	LogFileName = "aico.log"
)

// Load loads the configuration for the application
//
// This will load the configuration from the path specified by the AI_ASSISTANT_CONFIG_PATH
// environment variable, or from the default location if the environment
// variable is not set.
//
// This may return an error if the file cannot be read or parsed.
// If the file does not exist, this will return [ErrConfigFileNotFound].
func Load() (*Config, error) {
	return load(ConfigFilePath())
}

// InitAndLoad initializes the configuration for the application
func InitAndLoad() (*Config, error) {
	config, err := Load()
	if err == nil {
		return config, nil // already initialized
	}
	if errors.Is(err, ErrConfigFileNotFound) {
		conf := DefaultConfig()
		// mkdir for path
		if err := os.MkdirAll(filepath.Dir(ConfigFilePath()), 0755); err != nil {
			return nil, fmt.Errorf("mkdir: %w", err)
		}

		// write default config
		var buf bytes.Buffer
		if err := toml.NewEncoder(&buf).Encode(conf); err != nil {
			return nil, fmt.Errorf("encode toml: %w", err)
		}
		if err := os.WriteFile(ConfigFilePath(), buf.Bytes(), 0644); err != nil {
			return nil, fmt.Errorf("write file: %w", err)
		}
		conf.location = ConfigFilePath()
		return conf, nil
	}

	// Unexpected error
	return nil, err
}

// GetDefaultPersona returns the default persona
func (c *Config) GetDefaultPersona() *Personality {
	if p, ok := c.PersonaMap["default"]; ok {
		return &p
	}
	return DefaultConfig().GetDefaultPersona()
}

// GetPersona returns the persona with the given name
// If the persona does not exist, this will return nil.
func (c *Config) GetPersona(name string) (*Personality, bool) {
	if p, ok := c.PersonaMap[name]; ok {
		return &p, true
	}
	return nil, false
}

// DefaultConfig returns the default configuration
// This is used when initializing the configuration for the first time.
func DefaultConfig() *Config {
	return &Config{
		logfile: defaultLogfilePath(),
		Model:   DefaultModel,
		PersonaMap: map[string]Personality{
			"default": {
				Description: "Default",
				Message: "You're aico, my personal AI assistant." +
					"You're here to help me with my daily tasks.",
			},
		},
	}
}
