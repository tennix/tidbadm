package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create the initial cluster configuration",
	Long:  `Generate the initial cluster configuration for further customization.`,
	Run: func(cmd *cobra.Command, args []string) {

		tc, err := runner.Init()
		if err != nil {
			panic(err)
		}
		d, err := yaml.Marshal(tc)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", d)
	},
}
