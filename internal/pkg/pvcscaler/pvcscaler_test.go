package pvcscaler

import (
	"os"
	"path/filepath"
	"testing"

	"laghoule/pvcscaler/internal/pkg/k8s"
	"laghoule/pvcscaler/internal/pkg/test"

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
		pvcscaler, err := New(t.Context(), tt.kubeconfig, tt.namespace, tt.storageClass, tt.dryRun)
		if tt.error {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.NotNil(t, pvcscaler)
		}
	}
}

func TestGetWorkloads(t *testing.T) {
	fakeClient := test.NewFakeClient()
	test.CreateNamespace(fakeClient)
	test.CreatePVC(fakeClient, "nginx-pvc", test.Namespace, "standard")

	tests := []struct {
		name         string
		resource     test.K8sResource[any]
		namespace    []string
		storageClass string
		error        bool
	}{
		{
			name:         "GetWorkloads deployment for all namespaces",
			resource:     test.K8sResource[any]{Resource: test.CreateDeployment(fakeClient)},
			namespace:    []string{"all"},
			storageClass: "standard",
			error:        false,
		},
		{
			name:         "GetWorkloads statefulset for all namespaces",
			resource:     test.K8sResource[any]{Resource: test.CreateStatefulSet(fakeClient)},
			namespace:    []string{"all"},
			storageClass: "standard",
			error:        false,
		},
	}

	for _, tt := range tests {
		pvcscaler := &PVCscaler{
			k8sClient:    &k8s.Client{ClientSet: fakeClient, Context: t.Context()},
			context:      t.Context(),
			namespaces:   tt.namespace,
			storageClass: tt.storageClass,
		}

		t.Run(tt.name, func(t *testing.T) {
			err := pvcscaler.getWorkloads(tt.namespace, tt.storageClass)

			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, pvcscaler.workloads, 2)
			}
		})
	}
}

func TestDown(t *testing.T) {
	fakeClient := test.NewFakeClient()
	test.CreateNamespace(fakeClient)
	test.CreatePVC(fakeClient, "nginx-pvc", test.Namespace, "standard")
	test.CreateDeployment(fakeClient)

	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		resource           test.K8sResource[any]
		namespace          []string
		storageClass       string
		expectedOutputFile string
		actualOutputFile   string
		error              bool
	}{
		{
			name:               "Down deployment for all namespaces",
			resource:           test.K8sResource[any]{Resource: test.CreateDeployment(fakeClient)},
			namespace:          []string{"all"},
			storageClass:       "standard",
			expectedOutputFile: "testdata/pvcscaler.json",
			actualOutputFile:   filepath.Join(tmpDir, "pvscaler.json"),
			error:              false,
		},
	}

	for _, tt := range tests {
		pvcscaler := &PVCscaler{
			k8sClient:    &k8s.Client{ClientSet: fakeClient, Context: t.Context()},
			context:      t.Context(),
			namespaces:   tt.namespace,
			storageClass: tt.storageClass,
		}

		t.Run(tt.name, func(t *testing.T) {
			err := pvcscaler.Down(tt.actualOutputFile)

			if tt.error {
				assert.Error(t, err)
			} else {
				expected, err := os.ReadFile(tt.expectedOutputFile)
				assert.NoError(t, err)

				actual, err := os.ReadFile(tt.actualOutputFile)
				assert.NoError(t, err)
				assert.JSONEq(t, string(expected), string(actual))
			}
		})
	}
}

func TestUp(t *testing.T) {
	fakeClient := test.NewFakeClient()
	test.CreateNamespace(fakeClient)

	tests := []struct {
		name      string
		inputFile string
		resource  test.K8sResource[any]
		error     bool
	}{
		{
			name:      "Up deployment",
			inputFile: "testdata/pvcscaler.json",
			resource:  test.K8sResource[any]{Resource: test.CreateDeployment(fakeClient)},
			error:     false,
		},
		{
			name:      "No deployment found",
			inputFile: "testdata/noWorkloadFoundjson",
			resource:  test.K8sResource[any]{Resource: nil},
			error:     true,
		},
		{
			name:      "No input file found",
			inputFile: "testdata/notfound.json",
			error:     true,
		},
	}

	for _, tt := range tests {
		pvcscaler := &PVCscaler{
			k8sClient: &k8s.Client{ClientSet: fakeClient},
		}

		t.Run(tt.name, func(t *testing.T) {
			err := pvcscaler.Up(tt.inputFile)

			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestList(t *testing.T) {
	fakeClient := test.NewFakeClient()
	test.CreateNamespace(fakeClient)
	test.CreatePVC(fakeClient, "nginx-pvc", test.Namespace, "standard")
	test.CreateDeployment(fakeClient)

	tests := []struct {
		name      string
		namespace []string
		error     bool
	}{
		{
			name:      "List deployment for all namespaces",
			namespace: []string{"all"},
			error:     false,
		},
	}

	for _, tt := range tests {
		pvcscaler := &PVCscaler{
			k8sClient:  &k8s.Client{ClientSet: fakeClient, Context: t.Context()},
			context:    t.Context(),
			namespaces: tt.namespace,
		}

		t.Run(tt.name, func(t *testing.T) {
			err := pvcscaler.PrintList()

			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
