package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the configuration for the application
type Config struct {
	Chat Chat `yaml:"chat"`
}

// Chat is the configuration for the chat
type Chat struct {
	// Model is the model to use for the chat
	//
	// This can be one of [openai.chatAvailableModels].
	// If omitted, the default model for the application will be used.
	Model string `yaml:"model"`

	// Temperature is the temperature to use for the chat
	Temperature float64 `yaml:"temperature"`

	// Persona is the persona to use for the chat
	Persona Personality `yaml:"persona"`
}

// Personality is the personality to use for the chat
type Personality struct {
	// Description is the description of the personality
	Description string `yaml:"description"`

	// Messages is the list of messages to use for the personality
	Messages []string `yaml:"messages"`
}

var (
	ErrConfigFileNotFound = errors.New("config file not found")
)

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
	var config Config
	if err := yaml.Unmarshal(b, &config); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	return &config, nil
}

// configFilePath returns the path to the config file
func configFilePath(_ context.Context) string {
	if os.Getenv("CONFIG_PATH") != "" {
		return os.Getenv("CONFIG_PATH")
	}
	confDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(confDir, ApplicationName, ConfigFileName)
}

const (
	// DefaultModel is the default model to use for the chat
	DefaultModel = "gpt-3.5-turbo"

	// DefaultTemperature is the default temperature to use for the chat
	DefaultTemperature = 0.5

	// ApplicationName is the name of the application
	ApplicationName = "com.micheam.aico"

	// ConfigFileName is the name of the config file
	ConfigFileName = "config.yaml"
)

// Load loads the configuration for the application
//
// This will load the configuration from the path specified by the CONFIG_PATH
// environment variable, or from the default location if the environment
// variable is not set.
//
// This may return an error if the file cannot be read or parsed.
// If the file does not exist, this will return [ErrConfigFileNotFound].
func Load(ctx context.Context) (*Config, error) {
	return load(configFilePath(ctx))
}

// InitAndLoad initializes the configuration for the application
func InitAndLoad(ctx context.Context) (*Config, error) {
	config, err := Load(ctx)
	if err == nil {
		return config, nil // already initialized
	}
	if errors.Is(err, ErrConfigFileNotFound) {
		conf := DefaultConfig()
		// mkdir for path
		if err := os.MkdirAll(filepath.Dir(configFilePath(ctx)), 0755); err != nil {
			return nil, fmt.Errorf("mkdir: %w", err)
		}

		// write default config
		b, err := yaml.Marshal(conf)
		if err != nil {
			return nil, fmt.Errorf("marshal yaml: %w", err)
		}
		if err := os.WriteFile(configFilePath(ctx), b, 0644); err != nil {
			return nil, fmt.Errorf("write file: %w", err)
		}

		return conf, nil
	}

	// Unexpected error
	return nil, err
}

func DefaultConfig() *Config {
	return &Config{
		Chat: Chat{
			Model:       DefaultModel,
			Temperature: DefaultTemperature,
			Persona:     Personality{},
		},
	}
}
