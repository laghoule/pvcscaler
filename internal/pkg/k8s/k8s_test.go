package k8s

import (
	"context"
	"testing"

	"laghoule/pvcscaler/internal/pkg/test"

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

func NewTestClient(t *testing.T) (*Client, error) {
	return &Client{
		ClientSet: fake.NewSimpleClientset(),
		Context:   t.Context(),
	}, nil
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		kubeconfig string
		error      bool
	}{
		{
			name:       "not found kubeconfig",
			kubeconfig: "kubeconfig",
			error:      true,
		},
		{
			name:       "default kubeconfig",
			kubeconfig: "testdata/kubeconfig.yaml",
			error:      false,
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

func TestGetAllNamespaces(t *testing.T) {
	c, err := NewTestClient(t)
	assert.NoError(t, err)

	test.CreateNamespace(c.ClientSet)

	tests := []struct {
		name     string
		expected corev1.Namespace
		error    bool
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
