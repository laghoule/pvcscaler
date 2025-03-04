package k8s

import (
	"context"
	"fmt"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FIXME: uint isnt a good idea, use what k8s use

func (c *Client) getReplicas(ctx context.Context, namespace, name string, kind string) (uint, error) {
	var replicas *int32

	switch kind {
	case deploymentKind:
		dep, err := c.ClientSet.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to get deployment %q: %v", name, err)
		}
		replicas = dep.Spec.Replicas

	case statefulSetKind:
		sts, err := c.ClientSet.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to get statefulset %q: %v", name, err)
		}
		replicas = sts.Spec.Replicas

	default:
		return 0, fmt.Errorf("failed to get replicas for %s, unsupported kind: %q", name, kind)
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
	case deploymentKind:
		_, err := k8sClient.ClientSet.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, dryRun)
		if err != nil {
			return fmt.Errorf("failed to scale down deployment %q: %v", name, err)
		}

	case statefulSetKind:
		_, err := k8sClient.ClientSet.AppsV1().StatefulSets(namespace).UpdateScale(ctx, name, scale, dryRun)
		if err != nil {
			return fmt.Errorf("failed to scale down statefulset %q: %v", name, err)
		}

	default:
		return fmt.Errorf("failed to scale down, unsupported kind: %q", kind)
	}

	return nil
}
