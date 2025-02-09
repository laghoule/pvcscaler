package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"laghoule/pvcscaler/internal/pkg/pvcscaler"

	"github.com/spf13/cobra"
)

var (
	namespaces []string
	storageClass string
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
	// TODO: cancel is duplicated
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		fmt.Println("\nReceived interrupt, cancelling...")
		cancel()
	}()

	if ! validNamespaces(namespaces) {
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
