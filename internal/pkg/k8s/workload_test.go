package k8s

import (
	"testing"

	"laghoule/pvcscaler/internal/pkg/test"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

func createDeploymentWorkload(c kubernetes.Interface) *Workload {
	test.CreateDeployment(c)
	return &Workload{
		Name:      "nginx-deployment",
		Namespace: namespace,
		Kind:      "Deployment",
		Replicas:  1,
	}
}

func createStatefulsetWorkload(c kubernetes.Interface) *Workload {
	test.CreateStatefulSet(c)
	return &Workload{
		Name:      "nginx-statefulset",
		Namespace: namespace,
		Kind:      "StatefulSet",
		Replicas:  1,
	}
}

func TestGetWorkloads(t *testing.T) {
	c, err := NewTestClient(t)
	if err != nil {
		t.Error(err)
	}

	test.CreateDeployment(c.ClientSet)
	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	wloads, err := c.GetWorkloads(namespace, "standard")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(wloads))
	assert.Equal(t, "nginx-deployment", wloads[0].Name)
	assert.Equal(t, namespace, wloads[0].Namespace)
}

func TestGetDeploymentWorkloads(t *testing.T) {
	c, err := NewTestClient(t)
	if err != nil {
		t.Error(err)
	}

	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	tests := []struct {
		name         string
		dep          *appsv1.Deployment
		storageClass string
		expected     []Workload
		error        bool
	}{
		{
			name:         "deployment with pvc",
			dep:          test.CreateDeployment(c.ClientSet),
			storageClass: "standard",
			expected: []Workload{
				{
					Name:      "nginx-deployment",
					Namespace: namespace,
					Kind:      "Deployment",
					Replicas:  1,
				},
			},
			error: false,
		},
		{
			name:         "deployment with pvc and wrong storage class",
			dep:          test.CreateDeployment(c.ClientSet),
			storageClass: "not found",
			expected:     []Workload{},
			error:        false,
		},
		{
			name: "deployment with no pvc",
			dep: &appsv1.Deployment{
				Spec: appsv1.DeploymentSpec{
					Template: corev1.PodTemplateSpec{},
				},
			},
			expected: []Workload{},
			error:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := c.GetDeploymentWorkloads(tt.dep.Namespace, tt.storageClass)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestGetStatefulSetWorkloads(t *testing.T) {
	c, err := NewTestClient(t)
	if err != nil {
		t.Error(err)
	}

	test.CreateStatefulSet(c.ClientSet)
	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	wloads, err := c.GetStatefulSetWorkloads(namespace, "standard")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(wloads))
	assert.Equal(t, "nginx-statefulset", wloads[0].Name)
	assert.Equal(t, namespace, wloads[0].Namespace)
}
