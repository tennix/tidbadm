package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(planCmd)
}

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Plan for a TiDB cluster configuration",
	Long:  "Generate an execution plan for the configuration specified",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Plan")
		os.Exit(0)
	},
}
