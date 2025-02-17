package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	namespace = "default"
)

func int32PTR(n int) *int32 {
	i := int32(n)
	return &i
}

func createDeployment(c kubernetes.Interface) *appsv1.Deployment {
	depObj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment",
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32PTR(1),
		},
	}
	dep, _ := c.AppsV1().Deployments(namespace).Create(context.Background(), depObj, metav1.CreateOptions{})
	return dep
}

func createReplicaSet(c kubernetes.Interface) *appsv1.ReplicaSet {
	rsObj := &appsv1.ReplicaSet{
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
	rs, _ := c.AppsV1().ReplicaSets(namespace).Create(context.Background(), rsObj, metav1.CreateOptions{})
	return rs
}

func createStatefulSet(c kubernetes.Interface) *appsv1.StatefulSet {
	stsObj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-statefulset",
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32PTR(1),
		},
	}
	sts, _ := c.AppsV1().StatefulSets(namespace).Create(context.Background(), stsObj, metav1.CreateOptions{})
	return sts
}

func createDaemonSet(c kubernetes.Interface) *appsv1.DaemonSet {
	dsObj := &appsv1.DaemonSet{
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
	ds, _ := c.AppsV1().DaemonSets(namespace).Create(context.Background(), dsObj, metav1.CreateOptions{})
	return ds
}

func createPodWithPVC(c kubernetes.Interface) *corev1.Pod {
	podObj := &corev1.Pod{
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
	pod, _ := c.CoreV1().Pods(namespace).Create(context.Background(), podObj, metav1.CreateOptions{})
	return pod
}

func createStatefulSetPod(c kubernetes.Interface) *corev1.Pod {
	podObj := &corev1.Pod{
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
	pod, _ := c.CoreV1().Pods(namespace).Create(context.Background(), podObj, metav1.CreateOptions{})
	return pod
}

func createDaemonSetPod(c kubernetes.Interface) *corev1.Pod {
	podObj := &corev1.Pod{
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
	pod, _ := c.CoreV1().Pods(namespace).Create(context.Background(), podObj, metav1.CreateOptions{})
	return pod
}

func TestGetPodOwnerKind(t *testing.T) {
	c, err := NewTestClient()
	if err != nil {
		t.Error(err)
	}

	emptyPodObj := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment-7d475f5dd6-7x4xx",
			Namespace: "default",
		},
	}

	emptyPod, err := c.ClientSet.CoreV1().Pods(namespace).Create(context.Background(), emptyPodObj, metav1.CreateOptions{})
	assert.NoError(t, err)

	tests := []struct {
		name     string
		pod      *corev1.Pod
		expected string
		error    bool
	}{
		{
			name:     "Deployment",
			pod:      createPodWithPVC(c.ClientSet),
			expected: "ReplicaSet",
			error:    false,
		},
		{
			name:     "Empty deployment",
			pod:      &corev1.Pod{},
			expected: "",
			error:    true,
		},
		{
			name:     "StatefulSet",
			pod:      createStatefulSetPod(c.ClientSet),
			expected: "StatefulSet",
			error:    false,
		},
		{
			name:     "No owner",
			pod:      emptyPod,
			expected: "",
			error:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := c.getPodOwnerKind(context.TODO(), tt.pod.Namespace, tt.pod.Name)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}
