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
	expected := createDataset()
	dataset := dataset{}
	err := dataset.ReadFromFile("testdata/pvcscaler.json")
	assert.NoError(t, err)
	assert.Equal(t, expected, dataset)
}

func TestWriteToFile(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		dataset            dataset
		expectedOutputFile string
		actualOutputFile   string
		error              error
	}{
		{
			name:               "empty element",
			dataset:            dataset{},
			expectedOutputFile: "testdata/empty.json",
			actualOutputFile:   filepath.Join(tmpDir, "pvscaler.json"),
			error:              nil,
		},
		{
			name:               "one element",
			dataset:            createDataset(),
			expectedOutputFile: "testdata/pvcscaler.json",
			actualOutputFile:   filepath.Join(tmpDir, "pvscaler.json"),
			error:              nil,
		},
		{
			name:               "write error",
			dataset:            dataset{},
			expectedOutputFile: "invalid.json",
			actualOutputFile:   "/",
			error:              fmt.Errorf("error writing to file: open /: is a directory"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dataset.WriteToFile(tt.actualOutputFile)
			assert.Equal(t, tt.error, err)

			if tt.error != nil {
				return
			}

			expected, err := os.ReadFile(tt.expectedOutputFile)
			assert.NoError(t, err)

			actual, err := os.ReadFile(tt.expectedOutputFile)
			assert.NoError(t, err)
			assert.Equal(t, string(expected), string(actual))
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
