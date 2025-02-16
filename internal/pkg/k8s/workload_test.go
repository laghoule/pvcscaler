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

	dep := createDeployment()
	_, err = c.ClientSet.AppsV1().Deployments(namespace).Create(context.TODO(), dep, metav1.CreateOptions{})
	assert.NoError(t, err)

	rs := createTestReplicaSet()
	_, err = c.ClientSet.AppsV1().ReplicaSets(namespace).Create(context.TODO(), rs, metav1.CreateOptions{})
	assert.NoError(t, err)

	pod := createTestPodWithPVC()
	_, err = c.ClientSet.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

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
		replicas uint
		error    bool
	}

	testsDep := []test[*appsv1.Deployment]{
		{
			name:     "deployment",
			workload: createDeployment(),
			replicas: 1,
			error:    false,
		},
		{
			name:     "deployment error",
			workload: &appsv1.Deployment{},
			replicas: 0,
			error:    true,
		},
	}

	testsSts := []test[*appsv1.StatefulSet]{
		{
			name:     "statefulset",
			workload: createStatefulSet(),
			replicas: 1,
			error:    false,
		},
		{
			name:     "statefulset error",
			workload: &appsv1.StatefulSet{},
			replicas: 0,
			error:    true,
		},
	}

	testUnsupportedKind := test[*appsv1.DaemonSet]{
		name:     "unknown kind",
		workload: createDaemonSet(),
		replicas: 0,
		error:    true,
	}

	for _, tt := range testsDep {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewTestClient()
			assert.NoError(t, err)

			_, err = c.ClientSet.AppsV1().Deployments(namespace).Create(context.TODO(), tt.workload, metav1.CreateOptions{})
			assert.NoError(t, err)

			replicas, err := c.getReplicas(context.TODO(), tt.workload.Namespace, tt.workload.Name, "Deployment")
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.replicas, replicas)
			}
		})
	}

	for _, tt := range testsSts {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewTestClient()
			assert.NoError(t, err)

			_, err = c.ClientSet.AppsV1().StatefulSets(namespace).Create(context.TODO(), tt.workload, metav1.CreateOptions{})
			assert.NoError(t, err)

			replicas, err := c.getReplicas(context.TODO(), tt.workload.Namespace, tt.workload.Name, "StatefulSet")
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.replicas, replicas)
			}
		})
	}

	t.Run(testUnsupportedKind.name, func(t *testing.T) {
		c, err := NewTestClient()
		assert.NoError(t, err)

		_, err = c.ClientSet.AppsV1().DaemonSets(namespace).Create(context.TODO(), testUnsupportedKind.workload, metav1.CreateOptions{})
		assert.NoError(t, err)

		_, err = c.getReplicas(context.TODO(), testUnsupportedKind.workload.Namespace, testUnsupportedKind.workload.Name, "DaemonSet")
		assert.Error(t, err)
	})
}

func TestGetWorkloadOwnerType(t *testing.T) {
	c, err := NewTestClient()
	assert.NoError(t, err)

	rs := createTestReplicaSet()
	_, err = c.ClientSet.AppsV1().ReplicaSets(namespace).Create(context.TODO(), rs, metav1.CreateOptions{})
	assert.NoError(t, err)

	pod := createTestPodWithPVC()
	_, err = c.ClientSet.CoreV1().Pods(namespace).Create(context.TODO(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	ownerType, err := c.getWorkloadOwnerKind(context.TODO(), pod.Namespace, pod.Name)
	assert.NoError(t, err)
	assert.Equal(t, "Deployment", ownerType)
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
