package k8s

import (
	"testing"

	assert "github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetReplicas(t *testing.T) {
	type test[T any] struct {
		name     string
		workload T
		expected int32
		error    bool
	}

	c, err := NewFakeClient(t)
	assert.NoError(t, err)

	testsDep := []test[*appsv1.Deployment]{
		{
			name:     "deployment",
			workload: createDeployment(c.ClientSet),
			expected: 1,
			error:    false,
		},
		{
			name:     "deployment error",
			workload: &appsv1.Deployment{},
			expected: 0,
			error:    true,
		},
	}

	testsSts := []test[*appsv1.StatefulSet]{
		{
			name:     "statefulset",
			workload: createStatefulSet(c.ClientSet),
			expected: 1,
			error:    false,
		},
		{
			name:     "statefulset error",
			workload: &appsv1.StatefulSet{},
			expected: 0,
			error:    true,
		},
	}

	testUnsupportedKind := test[*appsv1.DaemonSet]{
		name:     "unknown kind",
		workload: createDaemonSet(c.ClientSet),
		expected: 0,
		error:    true,
	}

	for _, tt := range testsDep {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewFakeClient(t)
			assert.NoError(t, err)

			_, err = c.ClientSet.AppsV1().Deployments(namespace).Create(t.Context(), tt.workload, metav1.CreateOptions{})
			assert.NoError(t, err)

			actual, err := c.getReplicas(tt.workload.Namespace, tt.workload.Name, "Deployment")
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}

	for _, tt := range testsSts {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := c.getReplicas(tt.workload.Namespace, tt.workload.Name, "StatefulSet")
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}

	t.Run(testUnsupportedKind.name, func(t *testing.T) {
		c, err := NewFakeClient(t)
		assert.NoError(t, err)

		_, err = c.getReplicas(testUnsupportedKind.workload.Namespace, testUnsupportedKind.workload.Name, "DaemonSet")
		assert.Error(t, err)
	})
}

func TestScaleDown(t *testing.T) {
	c, err := NewFakeClient(t)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		workload *Workload
		expected int32
	}{
		{
			name:     "deployment",
			workload: createDeploymentWorkload(c.ClientSet),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workload.ScaleDown(c, namespace, tt.workload.Name, tt.workload.Kind)
			assert.NoError(t, err)

			actual, err := c.ClientSet.AppsV1().Deployments(namespace).GetScale(t.Context(), tt.workload.Name, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual.Spec.Replicas)

		})
	}
}

func TestScaleUp(t *testing.T) {
	c, err := NewFakeClient(t)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		workload *Workload
		expected int32
	}{
		{
			name:     "deployment",
			workload: createDeploymentWorkload(c.ClientSet),
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workload.ScaleUp(c, namespace, tt.workload.Name, tt.workload.Kind, tt.expected)
			assert.NoError(t, err)

			actual, err := c.ClientSet.AppsV1().Deployments(namespace).GetScale(t.Context(), tt.workload.Name, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual.Spec.Replicas)
		})
	}
}

func TestScale(t *testing.T) {
	c, err := NewFakeClient(t)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		workload *Workload
		expected uint
		error    bool
	}{
		{
			name:     "deployment",
			workload: createDeploymentWorkload(c.ClientSet),
			expected: 0,
			error:    false,
		},
		{
			name:     "statefulset",
			workload: createStatefulsetWorkload(c.ClientSet),
			expected: 0,
			error:    false,
		},
		{
			name:     "workload error",
			workload: &Workload{},
			expected: 0,
			error:    true,
		},
		{
			name: "unsupported kind error",
			workload: &Workload{
				Name:      "nginx",
				Namespace: namespace,
				Kind:      "DaemonSet",
				Replicas:  1,
			},
			expected: 0,
			error:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.workload.scale(c, namespace, tt.workload.Name, tt.workload.Kind, 0)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
