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

	pkg.GetSettings() // Initialize settings from JSON file
	logrus.Info("Settings initialized")

	pkg.EvaluationsInstance() // Load evaluation prompts from JSON file
	logrus.Info("Evaluations loaded")

	cmd.Execute()
}
