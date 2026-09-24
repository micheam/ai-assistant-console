package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"

	"micheam.com/aico/internal/assistant"
	"micheam.com/aico/internal/config"
	"micheam.com/aico/internal/providers/anthropic"
	"micheam.com/aico/internal/providers/cerebras"
	"micheam.com/aico/internal/providers/groq"
	"micheam.com/aico/internal/providers/openai"
	"micheam.com/aico/internal/theme"
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
			Action:  runListModels,
		},
		{
			Name:      "describe",
			Aliases:   []string{"desc"},
			Usage:     "show model information",
			ArgsUsage: "MODEL",
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
	provider, modelName, found := detectProviderByModelSpec(modelSpec, "")
	if !found {
		return fmt.Errorf("model not found: %s", modelSpec)
	}

	for _, model := range allAvailableModels() {
		if model.Provider() != provider || model.Name() != modelName {
			continue
		}
		qualifiedName := QualifiedName(model.Provider(), model.Name())
		if cmd.Bool(flagJSON.Name) {
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

		fmt.Fprintf(cmd.Root().Writer, "%s %s\n", theme.Bold("Model:"), model.Name())
		fmt.Fprintf(cmd.Root().Writer, "%s %s\n", theme.Bold("Qualified Name:"), qualifiedName)
		fmt.Fprintf(cmd.Root().Writer, "%s %s\n", theme.Bold("Provider:"), model.Provider())
		fmt.Fprintf(cmd.Root().Writer, "%s %s\n", theme.Bold("Description:"), model.Description())
		return nil
	}
	return fmt.Errorf("model not found: %s", modelSpec)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// DefaultModel returns the default model descriptor.
//
// Important:
//
//	Currently, only **Anthropic models** are supported as default model.
//	So, if an API key for Anthropic is not provided, it returns an error.
func DefaultModel(cmd *cli.Command) (assistant.GenerativeModel, error) {
	apikey := cmd.String(flagAPIKeyAnthropic.Name)
	if apikey == "" {
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
//  3. If no model is specified in any place, return a default model.
//
// Model specification formats:
//   - Simple: "gpt-4.1" (provider auto-detected, default_provider preferred if ambiguous)
//   - Qualified: "openai:gpt-4.1" (explicit provider)
func detectModel(cmd *cli.Command) (assistant.GenerativeModel, error) {
	conf, err := config.Load()
	if errors.Is(err, config.ErrConfigFileNotFound) {
		return DefaultModel(cmd)
	}
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	modelSpec := cmd.String(flagModel.Name)
	if modelSpec == "" {
		modelSpec = conf.Model
	}
	if modelSpec == "" {
		return DefaultModel(cmd)
	}

	return modelByName(cmd, modelSpec)
}

func modelByName(cmd *cli.Command, name string) (assistant.GenerativeModel, error) {
	conf, err := config.Load()
	if errors.Is(err, config.ErrConfigFileNotFound) {
		return DefaultModel(cmd)
	}
	if err != nil {
		// A malformed config must not silently degrade to the default
		// model: per-model settings would be dropped without notice.
		return nil, fmt.Errorf("load config: %w", err)
	}
	provider, modelName, found := detectProviderByModelSpec(name, conf.DefaultProvider)
	if !found {
		return DefaultModel(cmd)
	}
	var model assistant.GenerativeModel
	switch provider {
	case anthropic.ProviderName:
		apikey := cmd.String(flagAPIKeyAnthropic.Name)
		model, err = anthropic.NewGenerativeModel(modelName, apikey)
	case openai.ProviderName:
		apikey := cmd.String(flagAPIKeyOpenAI.Name)
		model, err = openai.NewGenerativeModel(modelName, apikey)
	case groq.ProviderName:
		apikey := cmd.String(flagAPIKeyGroq.Name)
		model, err = groq.NewGenerativeModel(modelName, apikey)
	case cerebras.ProviderName:
		apikey := cmd.String(flagAPIKeyCerebras.Name)
		model, err = cerebras.NewGenerativeModel(modelName, apikey)
	default:
		return DefaultModel(cmd)
	}
	if err != nil {
		return nil, err
	}
	if err := applyModelSettings(conf, model); err != nil {
		return nil, err
	}
	return model, nil
}

// applyModelSettings hands the [models."<name>"] settings from the config,
// if any, to the model. Settings are read from the config on every call
// rather than persisted in the session, so editing the config takes effect
// on the next turn.
func applyModelSettings(conf *config.Config, model assistant.GenerativeModel) error {
	key, settings, ok := conf.ModelSettings(model.Provider(), model.Name())
	if !ok {
		return nil
	}
	effort, err := settings.ResolveEffort(model.Provider())
	if err != nil {
		return fmt.Errorf("models.%q: %w", key, err)
	}
	capable, ok := model.(assistant.GenerationOptionCapable)
	if !ok {
		return fmt.Errorf("model %s does not support per-model settings (models.%q)", QualifiedName(model.Provider(), model.Name()), key)
	}
	capable.SetGenerationOptions(assistant.GenerationOptions{
		Effort:    effort,
		MaxTokens: settings.MaxTokens,
	})
	return nil
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

// detectProviderByModelSpec detects the provider for a given model specification.
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
