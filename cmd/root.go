package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var (
	kubeconfig string
	dryRun     bool

	version   = "devel"
	gitCommit = "0000000000000000000000000000000000000000"
	buildDate = time.DateTime
)

var rootCmd = &cobra.Command{
	Use:               "pvcscaler",
	Short:             "pvcscaler is a pod with pvc scale down & up",
	Long:              `A Fast and and easy way to scale down and up pods with pvc.`,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	Run: func(cmd *cobra.Command, args []string) {
		//TODO: Do Stuff Here
	},
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	rootCmd.PersistentFlags().StringP("kubeconfig", "k", kubeconfig, "path to kubeconfig")
	rootCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "dry run mode")
}

func processSignal(cancelFunc context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		fmt.Println("\nReceived interrupt, cancelling...")
		cancelFunc()
	}()
}