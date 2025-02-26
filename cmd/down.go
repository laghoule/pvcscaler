package cmd

import (
	"context"
	"fmt"
	"os"

	"laghoule/pvcscaler/internal/pkg/pvcscaler"

	"github.com/spf13/cobra"
)

var (
	namespaces   []string
	storageClass string
	outputFile   string
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
		fmt.Printf("error: invalid namespace, cannot mix `all` with other namespace\n")
		os.Exit(1)
	}

	// FIXME: add file check

	pvcscaler, err := pvcscaler.New(kubeconfig, namespaces, storageClass, dryRun)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	err = pvcscaler.Down(ctx, outputFile)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

}

func validNamespaces(namespaces []string) bool {
	for _, namespace := range namespaces {
		if namespace == "all" && len(namespaces) > 1 {
			return false
		}
	}
	return true
}
