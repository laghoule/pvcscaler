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

func newDataset(workloads []k8s.Workload) dataset {
	return dataset{
		Workloads: workloads,
	}
}

func (d *dataset) writeToFile(filename string) error {
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

func (d *dataset) readFromFile(filename string) error {
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
