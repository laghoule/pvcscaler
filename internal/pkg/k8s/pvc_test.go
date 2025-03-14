package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func createPVC(c kubernetes.Interface, name, namespace, storageClass string) *corev1.PersistentVolumeClaim {
	pvcObj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
		},
	}
	pvc, _ := c.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), pvcObj, metav1.CreateOptions{})
	return pvc
}

func TestIsStorageClassMatched(t *testing.T) {
	c, err := NewFakeClient(t)
	assert.NoError(t, err)

	createPVC(c.ClientSet, "test-pvc", "default", "standard")

	tests := []struct {
		name         string
		pvcName      string
		namespace    string
		storageClass string
		expected     bool
		error        bool
	}{
		{
			name:         "storage class matches",
			pvcName:      "test-pvc",
			namespace:    "default",
			storageClass: "standard",
			expected:     true,
			error:        false,
		},
		{
			name:         "storage class doesn't match",
			pvcName:      "test-pvc",
			namespace:    "default",
			storageClass: "premium",
			expected:     false,
			error:        false,
		},
		{
			name:         "pvc not found",
			pvcName:      "nonexistent-pvc",
			namespace:    "default",
			storageClass: "standard",
			expected:     false,
			error:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := c.isStorageClassMatched(tt.pvcName, tt.namespace, tt.storageClass)
			if tt.error {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
