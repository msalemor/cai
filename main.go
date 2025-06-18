package main

import (
	"github.com/msalemor/cai/cmd"
	"github.com/msalemor/cai/pkg"
)

func main() {
	// Main entry point
	pkg.GetSettings()         // Initialize settings from JSON file
	pkg.EvaluationsInstance() // Load evaluation prompts from JSON file
	cmd.Execute()
}
