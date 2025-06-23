package cmd

import (
	"fmt"
	"os"

	"github.com/msalemor/cai/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	skipLogging bool

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "cai",
		Short: "A tool for evaluating AI models",
		Long:  `cai (Code Analysis with Intelligence) is a tool that can be used in pipelines to add custom code evaluations.`,
		Run: func(cmd *cobra.Command, args []string) {
			// if evaluationName == "" || sourceFolder == "" {
			// 	fmt.Println("Error: both evaluation name and source folder must be specified")
			// 	cmd.Help()
			// 	os.Exit(1)
			// }
			// fmt.Printf("Running evaluation '%s' on source folder '%s'\n", evaluationName, sourceFolder)

			// pkg.Process(sourceFolder, evaluationName)
			// // Add your main logic here
			println("Run 'cai evaluate --help' for usage instructions.")
		},
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// // Define flags for source folder and evaluation name
	// rootCmd.Flags().StringVarP(&sourceFolder, "source", "s", ".", "Source folder containing the data to evaluate")
	// rootCmd.Flags().StringVarP(&evaluationName, "evaluation", "e", "complexity", "Name of the evaluation to run")
	//rootCmd.PersistentFlags().BoolVarP(&skipLogging, "skip-logging", "v", false, "Skip logging output during evaluation")

	// // Mark flags as required
	// //rootCmd.MarkFlagRequired("source")
	// rootCmd.MarkFlagRequired("evaluation")
	// Main entry point
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	// Initialize settings the environment variables or .env file
	//pkg.LoadOpenAISettings()
	logrus.Info("Settings initialized")
	pkg.LoadOpenAISettings()
	// Load evaluation prompts from evaluations.json file
	//pkg.LoadEvaluations()
	logrus.Info("Evaluations loaded")
	pkg.LoadEvaluations()
}
