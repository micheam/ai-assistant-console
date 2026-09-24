package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

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
				// Output both simple and qualified names, including aliases, as completion candidates
				for _, model := range allAvailableModels() {
					fmt.Fprintln(cmd.Root().Writer, model.Name())
					fmt.Fprintln(cmd.Root().Writer, QualifiedName(model.Provider(), model.Name()))
					for _, alias := range aliasesOf(model.Provider(), model.Name()) {
						fmt.Fprintln(cmd.Root().Writer, alias)
						fmt.Fprintln(cmd.Root().Writer, QualifiedName(model.Provider(), alias))
					}
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
			Aliases:       aliasesOf(model.Provider(), model.Name()),
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
	Name          string   `json:"name"`
	QualifiedName string   `json:"qualified_name"`
	Provider      string   `json:"provider"`
	Aliases       []string `json:"aliases"`
	Description   string   `json:"description"`
	Selected      bool     `json:"selected"`
}

func (m *listItemView) String() string {
	s := m.QualifiedName
	if len(m.Aliases) > 0 {
		s += " (" + strings.Join(m.Aliases, ", ") + ")"
	}
	if m.Selected {
		s += " *"
	}
	return s
}

// aliasesOf returns the aliases of the given provider that resolve to
// modelName, sorted.
func aliasesOf(provider, modelName string) []string {
	p, ok := providerByName(provider)
	if !ok {
		return []string{}
	}
	aliases := []string{}
	for alias, target := range p.aliases() {
		if target == modelName {
			aliases = append(aliases, alias)
		}
	}
	slices.Sort(aliases)
	return aliases
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
	entry, ok := providerByName(provider)
	if !ok {
		return DefaultModel(cmd)
	}
	model, err := entry.newModel(modelName, cmd.String(entry.apiKeyFlag))
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
// It supports both simple names ("gpt-4.1") and qualified names ("openai:gpt-4.1").
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
// The model part may be an alias (e.g. "fable"), which resolves to the
// canonical model name of the provider it belongs to.
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
		if canonical, ok := lookupProviderModel(parsed.Provider, parsed.ModelName); ok {
			return parsed.Provider, canonical, true
		}
		return "", "", false
	}

	modelName = parsed.ModelName

	// Case 2: Check default provider first if set
	if defaultProvider != "" {
		if canonical, ok := lookupProviderModel(defaultProvider, modelName); ok {
			return defaultProvider, canonical, true
		}
	}

	// Case 3: Search all providers in order
	for _, p := range providers {
		if canonical, ok := lookupProviderModel(p.name, modelName); ok {
			return p.name, canonical, true
		}
	}

	return "", "", false
}

// providerEntry describes how to look up and construct models of one provider.
type providerEntry struct {
	name       string
	apiKeyFlag string
	describe   func(modelName string) (desc string, found bool)
	newModel   func(modelName, apiKey string) (assistant.GenerativeModel, error)
	aliases    func() map[string]string
}

// providers lists the supported providers in search order.
var providers = []providerEntry{
	{anthropic.ProviderName, flagAPIKeyAnthropic.Name, anthropic.DescribeModel, anthropic.NewGenerativeModel, anthropic.Aliases},
	{openai.ProviderName, flagAPIKeyOpenAI.Name, openai.DescribeModel, openai.NewGenerativeModel, openai.Aliases},
	{groq.ProviderName, flagAPIKeyGroq.Name, groq.DescribeModel, groq.NewGenerativeModel, groq.Aliases},
	{cerebras.ProviderName, flagAPIKeyCerebras.Name, cerebras.DescribeModel, cerebras.NewGenerativeModel, cerebras.Aliases},
}

func providerByName(name string) (providerEntry, bool) {
	for _, p := range providers {
		if p.name == name {
			return p, true
		}
	}
	return providerEntry{}, false
}

// lookupProviderModel checks if a provider supports the given model name or
// alias, and returns the canonical model name.
func lookupProviderModel(provider, modelName string) (canonical string, found bool) {
	p, ok := providerByName(provider)
	if !ok {
		return "", false
	}
	if target, ok := p.aliases()[modelName]; ok {
		modelName = target
	}
	if _, found := p.describe(modelName); !found {
		return "", false
	}
	return modelName, true
}
