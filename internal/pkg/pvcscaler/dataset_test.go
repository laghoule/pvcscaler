package pvcscaler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"laghoule/pvcscaler/internal/pkg/k8s"
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

func TestReadWorkloadsFromFile(t *testing.T) {
	expected := createDataset()
	dataset := dataset{}
	err := dataset.ReadFromFile("testdata/pvcscaler.json")
	assert.NoError(t, err)
	assert.Equal(t, expected, dataset)
}
