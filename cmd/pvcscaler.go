package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"laghoule/pvcscaler/internal/pkg/pvcscaler"
)

const (
	batchSize = 10
)

var (
	Version   = "0.0.0-devel"
	GitCommit = "devel"
)

func main() {
	fmt.Printf("pvcscaler version %s / %s\n\n", Version, GitCommit)

	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	namespace := flag.String("namespace", "default", "namespace to use (default: default)")
	storageClass := flag.String("storage-class", "synology-csi-nas", "storage class to match")
	waitTime := flag.Int("wait-time", 10, "time to wait before scaleup (default: 30s)") // FIXME
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		fmt.Println("\nReceived interrupt, exiting gracefully...")
		cancel()
	}()

	waitTimeDuration := time.Duration(*waitTime) * time.Second

	pvcscaler := pvcscaler.NewPVCscaler("/tmp/test.json", waitTimeDuration, *storageClass)

	err := pvcscaler.Run(ctx, *kubeconfig, *namespace)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
