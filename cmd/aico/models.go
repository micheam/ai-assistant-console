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
				// すべての利用可能なモデル名を補完候補として出力
				for _, model := range allAvailableModels() {
					fmt.Fprintln(cmd.Root().Writer, model.Name())
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
	models := []listItemView{}
	for _, model := range allAvailableModels() {
		models = append(models, listItemView{
			Name:        model.Name(),
			Provider:    model.Provider(),
			Description: model.Description(),
			Selected:    model.Name() == conf.Model,
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
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	Selected    bool   `json:"selected"`
}

func (m *listItemView) String() string {
	if m.Selected {
		return fmt.Sprintf("%s *", m.Name)
	}
	return m.Name
}

func runDescribeModel(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() < 1 {
		return fmt.Errorf("model name is required")
	}
	modelName := cmd.Args().Get(0)
	for _, model := range allAvailableModels() {
		if model.Name() == modelName {
			if cmd.Bool("json") {
				info := map[string]string{
					"name":        model.Name(),
					"provider":    model.Provider(),
					"description": model.Description(),
				}
				encoder := json.NewEncoder(cmd.Root().Writer)
				encoder.SetIndent("", "  ")
				return encoder.Encode(info)
			}

			fmt.Fprintf(cmd.Root().Writer, "Model: %s\n", model.Name())
			fmt.Fprintf(cmd.Root().Writer, "Provider: %s\n", model.Provider())
			fmt.Fprintf(cmd.Root().Writer, "Description: %s\n", model.Description())
			return nil
		}
	}
	return fmt.Errorf("model not found: %s", modelName)
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
// 1. If the --model flag is provided, use that model.
// 2. Otherwise, use the model specified in the configuration file.
// 3. If no model is specified in either place, return a default model.
func detectModel(ctx context.Context, cmd *cli.Command) (assistant.GenerativeModel, error) {
	logger := logging.LoggerFrom(ctx)
	conf, err := config.Load()
	if err != nil {
		logger.Warn("failed to load config", "error", err)
		return DefaultModel(ctx, cmd)
	}

	modelName := cmd.String(flagModel.Name)
	if modelName == "" {
		modelName = conf.Model
	}
	if modelName == "" {
		return DefaultModel(ctx, cmd)
	}

	provider, found := detectProvierByModelName(modelName)
	if !found {
		logger.Warn("unable to detect provider for model, using default model", "model", modelName)
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

func detectProvierByModelName(modelName string) (string, bool) {
	if _, found := anthropic.DescribeModel(modelName); found {
		return anthropic.ProviderName, true
	}
	if _, found := openai.DescribeModel(modelName); found {
		return openai.ProviderName, true
	}
	if _, found := groq.DescribeModel(modelName); found {
		return groq.ProviderName, true
	}
	if _, found := cerebras.DescribeModel(modelName); found {
		return cerebras.ProviderName, true
	}
	return "", false
}
