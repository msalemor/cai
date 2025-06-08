package pkg

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

type Evaluation struct {
	File        string  `json:"file"`
	Score       float64 `json:"score"`       // Score of the evaluation
	Explanation string  `json:"explanation"` // Explanation of the score
}

func evaluate(file, systemPrompt, sourceCode string) *Evaluation {
	// Placeholder for evaluation logic
	// This function should contain the logic to evaluate the source code files

	eval, err := ChatCompletion(systemPrompt, sourceCode)

	if err != nil {
		return &Evaluation{
			File:        file,
			Score:       -1.0, // Indicating an error in evaluation
			Explanation: fmt.Sprintf("Error evaluating file %s: %v", file, err),
		}
	}

	return &Evaluation{
		File:        file,
		Score:       eval.Score,
		Explanation: eval.Explanation,
	}
}

func Process(sourceFolder, evaluationName, skipList, overrideList string) {
	files, err := ListSourceFiles(sourceFolder, skipList, overrideList)
	if err != nil {
		panic("Failed to list source files: " + err.Error())
	}

	if len(files) == 0 {
		println("No source files found in the specified folder")
		return
	}

	evaluationPrompt := GetEvaluationPrompt(evaluationName)
	if evaluationPrompt == nil {
		panic(fmt.Sprintf("Evaluation prompt '%s' not found", evaluationName))
	}

	var evaluations []Evaluation
	for _, file := range files {
		// Process each file as needed
		// For example, you could read the file, analyze its content, etc.
		// Here we just print the file name for demonstration purposes
		println("Processing file:", file)

		sourceCode, err := os.ReadFile(file)
		if err != nil {
			panic(fmt.Sprintf("Failed to read file %s: %v\n", file, err))
		}
		//fmt.Printf("Read %d bytes from %s\n", len(sourceCode), file)

		evaluation := evaluate(file, evaluationPrompt.SystemPrompt, string(sourceCode))
		evaluations = append(evaluations, *evaluation)
	}

	failure := false
	for _, eval := range evaluations {
		if eval.Score < 5.0 {
			color.Red("File: %s\nScore: %.2f\nReason:\n%s\n---------------------------------------------------------------------------\n", eval.File, eval.Score, eval.Explanation)
		} else {
			color.Green("File: %s\nScore: %.2f\nReason:\n%s\n---------------------------------------------------------------------------\n", eval.File, eval.Score, eval.Explanation)
		}
		if eval.Score < 5.0 {
			failure = true
		}
	}

	if failure {
		fmt.Printf("\n\n%s\nFailure: one or more files scored below the evaluation threshold\n", evaluationPrompt.Description)
		os.Exit(1)
	}
}
