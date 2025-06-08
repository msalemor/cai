package main

import (
	"github.com/msalemor/cai/cmd"
	"github.com/msalemor/cai/pkg"
)

func main() {
	pkg.GetSettings()         // Initialize settings from JSON file
	pkg.EvaluationsInstance() // Load evaluation prompts from JSON file
	cmd.Execute()
}
