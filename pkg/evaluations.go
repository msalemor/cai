package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type EvaluationPrompt struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	SystemPrompt string `json:"systemPrompt"`
}

var (
	evaluationPrompts []EvaluationPrompt
	promptsOnce       sync.Once
)

func EvaluationsInstance() []EvaluationPrompt {
	promptsOnce.Do(func() {
		jsonData, err := os.ReadFile("./evaluations.json")
		if err != nil {
			panic(fmt.Sprintf("failed to read evaluation prompts file: %s", err))
		}

		err = json.Unmarshal(jsonData, &evaluationPrompts)
		if err != nil {
			panic(fmt.Sprintf("failed to unmarshal evaluation prompts: %s", err))
		}
	})

	return evaluationPrompts
}

func GetEvaluationPrompt(name string) *EvaluationPrompt {
	for _, prompt := range evaluationPrompts {
		if prompt.Name == name {
			return &prompt
		}
	}
	return nil
}
