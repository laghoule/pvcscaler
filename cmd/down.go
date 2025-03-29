package cmd

import (
	"context"
	"fmt"

	"laghoule/pvcscaler/internal/pkg/pvcscaler"

	"github.com/spf13/cobra"
)

var (
	outputFile string
)

var downCmd = &cobra.Command{
	Use:   "down",
	Short: "Scale down",
	Long:  `Scale down pods with pvc.`,
	Run: func(cmd *cobra.Command, args []string) {
		down()
	},
}

func init() {
	rootCmd.AddCommand(downCmd)
	downCmd.PersistentFlags().StringArrayVarP(&namespaces, "namespace", "n", []string{"all"}, "namespace to use")
	downCmd.PersistentFlags().StringVarP(&storageClass, "storageclass", "s", "default", "storage class to target")
	downCmd.PersistentFlags().StringVarP(&outputFile, "outputFile", "o", "", "pvscaler state file")
}

func down() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	processSignal(cancelFunc)

	if !validNamespaces(namespaces) {
		exitOnError(fmt.Errorf("error: invalid namespace, cannot mix `all` with other namespace\n"))
	}

	// FIXME: add file check

	pvcscaler, err := pvcscaler.New(ctx, kubeconfig, namespaces, storageClass, dryRun)
	exitOnError(err)

	err = pvcscaler.Down(outputFile)
	exitOnError(err)

}
