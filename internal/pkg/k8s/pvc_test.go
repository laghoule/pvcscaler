package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createTestPVC(name, namespace, storageClass string) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClass,
		},
	}
}

func TestIsStorageClassMatched(t *testing.T) {
	client, err := NewTestClient()
	assert.NoError(t, err)

	testPVC := createTestPVC("test-pvc", "default", "standard")
	_, err = client.ClientSet.CoreV1().PersistentVolumeClaims("default").Create(context.Background(), testPVC, metav1.CreateOptions{})
	assert.NoError(t, err)

	tests := []struct {
		name         string
		pvcName      string
		namespace    string
		storageClass string
		want         bool
		wantErr      bool
	}{
		{
			name:         "storage class matches",
			pvcName:      "test-pvc",
			namespace:    "default",
			storageClass: "standard",
			want:         true,
			wantErr:      false,
		},
		{
			name:         "storage class doesn't match",
			pvcName:      "test-pvc",
			namespace:    "default",
			storageClass: "premium",
			want:         false,
			wantErr:      false,
		},
		{
			name:         "pvc not found",
			pvcName:      "nonexistent-pvc",
			namespace:    "default",
			storageClass: "standard",
			want:         false,
			wantErr:      false,
		},
		// {
		// 	name:         "error getesting pvc",
		// 	pvcName:      "error-pvc",
		// 	namespace:    "nonexistant .namespace",
		// 	storageClass: "standard",
		// 	want:         false,
		// 	wantErr:      true,
		// },
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := client.isStorageClassMatched(context.TODO(), test.pvcName, test.namespace, test.storageClass)
			if test.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}
