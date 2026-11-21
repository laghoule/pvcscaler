package k8s

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	isDefaultStorageClassAnnotation = "storageclass.kubernetes.io/is-default-class"
)

func (c *Client) isStorageClassMatched(pvcName, namespace, storageClass string) (bool, error) {
	pvc, err := c.ClientSet.CoreV1().PersistentVolumeClaims(namespace).Get(c.Context, pvcName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == storageClass, nil
}

func (c *Client) getDefaultStorageClass() (string, error) {
	c.defaultStorageClassOnce.Do(func() {
		storageClasses, err := c.ClientSet.StorageV1().StorageClasses().List(c.Context, metav1.ListOptions{})
		if err != nil {
			c.defaultStorageClassErr = fmt.Errorf("failed to list storage classes: %v", err)
			return
		}

		for _, sc := range storageClasses.Items {
			if sc.Annotations[isDefaultStorageClassAnnotation] == "true" {
				c.defaultStorageClass = sc.Name
				return
			}
		}
	})

	return c.defaultStorageClass, c.defaultStorageClassErr
}

func isVolumeClaimTemplateStorageClassMatched(storageClassName *string, defaultStorageClass, storageClass string) bool {
	if storageClassName != nil {
		return *storageClassName == storageClass
	}

	return defaultStorageClass == storageClass
}
