package k8s

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	deploymentKind  = "Deployment"
	statefulSetKind = "StatefulSet"
)

// Workload identifies a scalable Kubernetes workload that mounts at least one
// PVC of the targeted storage class.
type Workload struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int32  `json:"replicas"`
}

// GetDeploymentWorkloads returns the Deployments in namespace that mount an
// existing PVC backed by storageClass.
func (c *Client) GetDeploymentWorkloads(namespace, storageClass string) ([]Workload, error) {
	deployments, err := c.ClientSet.AppsV1().Deployments(namespace).List(c.Context, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	workloads := []Workload{}

	for _, dep := range deployments.Items {
		matched, err := c.volumesUseStorageClass(dep.Spec.Template.Spec.Volumes, namespace, storageClass)
		if err != nil {
			return nil, err
		}

		if !matched {
			continue
		}

		replicas, err := c.getReplicas(namespace, dep.Name, deploymentKind)
		if err != nil {
			return nil, err
		}

		workloads = append(workloads, Workload{
			Name:      dep.Name,
			Kind:      deploymentKind,
			Namespace: namespace,
			Replicas:  replicas,
		})
	}

	return workloads, nil
}

// GetStatefulSetWorkloads returns the StatefulSets in namespace that use
// storageClass, whether through an existing PVC volume or through a volume
// claim template (including templates relying on the default storage class).
func (c *Client) GetStatefulSetWorkloads(namespace, storageClass string) ([]Workload, error) {
	statefulSets, err := c.ClientSet.AppsV1().StatefulSets(namespace).List(c.Context, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulset: %v", err)
	}

	defaultStorageClass, err := c.getDefaultStorageClass()
	if err != nil {
		return nil, err
	}

	workloads := []Workload{}

	for _, sts := range statefulSets.Items {
		// A statefulset can mount an existing PVC through regular volumes.
		matched, err := c.volumesUseStorageClass(sts.Spec.Template.Spec.Volumes, namespace, storageClass)
		if err != nil {
			return nil, err
		}

		// It can also request new PVCs through volume claim templates.
		if !matched {
			for _, vct := range sts.Spec.VolumeClaimTemplates {
				if isVolumeClaimTemplateStorageClassMatched(vct.Spec.StorageClassName, defaultStorageClass, storageClass) {
					matched = true
					break
				}
			}
		}

		if !matched {
			continue
		}

		replicas, err := c.getReplicas(namespace, sts.Name, statefulSetKind)
		if err != nil {
			return nil, err
		}

		workloads = append(workloads, Workload{
			Name:      sts.Name,
			Kind:      statefulSetKind,
			Namespace: namespace,
			Replicas:  replicas,
		})
	}

	return workloads, nil
}

// GetWorkloads returns every Deployment and StatefulSet in namespace that uses
// storageClass.
func (c *Client) GetWorkloads(namespace, storageClass string) ([]Workload, error) {
	workloads := []Workload{}

	depWorkloads, err := c.GetDeploymentWorkloads(namespace, storageClass)
	if err != nil {
		return nil, err
	}

	workloads = append(workloads, depWorkloads...)

	stsWorkloads, err := c.GetStatefulSetWorkloads(namespace, storageClass)
	if err != nil {
		return nil, err
	}

	workloads = append(workloads, stsWorkloads...)

	return workloads, nil
}

// volumesUseStorageClass reports whether any of the given volumes references an
// existing PVC backed by the requested storage class.
func (c *Client) volumesUseStorageClass(volumes []corev1.Volume, namespace, storageClass string) (bool, error) {
	for _, vol := range volumes {
		if vol.PersistentVolumeClaim == nil {
			continue
		}

		matched, err := c.isStorageClassMatched(vol.PersistentVolumeClaim.ClaimName, namespace, storageClass)
		if err != nil {
			return false, err
		}

		if matched {
			return true, nil
		}
	}

	return false, nil
}
