package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAvailableModels_DeprecationMetadata(t *testing.T) {
	models := AvailableModels()

	deprecatedModels := map[string]bool{
		"gpt-4o":      true,
		"gpt-4o-mini": true,
		"o1":          true,
		"o1-mini":     true,
	}

	for _, model := range models {
		t.Run(model.Name(), func(t *testing.T) {
			if deprecatedModels[model.Name()] {
				assert.True(t, model.Deprecated(), "model %s should be deprecated", model.Name())
				assert.NotEmpty(t, model.DeprecatedRemovedIn(), "model %s should have a RemovedIn version", model.Name())
			} else {
				assert.False(t, model.Deprecated(), "model %s should not be deprecated", model.Name())
				assert.Empty(t, model.DeprecatedRemovedIn(), "model %s should not have a RemovedIn version", model.Name())
			}
		})
	}
}
