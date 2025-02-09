package pvcscaler

import (
	"context"
	"fmt"
	"sync"

	"laghoule/pvcscaler/internal/pkg/k8s"
)

type PVCscaler struct {
	k8sClient    *k8s.Client
	workloads    []k8s.Workload
	namespaces   []string
	storageClass string
}

func New(kubeconfig string, namespaces []string, storageClass string) (*PVCscaler, error) {
	k8sClient, err := k8s.New(kubeconfig)
	if err != nil {
		return nil, err
	}

	return &PVCscaler{
		k8sClient:    k8sClient,
		namespaces:   namespaces,
		storageClass: storageClass,
	}, nil
}

func (p *PVCscaler) getWorkloads(ctx context.Context, k8sClient *k8s.Client, namespaces []string, storageClass string) error {
	var err error

	if len(namespaces) == 1 && namespaces[0] == "all" {
		namespaces, err = k8sClient.GetAllNamespaces(ctx)
		if err != nil {
			panic(err)
		}
	}

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
				p.workloads, err = k8sClient.GetWorkloads(ctx, ns, storageClass)
				if err != nil {
					errChan <- err
				}
			}
		}(ns)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		fmt.Printf("error: %v\n", err)
	}

	return nil
}

func (p *PVCscaler) Down(ctx context.Context, outputFile string) error {
	err := p.getWorkloads(ctx, p.k8sClient, p.namespaces, p.storageClass)
	if err != nil {
		return err
	}

	for _, workload := range p.workloads {
		// TODO: p.k8sClient.ClientSet ugly
		err := workload.ScaleDown(ctx, p.k8sClient.ClientSet, workload.Namespace, workload.Name, workload.Kind)
		if err != nil {
			return err
		}
	}

	dataset := getDataset(p.workloads)

	if outputFile != "" {
		err = dataset.WritetToFile(outputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PVCscaler) Up(ctx context.Context, outputFile string) error {
	var dataset dataset

	err := dataset.ReadFromFile(outputFile)
	if err != nil {
		return err
	}

	workloads := dataset.toWorkloads()
	for _, workload := range workloads {
		// FIXME: workloads.Replicas
		err := workload.ScaleUp(ctx, p.k8sClient.ClientSet, workload.Namespace, workload.Name, workload.Kind, int32(workload.Replicas))
		if err != nil {
			return err
		}
	}

	return nil
}
