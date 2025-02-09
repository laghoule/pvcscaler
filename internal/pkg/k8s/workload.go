package k8s

import (
	"context"
	"fmt"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Workload struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Replicas  uint   `json:"replicas"`
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

			ownerKind, err := c.getWorkloadOwnerKind(ctx, namespace, pod.Name)
			if err != nil {
				return nil, err
			}

			ownerName, err := c.getPodOwnerName(ctx, namespace, pod.Name)
			if err != nil {
				return nil, err
			}

			replicas, err := c.getReplicas(ctx, namespace, ownerName, ownerKind)
			if err != nil {
				return nil, err
			}

			fmt.Printf("Pod %s (%s: %s) use a PVC of storage class %q\n", pod.Name, ownerKind, ownerName, storageClass)
			workloads = append(workloads, Workload{
				Kind:      ownerKind,
				Name:      ownerName,
				Namespace: namespace,
				Replicas:  replicas,
			})
		}
	}

	return workloads, nil
}

// TODO: maybe use k8s type, not string

func (c *Client) getWorkloadOwnerKind(ctx context.Context, namespace, podName string) (string, error) {
	ownerKind, err := c.getPodOwnerKind(ctx, namespace, podName)
	if err != nil {
		return "", err
	}

	switch ownerKind {
	case "ReplicaSet":
		return "Deployment", nil
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
		return owner.Name, nil
	}

	return "", fmt.Errorf("no owner found for replica set %q", rsName)
}

// FIXME: uint isnt a good idea, use what k8s use

func (c *Client) getReplicas(ctx context.Context, namespace, name string, kind string) (uint, error) {
	var replicas *int32

	switch kind {
	case "Deployment":
		dep, err := c.ClientSet.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to get deployment %q: %v", name, err)
		}
		replicas = dep.Spec.Replicas
	case "StatefulSet":
		sts, err := c.ClientSet.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to get statefulset %q: %v", name, err)
		}
		replicas = sts.Spec.Replicas
	default:
		return 0, fmt.Errorf("failed to get replicas, unsupported kind: %q", kind)
	}

	return uint(*replicas), nil
}

func (w *Workload) ScaleDown(ctx context.Context, k8sClient *Client, namespace, name, kind string) error {
	return w.scale(ctx, k8sClient, namespace, name, kind, 0)
}

func (w *Workload) ScaleUp(ctx context.Context, k8sClient *Client, namespace, name, kind string, replicas int32) error {
	return w.scale(ctx, k8sClient, namespace, name, kind, replicas)
}

func (w *Workload) scale(ctx context.Context, k8sClient *Client, namespace, name, kind string, replicas int32) error {
	scale := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: replicas,
		},
	}

	dryRun := k8sClient.getDryRunUpdateOptionMetaV1()

	switch kind {
	case "Deployment":
		_, err := k8sClient.ClientSet.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, dryRun)
		if err != nil {
			return fmt.Errorf("failed to scale down deployment %q: %v", name, err)
		}
	case "StatefulSet":
		_, err := k8sClient.ClientSet.AppsV1().StatefulSets(namespace).UpdateScale(ctx, name, scale, dryRun)
		if err != nil {
			return fmt.Errorf("failed to scale down statefulset %q: %v", name, err)
		}
	default:
		return fmt.Errorf("failed to scale down, unsupported kind: %q", kind)
	}

	return nil
}
