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
	inputFile string
)

var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Scale up",
	Long:  `Scale up pods with pvc.`,
	Run: func(cmd *cobra.Command, args []string) {
		up()
	},
}

func init() {
	rootCmd.AddCommand(upCmd)
	upCmd.PersistentFlags().StringVarP(&inputFile, "inputFile", "i", "", "pvscaler state file")
}

func up() {
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

	pvcscaler, err := pvcscaler.New(kubeconfig, namespaces, storageClass)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(inputFile)

	err = pvcscaler.Up(ctx, inputFile)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
