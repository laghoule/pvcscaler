package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Workload struct {
	kind      string
	name      string
	namespace string
	replicas  uint
}

func (c *Client) GetWorkloads(ctx context.Context, namespace, storageClass string) ([]Workload, error) {
	pods, err := c.ClientSet.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	workloads := []Workload{}

	for _, pod := range pods.Items {
		if pod.Spec.Volumes == nil {
			continue
		}
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim == nil {
				continue
			}

			pvcMatchSC, err := c.isStorageClassMatched(ctx, vol.PersistentVolumeClaim.ClaimName, namespace, storageClass)
			if err != nil {
				return nil, err
			}

			if !pvcMatchSC {
				continue
			}

			kind, err := c.getWorkloadOwnerType(ctx, namespace, pod.Name)
			if err != nil {
				return nil, err
			}

			owner, err := c.getPodOwnerName(ctx, namespace, pod.Name)
			if err != nil {
				return nil, err
			}

			replicas, err := c.getReplicas(ctx, namespace, owner, kind)
			if err != nil {
				return nil, err
			}

			fmt.Printf("Pod %s (%s: %s) use a PVC of storage class %q\n", pod.Name, kind, owner, storageClass)
			workloads = append(workloads, Workload{
				kind:      kind,
				name:      owner,
				namespace: namespace,
				replicas:  replicas,
			})
		}
	}

	return workloads, nil
}

// TODO: maybe use k8s type, not string

func (c *Client) getWorkloadOwnerType(ctx context.Context, namespace, podName string) (string, error) {
	ownerKind, err := c.getPodOwnerKind(ctx, namespace, podName)
	if err != nil {
		return "", err
	}

	switch ownerKind {
	case "ReplicaSet":
		rsName, err := c.getReplicaSetFromPod(ctx, namespace, podName)
		if err != nil {
			return "", err
		}
		return c.getReplicaSetOwner(ctx, namespace, rsName)
	case "StatefulSet":
		return "StatefulSet", nil
	default:
		return "", fmt.Errorf("unsupported kind %q", ownerKind)
	}
}

// TODO: maybe use k8s type, not string

func (c *Client) getReplicaSetOwner(ctx context.Context, namespace, rsName string) (string, error) {
	rs, err := c.ClientSet.AppsV1().ReplicaSets(namespace).Get(ctx, rsName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get replica set %q: %v", rsName, err)
	}

	for _, owner := range rs.OwnerReferences {
		return owner.Kind, nil
	}

	return "", fmt.Errorf("no owner found for replica set %q", rsName)
}

// TODO: maybe use k8s type, not string

func (c *Client) getReplicaSetFromPod(ctx context.Context, namespace, podName string) (string, error) {
	pod, err := c.ClientSet.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod %q: %v", podName, err)
	}

	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "ReplicaSet" {
			return owner.Name, nil
		}
	}

	return "", nil
}

// FIXME: uint isnt a good idea, use what k8s use

func (c *Client) getReplicas(ctx context.Context, namespace, name string, kind string) (uint, error) {
	var replicas *int32

	switch kind {
	case "Deployment":
		rs, err := c.ClientSet.AppsV1().ReplicaSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to get ReplicaSet %q: %v", name, err)
		}
		replicas = rs.Spec.Replicas
	case "StatefulSet":
		sts, err := c.ClientSet.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to get StatefulSet %q: %v", name, err)
		}
		replicas = sts.Spec.Replicas
	default:
		return 0, fmt.Errorf("unsupported kind: %q", kind)
	}

	return uint(*replicas), nil
}
