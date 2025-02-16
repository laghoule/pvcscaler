package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetWorkloads(t *testing.T) {
	c, err := NewTestClient()
	if err != nil {
		t.Error(err)
	}

	createDeployment(c.ClientSet)
	createReplicaSet(c.ClientSet)
	createPodWithPVC(c.ClientSet)

	pvc := createTestPVC("pvc-0c2e9fda-e8e6-11e8-8c05-000c29c3a172", namespace, "standard")
	_, err = c.ClientSet.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), pvc, metav1.CreateOptions{})
	assert.NoError(t, err)

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

			pod := createStatefulSetPod()
			_, err = c.ClientSet.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
			assert.NoError(t, err)

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

// func TestScale(t *testing.T) {
// 	c, err := NewTestClient()
// 	assert.NoError(t, err)

// 	dep := createDeployment()
// 	_, err = c.ClientSet.AppsV1().Deployments(namespace).Create(context.TODO(), dep, metav1.CreateOptions{})
// 	assert.NoError(t, err)

// 	err = c.scale(context.TODO(), dep.Namespace, dep.Name, "Deployment", 2)
// 	assert.NoError(t, err)

// FIXME
// _, err = c.ClientSet.AppsV1().Deployments(namespace).Get(context.TODO(), dep.Name, metav1.GetOptions{})
// assert.NoError(t, err)
// assert.Equal(t, int32(2), *dep.Spec.Replicas)
//}
