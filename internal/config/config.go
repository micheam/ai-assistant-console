package config

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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
	location string `yaml:"-"`
	Chat     Chat   `yaml:"chat"`
	logfile  string `yaml:"logfile"` // logfile is the path to the logfile
}

// Location returns the location of the configuration file
func (c *Config) Location() string {
	return c.location
}

// Chat is the configuration for the chat
type Chat struct {
	// Model is the model to use for the chat
	//
	// This can be one of [openai.chatAvailableModels].
	// If omitted, the default model for the application will be used.
	Model string `yaml:"model"`

	// Persona is the persona to use for the chat
	Persona map[string]Personality `yaml:"persona"`

	// Session is the configuration for the chat session
	Session Session `yaml:"session"`
}

// Personality is the personality to use for the chat
type Personality struct {
	// Description is the description of the personality
	Description string `yaml:"description"`

	// Messages is the list of messages to use for the personality
	Messages []string `yaml:"messages"`
}

// Session
type Session struct {
	// Directory is the directory to store the chat session files
	// This can be a relative or absolute path.
	// Environment variables can be used in the path.
	// Sessions will be stored in this directory with the pattern [SessionFilePattern].
	//
	// e.g: "/Users/micheam/.aico/sessions"
	Directory string `yaml:"directory"`

	// DirectoryRaw is the raw directory path before environment variable expansion
	// This is used to store the raw directory path before expansion
	//
	// e.g: "$HOME/.aico/sessions"
	DirectoryRaw string `yaml:"-"`
}

var ErrConfigFileNotFound = errors.New("config file not found")

func (c *Config) Logfile() string {
	if c.logfile == "" {
		return defaultLogfilePath()
	}
	return c.logfile
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
	if err := yaml.NewDecoder(r).Decode(&config); err != nil {
		return nil, fmt.Errorf("decode yaml: %w", err)
	}

	// Resolve environment variables
	config.Chat.Session.DirectoryRaw = config.Chat.Session.Directory
	config.Chat.Session.Directory = os.ExpandEnv(config.Chat.Session.Directory)
	return &config, nil
}

const EnvKeyConfigPath = "AI_ASSISTANT_CONFIG_PATH"

// ConfigFilePath returns the path to the config file
func ConfigFilePath() string {
	if os.Getenv(EnvKeyConfigPath) != "" {
		return os.Getenv(EnvKeyConfigPath)
	}
	if os.Getenv("XDG_CONFIG_HOME") != "" {
		return filepath.Join(os.Getenv("XDG_CONFIG_HOME"), ApplicationName, ConfigFileName)
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, ApplicationName, ConfigFileName)
}

func defaultLogfilePath() string {
	// Same directory as the config file
	return filepath.Join(filepath.Dir(ConfigFilePath()), LogFileName)
}

const (
	// DefaultModel is the default model to use for the chat
	DefaultModel = "gpt-4o-mini"

	// ApplicationName is the name of the application
	ApplicationName = "com.micheam.aico"

	// ConfigFileName is the name of the config file
	ConfigFileName = "config.yaml"

	// LogFileName is the name of the log file
	LogFileName = "aico.log"

	// SessionFilePattern is the pattern for the session file
	SessionFilePattern = "{{.ID}}.pb"
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
		b, err := yaml.Marshal(conf)
		if err != nil {
			return nil, fmt.Errorf("marshal yaml: %w", err)
		}
		if err := os.WriteFile(ConfigFilePath(), b, 0644); err != nil {
			return nil, fmt.Errorf("write file: %w", err)
		}
		conf.location = ConfigFilePath()
		return conf, nil
	}

	// Unexpected error
	return nil, err
}

// GetDefaultPersona returns the default persona
func (c *Chat) GetDefaultPersona() *Personality {
	if p, ok := c.Persona["default"]; ok {
		return &p
	}
	return &defaultPersona
}

// GetPersona returns the persona with the given name
// If the persona does not exist, this will return nil.
func (c *Chat) GetPersona(name string) (*Personality, bool) {
	if p, ok := c.Persona[name]; ok {
		return &p, true
	}
	return nil, false
}

var defaultPersona Personality = Personality{
	Description: "Default",
	Messages: []string{
		"You're aico, my personal AI assistant.",
		"You're here to help me with my daily tasks.",
	},
}

func DefaultConfig() *Config {
	return &Config{
		logfile: defaultLogfilePath(),
		Chat: Chat{
			Model: DefaultModel,
			Persona: map[string]Personality{
				"default": defaultPersona,
			},
			Session: Session{
				Directory:    "./sessions",
				DirectoryRaw: "./sessions",
			},
		},
	}
}
