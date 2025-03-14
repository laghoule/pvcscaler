package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	deploymentKind  = "Deployment"
	statefulSetKind = "StatefulSet"
)

type Workload struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  int32  `json:"replicas"`
}

func (c *Client) GetDeploymentWorkloads(ctx context.Context, namespace, storageClass string) ([]Workload, error) {
	dep, err := c.ClientSet.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	workloads := []Workload{}

	for _, dep := range dep.Items {
		if dep.Spec.Template.Spec.Volumes == nil {
			continue
		}
		for _, vol := range dep.Spec.Template.Spec.Volumes {
			if vol.PersistentVolumeClaim == nil {
				continue
			}

			pvcMatchSC, err := c.isStorageClassMatched(vol.PersistentVolumeClaim.ClaimName, namespace, storageClass)
			if err != nil {
				return nil, err
			}

			if !pvcMatchSC {
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
	}

	return workloads, nil
}

func (c *Client) GetStatefulSetWorkloads(ctx context.Context, namespace, storageClass string) ([]Workload, error) {
	sts, err := c.ClientSet.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list statefulset: %v", err)
	}

	workloads := []Workload{}

	for _, sts := range sts.Items {
		if sts.Spec.VolumeClaimTemplates == nil {
			continue
		}

		for _, vct := range sts.Spec.VolumeClaimTemplates {
			if vct.Spec.StorageClassName == nil {
				continue
			}

			if *vct.Spec.StorageClassName != storageClass {
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
	}

	return workloads, nil
}

func (c *Client) GetWorkloads(ctx context.Context, namespace, storageClass string) ([]Workload, error) {
	workloads := []Workload{}

	depWorkloads, err := c.GetDeploymentWorkloads(ctx, namespace, storageClass)
	if err != nil {
		return nil, err
	}

	workloads = append(workloads, depWorkloads...)

	stsWorkloads, err := c.GetStatefulSetWorkloads(ctx, namespace, storageClass)
	if err != nil {
		return nil, err
	}

	workloads = append(workloads, stsWorkloads...)

	return workloads, nil
}
