package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
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
	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	wloads, err := c.GetWorkloads(context.TODO(), namespace, "standard")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(wloads))
	assert.Equal(t, "nginx-deployment", wloads[0].Name)
	assert.Equal(t, namespace, wloads[0].Namespace)
}

func TestGetDeploymentWorkloads(t *testing.T) {
	c, err := NewTestClient()
	if err != nil {
		t.Error(err)
	}

	createDeployment(c.ClientSet)
	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	wloads, err := c.GetDeploymentWorkloads(context.TODO(), namespace, "standard")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(wloads))
	assert.Equal(t, "nginx-deployment", wloads[0].Name)
	assert.Equal(t, namespace, wloads[0].Namespace)
}

func TestGetStatefulSetWorkloads(t *testing.T) {
	c, err := NewTestClient()
	if err != nil {
		t.Error(err)
	}

	createStatefulSet(c.ClientSet)
	createPVC(c.ClientSet, "nginx-pvc", namespace, "standard")

	wloads, err := c.GetStatefulSetWorkloads(context.TODO(), namespace, "standard")
	assert.NoError(t, err)

	assert.Equal(t, 1, len(wloads))
	assert.Equal(t, "nginx-statefulset", wloads[0].Name)
	assert.Equal(t, namespace, wloads[0].Namespace)
}