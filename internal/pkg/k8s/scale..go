package k8s

import (
	"fmt"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *Client) getReplicas(namespace, name string, kind string) (int32, error) {
	var replicas *int32

	switch kind {
	case deploymentKind:
		dep, err := c.ClientSet.AppsV1().Deployments(namespace).Get(c.Context, name, metav1.GetOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to get deployment %q: %v", name, err)
		}
		replicas = dep.Spec.Replicas

	case statefulSetKind:
		sts, err := c.ClientSet.AppsV1().StatefulSets(namespace).Get(c.Context, name, metav1.GetOptions{})
		if err != nil {
			return 0, fmt.Errorf("failed to get statefulset %q: %v", name, err)
		}
		replicas = sts.Spec.Replicas

	default:
		return 0, fmt.Errorf("failed to get replicas for %s, unsupported kind: %q", name, kind)
	}

	return *replicas, nil
}

func (w *Workload) ScaleDown(k8sClient *Client, namespace, name, kind string) error {
	return w.scale(k8sClient, namespace, name, kind, 0)
}

func (w *Workload) ScaleUp(k8sClient *Client, namespace, name, kind string, replicas int32) error {
	return w.scale(k8sClient, namespace, name, kind, replicas)
}

func (w *Workload) scale(k8sClient *Client, namespace, name, kind string, replicas int32) error {
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
		_, err := k8sClient.ClientSet.AppsV1().Deployments(namespace).UpdateScale(k8sClient.Context, name, scale, dryRun)
		if err != nil {
			return fmt.Errorf("failed to scale down deployment %q: %v", name, err)
		}

	case statefulSetKind:
		_, err := k8sClient.ClientSet.AppsV1().StatefulSets(namespace).UpdateScale(k8sClient.Context, name, scale, dryRun)
		if err != nil {
			return fmt.Errorf("failed to scale down statefulset %q: %v", name, err)
		}

	default:
		return fmt.Errorf("failed to scale down, unsupported kind: %q", kind)
	}

	return nil
}
