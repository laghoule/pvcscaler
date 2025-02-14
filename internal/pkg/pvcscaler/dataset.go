package pvcscaler

import (
	"encoding/json"
	"fmt"
	"os"

	"laghoule/pvcscaler/internal/pkg/k8s"
)

type dataset struct {
	Workloads []k8s.Workload `json:"workloads"`
}

func getDataset(worloads []k8s.Workload) dataset {
	var dataset dataset
	dataset.Workloads = append(dataset.Workloads, worloads...)

	return dataset
}

// TODO: weird name
func (d *dataset) toWorkloads() []k8s.Workload {
	return d.Workloads
}

func (d *dataset) WriteToFile(filename string) error {
	jsonData, err := json.MarshalIndent(d.Workloads, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	return nil
}

func (d *dataset) ReadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &d.Workloads)
	if err != nil {
		return err
	}

	return nil
}
