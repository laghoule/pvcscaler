package pvcscaler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"laghoule/pvcscaler/internal/pkg/k8s"
)

type PVCscaler struct {
	workloads       []k8s.Workload
	scaleUpWaitTime time.Duration
	storageClass    string
}

func NewPVCscaler(file string, scaleUpWaitTime time.Duration, storageClass string) *PVCscaler {
	return &PVCscaler{
		scaleUpWaitTime: scaleUpWaitTime,
		storageClass:    storageClass,
	}
}

func (p *PVCscaler) Run(ctx context.Context, kubeconfig, namespace string) error {
	k8s, err := k8s.New(kubeconfig)
	if err != nil {
		panic(err)
	}

	var namespaces []string
	if namespace == "all" {
		namespaces, err = k8s.GetAllNamespaces(ctx)
		if err != nil {
			panic(err)
		}
	} else {
		namespaces = []string{namespace}
	}

	var dataset dataset
	var wg sync.WaitGroup
	var errChan = make(chan error, len(namespaces))

	for _, ns := range namespaces {
		wg.Add(1)
		go func(ns string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				p.workloads, err = k8s.GetWorkloads(ctx, ns, p.storageClass)
				if err != nil {
					errChan <- err
				}
				dataset.Workloads = append(dataset.Workloads, p.workloads...)
			}
		}(ns)
	}

	wg.Wait()
	close(errChan)

	fmt.Println(p.workloads)

	// for _, workload := range p.workloads {
	// 	err := workload.ScaleDown(ctx, k8s.ClientSet, workload.Namespace, workload.Name, workload.Kind)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	for err := range errChan {
		fmt.Printf("Error: %v\n", err)
	}

	return nil
}
