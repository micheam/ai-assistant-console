package chat

import "micheam.com/aico/internal/openai/models"

const DefaultModel = models.GPT4OMini

func AvailableModels() []models.Model {
	return []models.Model{
		models.GPT4O,
		models.GPT4OMini,
		models.O1,
		models.O1Mini,
		models.O3Mini,
	}
}

func IsAvailableModel(model models.Model) bool {
	for _, m := range AvailableModels() {
		if m == model {
			return true
		}
	}
	return false
}
