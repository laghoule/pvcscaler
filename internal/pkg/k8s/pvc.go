package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) isStorageClassMatched(ctx context.Context, pvcName, namespace, storageClass string) (bool, error) {
	pvc, err := c.ClientSet.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == storageClass, nil
}
