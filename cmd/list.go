package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List current TiDB clusters",
	Long:  "Get a list of the managed TiDB clusters",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("List")
		os.Exit(0)
	},
}
