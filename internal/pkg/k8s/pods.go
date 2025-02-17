package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: maybe use k8s type, not string

func (c *Client) getPodOwnerKind(ctx context.Context, namespace, podName string) (string, error) {
	pod, err := c.ClientSet.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod %q: %v", podName, err)
	}

	for _, owner := range pod.OwnerReferences {
		return owner.Kind, nil
	}

	return "", nil
}

func (c *Client) getPodOwnerName(ctx context.Context, namespace, podName string) (string, error) {
	pod, err := c.ClientSet.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod %q: %v", podName, err)
	}

	for _, oref := range pod.OwnerReferences {
		if oref.Kind == "ReplicaSet" {
			owner, err := c.getReplicaSetOwner(ctx, namespace, oref.Name)
			if err != nil {
				return "", err
			}
			return owner, nil
		} else {
			return oref.Name, nil
		}

	}

	return "", nil
}
