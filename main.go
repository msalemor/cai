package main

import (
	"github.com/msalemor/cai/cmd"
	"github.com/msalemor/cai/pkg"
	"github.com/sirupsen/logrus"
)

func main() {
	// Configure logrus
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Main entry point
	logrus.Info("Starting application")

	// Initialize settings the environment variables or .env file
	pkg.LoadOpenAISettings()
	logrus.Info("Settings initialized")

	// Load evaluation prompts from evaluations.json file
	pkg.LoadEvaluations()
	logrus.Info("Evaluations loaded")

	cmd.Execute()
}
