package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespace = "default"
)

func int32PTR(n int) *int32 {
	i := int32(n)
	return &i
}

func createDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32PTR(1),
		},
	}
}

// FIXME: change name
func createTestReplicaSet() *appsv1.ReplicaSet {
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment-7d475f5dd6",
			Namespace: namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "Deployment",
					Name: "nginx-deployment",
				},
			},
		},
		Spec: appsv1.ReplicaSetSpec{
			Replicas: int32PTR(1),
		},
	}
}

func createStatefulSet() *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-statefulset",
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32PTR(1),
		},
	}
}

func createDaemonSet() *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-daemonset",
			Namespace: namespace,
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
					},
				},
			},
		},
	}
}

// FIXME: change name
func createTestPodWithPVC() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment-7d475f5dd6-7x4xw",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "ReplicaSet",
					Name: "nginx-deployment-7d475f5dd6",
				},
			},
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{
				{
					Name: "pvc-0c2e9fda-e8e6-11e8-8c05-000c29c3a172",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: "pvc-0c2e9fda-e8e6-11e8-8c05-000c29c3a172",
						},
					},
				},
			},
		},
	}
}

func createStatefulSetPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-statefulset-0",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "StatefulSet",
					Name: "nginx-statefulset",
				},
			},
		},
	}
}

func createDaemonSetPod() *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-daemonset-7d475f5dd6",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind: "DaemonSet",
					Name: "nginx-daemonset",
				},
			},
		},
	}
}

func TestGetPodOwnerKind(t *testing.T) {
	c, err := NewTestClient()
	if err != nil {
		t.Error(err)
	}

	pod := createTestPodWithPVC()
	_, err = c.ClientSet.CoreV1().Pods(namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	// TODO: Add statefulset

	kind, err := c.getPodOwnerKind(context.TODO(), pod.Namespace, pod.Name)
	assert.NoError(t, err)
	assert.Equal(t, "ReplicaSet", kind)
}
