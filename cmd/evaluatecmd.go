package cmd

import (
	"github.com/msalemor/cai/pkg"
	"github.com/spf13/cobra"
)

var (
	sourceFolder   string
	evaluationName string
	skipList       string
	overrideList   string
	pparallel      bool
	junitFileName  string

	evaluateCmd = &cobra.Command{
		Use:     "evaluate",
		Aliases: []string{"eval"},
		Short:   "Run evaluation with the specified name at the given source folder",
		Long:    `Run evaluation with the specified name at the given source folder.`,
		Run: func(cmd *cobra.Command, args []string) {
			if sourceFolder == "" || evaluationName == "" {
				println("Error: both source folder and evaluation name must be specified")
				cmd.Help()
				return
			}
			if pparallel {
				pkg.PProcess(sourceFolder, evaluationName, skipList, overrideList)
			} else {
				pkg.Process(sourceFolder, evaluationName, skipList, overrideList, junitFileName)
			}
		},
	}
)

func init() {
	// Define flags for source folder and evaluation name
	evaluateCmd.Flags().StringVarP(&sourceFolder, "source", "s", ".", "folder containing the source code files to evaluate")
	evaluateCmd.Flags().StringVarP(&evaluationName, "evaluation", "e", "complexity", "Name of the evaluation to run")
	evaluateCmd.Flags().StringVarP(&skipList, "skip", "k", "", "Skip files matching the provided pattern (e.g. [.go,.py])")
	evaluateCmd.Flags().StringVarP(&overrideList, "override", "o", "", "Override files matching the provided pattern (e.g. [.go,.py])")
	evaluateCmd.Flags().StringVarP(&junitFileName, "junit", "j", "junit.xml", "Name of the JUNIT file to create")

	evaluateCmd.Flags().BoolVarP(&pparallel, "parallel", "p", false, "Run evaluations in parallel")

	// If this is meant to be a subcommand of a root command, you would add:
	rootCmd.AddCommand(evaluateCmd)
}
