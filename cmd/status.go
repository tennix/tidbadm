package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the TiDB cluster status",
	Long:  "Query the status of the TiDB cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Status")
		os.Exit(0)
	},
}
