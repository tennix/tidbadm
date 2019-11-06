package cmd

import (
	"fmt"
	"os"

	rt "github.com/pingcap/tidbadm/runtime"
	"github.com/pingcap/tidbadm/runtime/ansible"
	"github.com/pingcap/tidbadm/runtime/k8s"
	"github.com/spf13/cobra"
)

var (
	runtime string
	runner  rt.Runner
)

var (
	rootCmd = &cobra.Command{
		Use:   "tidbadm",
		Short: "Easily maintain TiDB clusters",
		Long: `tidbadm is a command line tool
to help maintain the lifecycle of TiDB clusters.`,
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&runtime, "runtime", "r", "ansible", "runtime: ansible | k8s")

	switch runtime {
	case string(rt.AnsibleRuntime):
		runner = ansible.New()
	case string(rt.K8sRuntime):
		runner = k8s.New()
	}
}
