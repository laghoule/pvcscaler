package cmd

import (
	"context"
	"fmt"

	"laghoule/pvcscaler/internal/pkg/pvcscaler"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workloads with pvc",
	Long:  `List workloads with pvc in the specified storage class.`,
	Run: func(cmd *cobra.Command, args []string) {
		list()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.PersistentFlags().StringArrayVarP(&namespaces, "namespace", "n", []string{"all"}, "namespace to use")
	listCmd.PersistentFlags().StringVarP(&storageClass, "storageclass", "s", "default", "storage class to target")
}

func list() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	processSignal(cancelFunc)

	if !validNamespaces(namespaces) {
		exitOnError(fmt.Errorf("error: invalid namespace, cannot mix `all` with other namespace\n"))
	}

	pvcscaler, err := pvcscaler.New(ctx, kubeconfig, namespaces, storageClass, dryRun)
	exitOnError(err)

	err = pvcscaler.PrintList()
	exitOnError(err)
}
