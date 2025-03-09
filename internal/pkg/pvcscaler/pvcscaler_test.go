package pvcscaler

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		kubeconfig   string
		namespace    []string
		storageClass string
		dryRun       bool
		error        bool
	}{
		{
			name:         "New PVCscaler",
			kubeconfig:   "testdata/kubeconfig.yaml",
			namespace:    []string{"default"},
			storageClass: "standard",
			dryRun:       false,
			error:        false,
		},

		{
			name:         "New PVCscaler with empty storageClass",
			kubeconfig:   "filenotfound",
			namespace:    []string{"default"},
			storageClass: "",
			dryRun:       false,
			error:        true,
		},
	}

	for _, tt := range tests {
		pvcscaler, err := New(tt.kubeconfig, tt.namespace, tt.storageClass, tt.dryRun)
		if tt.error {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, pvcscaler)
		}
	}
}
