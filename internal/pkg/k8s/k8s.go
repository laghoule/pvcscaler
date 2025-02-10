package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client struct {
	ClientSet kubernetes.Interface
	DryRun    bool
}

func New(kubeconfig string, dryRun bool) (*Client, error) {
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	// Set QPS and Burst for better performance
	config.RateLimiter = nil
	config.QPS = 25
	config.Burst = 50

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Client{
		ClientSet: clientset,
		DryRun:    dryRun,
	}, nil
}

func (c *Client) GetAllNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.ClientSet.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var namespaceNames []string
	for _, namespace := range namespaces.Items {
		namespaceNames = append(namespaceNames, namespace.Name)
	}

	return namespaceNames, nil
}

func (c *Client) getDryRunUpdateOptionMetaV1() metav1.UpdateOptions {
	if c.DryRun {
		return metav1.UpdateOptions{
			DryRun: []string{metav1.DryRunAll},
		}
	}

	return metav1.UpdateOptions{}
}
