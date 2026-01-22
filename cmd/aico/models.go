package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/logging"
	"micheam.com/aico/internal/providers/anthropic"
	"micheam.com/aico/internal/providers/cerebras"
	"micheam.com/aico/internal/providers/groq"
	"micheam.com/aico/internal/providers/openai"
)

var CmdModels = &cli.Command{
	Name:  "models",
	Usage: "manage AI models",

	// default action: list models
	Action: runListModels,
	Commands: []*cli.Command{
		{
			Name:    "list",
			Aliases: []string{"ls"},
			Usage:   "list available models",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "json",
					Usage: "output in JSON format",
				},
			},
			Action: runListModels,
		},
		{
			Name:      "describe",
			Aliases:   []string{"desc"},
			Usage:     "show model information",
			ArgsUsage: "MODEL",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "json",
					Usage: "output in JSON format",
				},
			},
			ShellComplete: func(ctx context.Context, cmd *cli.Command) {
				// Output both simple and qualified names as completion candidates
				for _, model := range allAvailableModels() {
					fmt.Fprintln(cmd.Root().Writer, model.Name())
					fmt.Fprintln(cmd.Root().Writer, QualifiedName(model.Provider(), model.Name()))
				}
			},
			Action: runDescribeModel,
		},
	},
}

// -----------------------------------------------------------------------------
// Actions
// -----------------------------------------------------------------------------

func runListModels(ctx context.Context, cmd *cli.Command) error {
	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Parse configured model spec to determine selection
	var selectedProvider, selectedModel string
	if conf.Model != "" {
		selectedProvider, selectedModel, _ = detectProviderByModelSpec(conf.Model, conf.DefaultProvider)
	}

	models := []listItemView{}
	for _, model := range allAvailableModels() {
		qualifiedName := QualifiedName(model.Provider(), model.Name())
		isSelected := model.Provider() == selectedProvider && model.Name() == selectedModel
		models = append(models, listItemView{
			Name:          model.Name(),
			QualifiedName: qualifiedName,
			Provider:      model.Provider(),
			Description:   model.Description(),
			Selected:      isSelected,
		})
	}
	if cmd.Bool(flagJSON.Name) {
		encoder := json.NewEncoder(cmd.Root().Writer)
		encoder.SetIndent("", "  ")
		return encoder.Encode(models)
	}
	for _, model := range models {
		fmt.Fprintln(cmd.Root().Writer, model.String())
	}
	return nil
}

func allAvailableModels() []assistant.ModelDescriptor {
	models := []assistant.ModelDescriptor{}
	models = append(models, anthropic.AvailableModels()...)
	models = append(models, openai.AvailableModels()...)
	models = append(models, groq.AvailableModels()...)
	models = append(models, cerebras.AvailableModels()...)
	return models
}

type listItemView struct {
	Name          string `json:"name"`
	QualifiedName string `json:"qualified_name"`
	Provider      string `json:"provider"`
	Description   string `json:"description"`
	Selected      bool   `json:"selected"`
}

func (m *listItemView) String() string {
	if m.Selected {
		return fmt.Sprintf("%s *", m.QualifiedName)
	}
	return m.QualifiedName
}

func runDescribeModel(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 1 {
		return fmt.Errorf("model name is required")
	}
	modelSpec := cmd.Args().Get(0)
	parsed := ParseModelSpec(modelSpec)

	for _, model := range allAvailableModels() {
		// Match by qualified name or simple name
		matchesQualified := parsed.Provider != "" &&
			model.Provider() == parsed.Provider &&
			model.Name() == parsed.ModelName
		matchesSimple := parsed.Provider == "" && model.Name() == parsed.ModelName

		if matchesQualified || matchesSimple {
			qualifiedName := QualifiedName(model.Provider(), model.Name())
			if cmd.Bool("json") {
				info := map[string]string{
					"name":           model.Name(),
					"qualified_name": qualifiedName,
					"provider":       model.Provider(),
					"description":    model.Description(),
				}
				encoder := json.NewEncoder(cmd.Root().Writer)
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			}

			fmt.Fprintf(cmd.Root().Writer, "Model: %s\n", model.Name())
			fmt.Fprintf(cmd.Root().Writer, "Qualified Name: %s\n", qualifiedName)
			fmt.Fprintf(cmd.Root().Writer, "Provider: %s\n", model.Provider())
			fmt.Fprintf(cmd.Root().Writer, "Description: %s\n", model.Description())
			return nil
		}
	}
	return fmt.Errorf("model not found: %s", modelSpec)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

type model struct {
	Name        string
	provider    string
	Description string
}

// DefaultModel returns the default model descriptor.
//
// Important:
//
//	Currently, only **Anthropic models** are supported as default model.
//	So, if an API key for Anthropic is not provided, it returns an error.
func DefaultModel(ctx context.Context, cmd *cli.Command) (assistant.GenerativeModel, error) {
	apikey := cmd.String(flagAPIKeyAnthropic.Name)
	if apikey != "" {
		return nil, errors.New(flagAPIKeyAnthropic.Name + " is required for default model, but not provided")
	}
	return anthropic.NewGenerativeModel(
		anthropic.AvailableModels()[0].Name(),
		apikey,
	)
}

// detectModel attempts to detect the model from the app configuration and command flags.
//
// Following is the detection priority:
//  1. If the --model flag is provided, use that model.
//  2. Otherwise, use the model specified in the configuration file.
//  3. If no model is specified in either place, return a default model.
//
// Model specification formats:
//   - Simple: "gpt-4o" (provider auto-detected, default_provider preferred if ambiguous)
//   - Qualified: "openai:gpt-4o" (explicit provider)
func detectModel(ctx context.Context, cmd *cli.Command) (assistant.GenerativeModel, error) {
	logger := logging.LoggerFrom(ctx)
	conf, err := config.Load()
	if err != nil {
		logger.Warn("failed to load config", "error", err)
		return DefaultModel(ctx, cmd)
	}

	modelSpec := cmd.String(flagModel.Name)
	if modelSpec == "" {
		modelSpec = conf.Model
	}
	if modelSpec == "" {
		return DefaultModel(ctx, cmd)
	}

	provider, modelName, found := detectProviderByModelSpec(modelSpec, conf.DefaultProvider)
	if !found {
		logger.Warn("unable to detect provider for model, using default model", "model", modelSpec)
		return DefaultModel(ctx, cmd)
	}

	switch provider {
	case anthropic.ProviderName:
		apikey := cmd.String(flagAPIKeyAnthropic.Name)
		return anthropic.NewGenerativeModel(modelName, apikey)
	case openai.ProviderName:
		apikey := cmd.String(flagAPIKeyOpenAI.Name)
		return openai.NewGenerativeModel(modelName, apikey)
	case groq.ProviderName:
		apikey := cmd.String(flagAPIKeyGroq.Name)
		return groq.NewGenerativeModel(modelName, apikey)
	case cerebras.ProviderName:
		apikey := cmd.String(flagAPIKeyCerebras.Name)
		return cerebras.NewGenerativeModel(modelName, apikey)
	default:
		logger.Warn("unsupported provider for model, using default model", "provider", provider, "model", modelName)
		return DefaultModel(ctx, cmd)
	}
}

// ModelSpec represents a parsed model specification.
// It supports both simple names ("gpt-4o") and qualified names ("openai:gpt-4o").
type ModelSpec struct {
	Provider  string // Provider name (empty if not specified)
	ModelName string // Model name
}

// ParseModelSpec parses a model specification string.
//
// Supported formats:
//   - "model-name" -> ModelSpec{Provider: "", ModelName: "model-name"}
//   - "provider:model-name" -> ModelSpec{Provider: "provider", ModelName: "model-name"}
func ParseModelSpec(spec string) ModelSpec {
	// Find the first colon
	for i, c := range spec {
		if c == ':' {
			return ModelSpec{
				Provider:  spec[:i],
				ModelName: spec[i+1:],
			}
		}
	}
	return ModelSpec{ModelName: spec}
}

// QualifiedName returns the fully qualified model name in "provider:model" format.
func QualifiedName(provider, modelName string) string {
	return provider + ":" + modelName
}

// detectProviderByModelName detects the provider for a given model specification.
//
// Detection priority:
//  1. If the spec contains an explicit provider (e.g., "groq:llama-3.3-70b"), use that.
//  2. If defaultProvider is set and supports the model, use that.
//  3. Otherwise, search providers in order: anthropic, openai, groq, cerebras.
//
// Returns the provider name, the actual model name, and whether the model was found.
func detectProviderByModelSpec(spec string, defaultProvider string) (provider string, modelName string, found bool) {
	parsed := ParseModelSpec(spec)

	// Case 1: Explicit provider in spec (e.g., "groq:llama-3.3-70b")
	if parsed.Provider != "" {
		if validateProviderModel(parsed.Provider, parsed.ModelName) {
			return parsed.Provider, parsed.ModelName, true
		}
		return "", "", false
	}

	modelName = parsed.ModelName

	// Case 2: Check default provider first if set
	if defaultProvider != "" {
		if validateProviderModel(defaultProvider, modelName) {
			return defaultProvider, modelName, true
		}
	}

	// Case 3: Search all providers in order
	providers := []string{
		anthropic.ProviderName,
		openai.ProviderName,
		groq.ProviderName,
		cerebras.ProviderName,
	}
	for _, p := range providers {
		if validateProviderModel(p, modelName) {
			return p, modelName, true
		}
	}

	return "", "", false
}

// validateProviderModel checks if a provider supports the given model name.
func validateProviderModel(provider, modelName string) bool {
	switch provider {
	case anthropic.ProviderName:
		_, found := anthropic.DescribeModel(modelName)
		return found
	case openai.ProviderName:
		_, found := openai.DescribeModel(modelName)
		return found
	case groq.ProviderName:
		_, found := groq.DescribeModel(modelName)
		return found
	case cerebras.ProviderName:
		_, found := cerebras.DescribeModel(modelName)
		return found
	default:
		return false
	}
}

// detectProvierByModelName is kept for backward compatibility.
// Deprecated: Use detectProviderByModelSpec instead.
func detectProvierByModelName(modelName string) (string, bool) {
	provider, _, found := detectProviderByModelSpec(modelName, "")
	return provider, found
}
