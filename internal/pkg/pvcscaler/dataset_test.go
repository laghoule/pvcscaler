package pvcscaler

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"laghoule/pvcscaler/internal/pkg/k8s"

	"github.com/stretchr/testify/assert"
)

func createDataset() dataset {
	return dataset{
		Workloads: createWorkloads(),
	}
}

func createWorkloads() []k8s.Workload {
	return []k8s.Workload{
		{
			Kind:      "Deployment",
			Name:      "prometheus-grafana",
			Namespace: "monitoring",
			Replicas:  1,
		},
	}
}

func TestReadFromFile(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		error    error
	}{
		{
			name:     "read file",
			filePath: "testdata/dataset.json",
			error:    nil,
		},
		{
			name:     "read file error",
			filePath: "notfound.json",
			error:    fmt.Errorf("open notfound.json: no such file or directory"),
		},
		{
			name:     "invalid json file",
			filePath: "testdata/invalid.json",
			error:    fmt.Errorf("invalid character 'i' looking for beginning of object key string"),
		},
	}

	expected := createDataset()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataset := dataset{}
			err := dataset.ReadFromFile(tt.filePath)
			if tt.error != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.error.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected, dataset)
			}
		})
	}
}

func TestWriteToFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		dataset            dataset
		expectedOutputFile string
		actualOutputFile   string
		error              bool
	}{
		{
			name:               "empty element",
			dataset:            dataset{},
			expectedOutputFile: "testdata/empty.json",
			actualOutputFile:   filepath.Join(tmpDir, "pvscaler.json"),
			error:              false,
		},
		{
			name:               "one element",
			dataset:            createDataset(),
			expectedOutputFile: "testdata/dataset.json",
			actualOutputFile:   filepath.Join(tmpDir, "pvscaler.json"),
			error:              false,
		},
		{
			name:               "write error",
			dataset:            dataset{},
			expectedOutputFile: "invalid.json",
			actualOutputFile:   "/",
			error:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dataset.WriteToFile(tt.actualOutputFile)

			if tt.error {
				assert.Error(t, err)
			} else {
				expected, err := os.ReadFile(tt.expectedOutputFile)
				assert.NoError(t, err)

				actual, err := os.ReadFile(tt.expectedOutputFile)
				assert.NoError(t, err)
				assert.Equal(t, string(expected), string(actual))
			}
		})
	}
}

func TestGetDataset(t *testing.T) {
	expected := createDataset()
	workloads := createWorkloads()
	actual := getDataset(workloads)
	assert.Equal(t, expected, actual)
}

func TestToWorkloads(t *testing.T) {
	expected := createWorkloads()
	dataset := createDataset()
	actual := dataset.toWorkloads()
	assert.Equal(t, expected, actual)
}
