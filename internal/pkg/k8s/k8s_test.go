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

func TestNewClient(t *testing.T) {
	client, err := New("")
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

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
