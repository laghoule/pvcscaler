package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func NewTestClient() (*Client, error) {
	return &Client{
		ClientSet: fake.NewSimpleClientset(),
	}, nil
}

// FIXME: not working in github actions
// func TestNewClient(t *testing.T) {
// 	client, err := New("", false)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, client)
// }

func createTestNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "default",
		},
	}
}

func TestGetAllNamespaces(t *testing.T) {
	client, err := NewTestClient()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	ns := createTestNamespace()
	_, err = client.ClientSet.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	assert.NoError(t, err)

	namespaces, err := client.GetAllNamespaces(context.TODO())
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
