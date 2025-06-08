package pkg

import (
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
)

func pevaluate(file, systemPrompt, sourceCode string) *Evaluation {
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

func PProcess(sourceFolder, evaluationName, skipList, overrideList string) {
	files, err := ListSourceFiles(sourceFolder, skipList, overrideList)
	if err != nil {
		panic("Failed to list source files: " + err.Error())
	}

	if len(files) == 0 {
		panic("No source files found in the specified folder")
	}

	evaluationPrompt := GetEvaluationPrompt(evaluationName)
	if evaluationPrompt == nil {
		panic(fmt.Sprintf("Evaluation prompt '%s' not found", evaluationName))
	}

	var wg sync.WaitGroup
	evaluationsChan := make(chan Evaluation, len(files))

	for _, file := range files {
		wg.Add(1)
		// Start a goroutine for each file
		go func(file string) {
			defer wg.Done()

			println("Processing file:", file)

			sourceCode, err := os.ReadFile(file)
			if err != nil {
				fmt.Printf("Failed to read file %s: %v\n", file, err)
				evaluationsChan <- Evaluation{
					File:        file,
					Score:       -1.0,
					Explanation: fmt.Sprintf("Failed to read file: %v", err),
				}
				return
			}

			evaluation := pevaluate(file, evaluationPrompt.SystemPrompt, string(sourceCode))
			evaluationsChan <- *evaluation
		}(file)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(evaluationsChan)
	}()

	// Collect results
	var evaluations []Evaluation
	for eval := range evaluationsChan {
		evaluations = append(evaluations, eval)
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
