package pkg

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
)

type Evaluation struct {
	File        string  `json:"file"`
	Score       float64 `json:"score"`       // Score of the evaluation
	Explanation string  `json:"explanation"` // Explanation of the score
	Elapsed     float64 `json:"time"`        // Time taken for the evaluation
}

func evaluate(file, systemPrompt, sourceCode string) *Evaluation {
	// Placeholder for evaluation logic
	// This function should contain the logic to evaluate the source code files

	start := time.Now()
	eval, err := ChatCompletion(systemPrompt, sourceCode)
	elapsed := time.Since(start).Seconds()

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
		Elapsed:     elapsed,
	}
}

func CreateJUnitFile(evaluations []Evaluation, evaluationName, junitFileName string, totalElapsedTime float64) error {
	// Placeholder for JUNIT file creation logic
	// This function should create a JUNIT XML file based on the evaluations

	//junitFileName := fmt.Sprintf("%s", evaluationName)
	file, err := os.Create(junitFileName)
	if err != nil {
		return fmt.Errorf("failed to create JUNIT file: %w", err)
	}
	defer file.Close()

	// Write JUNIT XML structure
	file.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	file.WriteString(fmt.Sprintf("<testsuites time=\"%f\">\n", totalElapsedTime))
	for _, eval := range evaluations {
		file.WriteString(fmt.Sprintf("  <testcase classname=\"%s\" name=\"%s\" time=\"%f\">\n", evaluationName, eval.File, eval.Elapsed))
		if eval.Score < 5.0 {
			file.WriteString(fmt.Sprintf("    <failure message=\"Score: %.2f\">%s</failure>\n", eval.Score, eval.Explanation))
		}
		file.WriteString("  </testcase>\n")
	}
	file.WriteString("</testsuites>\n")

	logrus.Info(fmt.Sprintf("JUNIT file created: %s\n", junitFileName))
	return nil
}

func Process(sourceFolder, evaluationName, skipList, overrideList, junitFileName string) {
	logrus.Info("Scanning source files in folder: ", sourceFolder)
	sourceFiles, err := BuildSourceFileList(sourceFolder, skipList, overrideList)
	if err != nil {
		logrus.Warn("Failed to list source files: ", err.Error())
		os.Exit(1)
	}

	if len(sourceFiles) == 0 {
		println("No source files found in the specified folder")
		return
	}

	logrus.Infof("Getting evaluation prompt for: %s", evaluationName)
	evaluationPrompt := GetEvaluationPrompt(evaluationName)
	if evaluationPrompt == nil {
		panic(fmt.Sprintf("Evaluation prompt '%s' not found", evaluationName))
	}

	logrus.Info("Perform evaluations")
	var evaluationResults []Evaluation
	for _, file := range sourceFiles {
		// Process each file as needed
		// For example, you could read the file, analyze its content, etc.
		// Here we just print the file name for demonstration purposes
		logrus.Info("Processing file: ", file)
		sourceCode, err := os.ReadFile(file)
		if err != nil {
			panic(fmt.Sprintf("Failed to read file %s: %v\n", file, err))
		}

		// Skip empty files
		if len(sourceCode) != 0 {
			evaluation := evaluate(file, evaluationPrompt.SystemPrompt, string(sourceCode))
			evaluationResults = append(evaluationResults, *evaluation)
		}
	}

	// Create JUNIT file
	logrus.Info("Creating JUNIT file with the results")
	totalElapsedTime := 0.0
	for _, eval := range evaluationResults {
		totalElapsedTime += eval.Elapsed
	}

	err = CreateJUnitFile(evaluationResults, evaluationName, junitFileName, totalElapsedTime)
	if err != nil {
		logrus.Warn(fmt.Sprintf("Failed to create JUNIT file: %v\n", err))
		os.Exit(1)
	}

	// Prints the evaluation results
	logrus.Info("Printing the evaluation results")
	failure := false
	for _, eval := range evaluationResults {
		if eval.Score < 5.0 {
			color.Red("File: %s\nScore: %.2f\nReason:\n%s\n---------------------------------------------------------------------------\n", eval.File, eval.Score, eval.Explanation)
		} else {
			color.Green("File: %s\nScore: %.2f\nReason:\n%s\n---------------------------------------------------------------------------\n", eval.File, eval.Score, eval.Explanation)
		}
		if eval.Score < 5.0 {
			failure = true
		}
	}

	// If a failure occurred, print a message and exit with a error
	if failure {
		logrus.Warn("Failure: one or more files scored below the evaluation threshold: ", evaluationPrompt.Description)
		os.Exit(1)
	}
}
