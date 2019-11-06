package cmd

import (
	"os"

	rt "github.com/pingcap/tidbadm/runtime"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	pd     string
	config string
)

func init() {
	applyCmd.Flags().StringVarP(&pd, "pd", "", "", "pd endpoints separated by comma")
	applyCmd.Flags().StringVarP(&config, "config", "", "tidb-cluster.yaml", "tidb cluster configuration file")
	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a TiDB cluster configuration",
	Long:  `The TiDB cluster will be updated to the configuration specified`,
	Run: func(cmd *cobra.Command, args []string) {

		var tc rt.TidbCluster
		if config != "" {
			f, err := os.Open(config)
			if err != nil {
				panic(err)
			}
			dec := yaml.NewDecoder(f)
			if err := dec.Decode(&tc); err != nil {
				panic(err)
			}
		}

		if err := runner.Apply(tc); err != nil {
			panic(err)
		}
	},
}
