package pvcscaler

import (
	"context"
	"fmt"
	"sync"

	"github.com/laghoule/pvcscaler/internal/pkg/k8s"

	"github.com/fatih/color"
	"github.com/rodaine/table"
)

type PVCscaler struct {
	k8sClient *k8s.Client
	context   context.Context
	dryRun    bool

	workloads    []k8s.Workload
	namespaces   []string
	storageClass string
}

func New(ctx context.Context, kubeconfig string, namespaces []string, storageClass string, dryRun bool) (*PVCscaler, error) {
	k8sClient, err := k8s.New(ctx, kubeconfig, dryRun)
	if err != nil {
		return nil, err
	}

	return &PVCscaler{
		k8sClient:    k8sClient,
		context:      ctx,
		dryRun:       dryRun,
		namespaces:   namespaces,
		storageClass: storageClass,
	}, nil
}

func (p *PVCscaler) getWorkloads(namespaces []string, storageClass string) error {
	var err error

	if len(namespaces) == 1 && namespaces[0] == "all" {
		namespaces, err = p.k8sClient.GetAllNamespaces()
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	var errChan = make(chan error, len(namespaces))

	for _, ns := range namespaces {
		wg.Add(1)
		go func(ns string) {
			defer wg.Done()
			select {
			case <-p.context.Done():
				return
			default:
				workloads, err := p.k8sClient.GetWorkloads(ns, storageClass)
				if err != nil {
					errChan <- err
					return
				}
				p.workloads = append(p.workloads, workloads...)
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

func (p *PVCscaler) Down(outputFile string) error {
	err := p.getWorkloads(p.namespaces, p.storageClass)
	if err != nil {
		return err
	}

	fmt.Println("Scaling down these workloads:")

	for _, workload := range p.workloads {
		fmt.Printf(" ✴️ %s %s/%s\n", workload.Kind, workload.Namespace, workload.Name)
		err := workload.ScaleDown(p.k8sClient, workload.Namespace, workload.Name, workload.Kind)
		if err != nil {
			return err
		}
	}

	dataset := newDataset(p.workloads)

	if outputFile != "" {
		err = dataset.writeToFile(outputFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PVCscaler) Up(outputFile string) error {
	var dataset dataset

	err := dataset.readFromFile(outputFile)
	if err != nil {
		return err
	}

	fmt.Println("Scaling up these workloads:")

	for _, workload := range dataset.Workloads {
		fmt.Printf(" ✴️ %s %s/%s\n", workload.Kind, workload.Namespace, workload.Name)
		err := workload.ScaleUp(p.k8sClient, workload.Namespace, workload.Name, workload.Kind, int32(workload.Replicas))
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *PVCscaler) PrintList() error {
	err := p.getWorkloads(p.namespaces, p.storageClass)
	if err != nil {
		return err
	}

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	tbl := table.New("Namespace", "Name", "Type", "Replicas", "StorageClass")
	tbl.WithHeaderFormatter(headerFmt)

	for _, workload := range p.workloads {
		tbl.AddRow(workload.Namespace, workload.Name, workload.Kind, workload.Replicas, p.storageClass)
	}

	tbl.Print()
	return nil
}
