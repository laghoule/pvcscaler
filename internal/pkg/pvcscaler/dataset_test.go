package pvcscaler

import (
	"os"
	"path/filepath"
	"testing"

	"laghoule/pvcscaler/internal/pkg/k8s"

	"github.com/stretchr/testify/assert"
)

func createDataset() dataset {
	return dataset{
		Workloads: []k8s.Workload{
			{
				Kind:      "Deployment",
				Name:      "prometheus-grafana",
				Namespace: "monitoring",
				Replicas:  1,
			},
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
	dataset := dataset{}
	tmpDir := t.TempDir()

	err := dataset.WritetToFile(filepath.Join(tmpDir, "pvcscaler.json"))
	assert.NoError(t, err)

	expected, err := os.ReadFile("testdata/pvcscaler.json")
	assert.NoError(t, err)
	
	actual, err := os.ReadFile(filepath.Join(tmpDir, "pvcscaler.json"))
	assert.NoError(t, err)
	assert.Equal(t, string(expected), string(actual))
}
