package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(destroyCmd)
}

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a TiDB cluster",
	Long:  "Destroy the specified TiDB cluster",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Destroy")
		os.Exit(0)
	},
}
