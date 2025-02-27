package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createDeploymentWorkload(c kubernetes.Interface) *Workload {
	createDeployment(c)
	return &Workload{
		Name:      "nginx-deployment",
		Namespace: namespace,
		Kind:      "Deployment",
		Replicas:  1,
	}
}

func createStatefulsetWorkload(c kubernetes.Interface) *Workload {
	createStatefulSet(c)
	return &Workload{
		Name:      "nginx-statefulset",
		Namespace: namespace,
		Kind:      "StatefulSet",
		Replicas:  1,
	}
}

func TestGetWorkloads(t *testing.T) {
	c, err := NewTestClient()
	if err != nil {
		t.Error(err)
	}

	createDeployment(c.ClientSet)
	createReplicaSet(c.ClientSet)
	createPodWithPVC(c.ClientSet)
	createPVC(c.ClientSet, "pvc-0c2e9fda-e8e6-11e8-8c05-000c29c3a172", namespace, "standard")

	wloads, err := c.GetWorkloads(context.TODO(), namespace, "standard")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(wloads))
	assert.Equal(t, "nginx-deployment", wloads[0].Name)
	assert.Equal(t, namespace, wloads[0].Namespace)
}

func TestGetReplicas(t *testing.T) {
	type test[T any] struct {
		name     string
		workload T
		expected uint
		error    bool
	}

	c, err := NewTestClient()
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
			c, err := NewTestClient()
			assert.NoError(t, err)

			_, err = c.ClientSet.AppsV1().Deployments(namespace).Create(context.TODO(), tt.workload, metav1.CreateOptions{})
			assert.NoError(t, err)

			actual, err := c.getReplicas(context.TODO(), tt.workload.Namespace, tt.workload.Name, "Deployment")
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
			actual, err := c.getReplicas(context.TODO(), tt.workload.Namespace, tt.workload.Name, "StatefulSet")
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}

	t.Run(testUnsupportedKind.name, func(t *testing.T) {
		c, err := NewTestClient()
		assert.NoError(t, err)

		_, err = c.getReplicas(context.TODO(), testUnsupportedKind.workload.Namespace, testUnsupportedKind.workload.Name, "DaemonSet")
		assert.Error(t, err)
	})
}

func TestGetWorkloadOwnerKind(t *testing.T) {
	type test[T any] struct {
		name     string
		workload T
		expected string
		error    bool
	}

	c, err := NewTestClient()
	assert.NoError(t, err)

	testsDep := []test[*appsv1.ReplicaSet]{
		{
			name:     "deployment",
			workload: createReplicaSet(c.ClientSet),
			expected: "Deployment",
			error:    false,
		},
		{
			name:     "deployment error",
			workload: &appsv1.ReplicaSet{},
			expected: "",
			error:    true,
		},
	}

	testsSts := []test[*appsv1.StatefulSet]{
		{
			name:     "statefulset",
			workload: createStatefulSet(c.ClientSet),
			expected: "StatefulSet",
			error:    false,
		},
		{
			name:     "statefulset error",
			workload: &appsv1.StatefulSet{},
			expected: "",
			error:    true,
		},
	}

	testUnsupportedKind := test[*appsv1.DaemonSet]{
		name:     "unknown kind",
		workload: createDaemonSet(c.ClientSet),
		expected: "",
		error:    true,
	}

	for _, tt := range testsDep {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewTestClient()
			assert.NoError(t, err)

			pod := createPodWithPVC(c.ClientSet)
			actual, err := c.getWorkloadOwnerKind(context.TODO(), tt.workload.Namespace, pod.Name)
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
			c, err := NewTestClient()
			assert.NoError(t, err)

			pod := createStatefulSetPod(c.ClientSet)
			actual, err := c.getWorkloadOwnerKind(context.TODO(), tt.workload.Namespace, pod.Name)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}

	t.Run(testUnsupportedKind.name, func(t *testing.T) {
		c, err := NewTestClient()
		assert.NoError(t, err)

		pod := createDaemonSetPod(c.ClientSet)
		_, err = c.getWorkloadOwnerKind(context.TODO(), testUnsupportedKind.workload.Namespace, pod.Name)
		assert.Error(t, err)
	})
}

func TestScaleDown(t *testing.T) {
	c, err := NewTestClient()
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
			err := tt.workload.ScaleDown(context.TODO(), c, namespace, tt.workload.Name, tt.workload.Kind)
			assert.NoError(t, err)

			actual, err := c.ClientSet.AppsV1().Deployments(namespace).GetScale(context.TODO(), tt.workload.Name, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual.Spec.Replicas)

		})
	}
}

func TestScaleUp(t *testing.T) {
	c, err := NewTestClient()
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
			err := tt.workload.ScaleUp(context.TODO(), c, namespace, tt.workload.Name, tt.workload.Kind, tt.expected)
			assert.NoError(t, err)

			actual, err := c.ClientSet.AppsV1().Deployments(namespace).GetScale(context.TODO(), tt.workload.Name, metav1.GetOptions{})
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual.Spec.Replicas)

		})
	}
}

func TestScale(t *testing.T) {
	c, err := NewTestClient()
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
			err := tt.workload.scale(context.TODO(), c, namespace, tt.workload.Name, tt.workload.Kind, 0)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
