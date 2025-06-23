package cmd

import (
	"github.com/msalemor/cai/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List available evaluations",
	Long:  `List all available evaluations that can be run with the 'evaluate' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("Listing available evaluations...")
		evaluations := pkg.LoadEvaluations()

		if len(evaluations) == 0 {
			println("No evaluations found.")
			return
		}

		println("\nAvailable evaluations:\n")
		for _, eval := range evaluations {
			println("- " + eval.Name + ": " + eval.Description)
		}
	},
}

func init() {
	// Add the ls command to the root command
	rootCmd.AddCommand(lsCmd)
}
