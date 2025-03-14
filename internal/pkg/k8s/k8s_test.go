package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubernetes "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

const (
	namespace = "default"
)

func NewFakeClient(t *testing.T) (*Client, error) {
	return &Client{
		ClientSet: fake.NewSimpleClientset(),
		Context: t.Context(),
	}, nil
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		kubeconfig string
		error bool
	}{
		{
			name: "not found kubeconfig",
			kubeconfig: "kubeconfig",
			error: true,
		},
		{
			name: "default kubeconfig",
			kubeconfig: "testdata/kubeconfig.yaml",
			error: false,
		},
	
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(t.Context(), tt.kubeconfig, false)
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

func createNamespace(c kubernetes.Interface) *corev1.Namespace  {
	nsObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
	}
	ns, _ := c.CoreV1().Namespaces().Create(context.TODO(), nsObj, metav1.CreateOptions{})
	return ns
}

func TestGetAllNamespaces(t *testing.T) {
	c, err := NewFakeClient(t)
	assert.NoError(t, err)

	createNamespace(c.ClientSet)

	tests := []struct {
		name    string
		expected corev1.Namespace
		error bool
	}{
		{
			name: "get all namespaces",
			expected: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			},
			error: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := c.GetAllNamespaces()
			if tt.error {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.expected.Name, actual[0])
			}
		})
	}

	namespaces, err := c.GetAllNamespaces()
	assert.NoError(t, err)
	assert.NotNil(t, namespaces)
	assert.Equal(t, "default", namespaces[0])
}

func TestClientGetDryRunUpdateOptionMetaV1(t *testing.T) {
	tests := []struct {
		name   string
		dryRun bool
		want   metav1.UpdateOptions
	}{
		{
			name:   "with dry run enabled",
			dryRun: true,
			want: metav1.UpdateOptions{
				DryRun: []string{metav1.DryRunAll},
			},
		},
		{
			name:   "with dry run disabled",
			dryRun: false,
			want:   metav1.UpdateOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				DryRun: tt.dryRun,
			}
			got := c.getDryRunUpdateOptionMetaV1()
			assert.Equal(t, got, tt.want)
		})
	}
}

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
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "nginx-pvc",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nginx-pvc",
								},
							},
						},
					},
				},
			},
			Replicas: int32PTR(1),
		},
	}
	dep, _ := c.AppsV1().Deployments(namespace).Create(context.TODO(), depObj, metav1.CreateOptions{})
	return dep
}

func createStatefulSet(c kubernetes.Interface) *appsv1.StatefulSet {
	scName := "standard"
	stsObj := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-statefulset",
			Namespace: namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: int32PTR(1),
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "nginx-pvc",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						VolumeName: "nginx-pvc",
						StorageClassName: &scName,
					},
				},
			},
		},
	}
	sts, _ := c.AppsV1().StatefulSets(namespace).Create(context.TODO(), stsObj, metav1.CreateOptions{})
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
	ds, _ := c.AppsV1().DaemonSets(namespace).Create(context.TODO(), dsObj, metav1.CreateOptions{})
	return ds
}
