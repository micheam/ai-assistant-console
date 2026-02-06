package anthropic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"micheam.com/aico/internal/assistant"
)

func TestAvailableModels_DeprecationMetadata(t *testing.T) {
	models := AvailableModels()

	deprecatedModels := map[string]bool{
		"claude-opus-4-5": true,
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

func TestClaudeOpus4_5_ImplementsModelDescriptor(t *testing.T) {
	var _ assistant.ModelDescriptor = (*ClaudeOpus4_5)(nil)
}
