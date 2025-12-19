package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client with error handling
type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
}

// NewClient initializes Kubernetes client from kubeconfig
// Uses the following precedence:
// 1. kubeconfigPath parameter (if provided)
// 2. KUBECONFIG environment variable
// 3. ~/.kube/config (default)
// 4. In-cluster config (if running inside a pod)
func NewClient(kubeconfigPath string) (*Client, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first if no kubeconfig specified
	if kubeconfigPath == "" {
		config, err = rest.InClusterConfig()
		if err == nil {
			clientset, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, fmt.Errorf("failed to create in-cluster client: %w", err)
			}
			return &Client{
				Clientset: clientset,
				Config:    config,
			}, nil
		}
	}

	// Fall back to kubeconfig
	kubeconfig := kubeconfigPath
	if kubeconfig == "" {
		// Check KUBECONFIG env var
		if envKubeconfig := os.Getenv("KUBECONFIG"); envKubeconfig != "" {
			kubeconfig = envKubeconfig
		} else {
			// Default to ~/.kube/config
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Check if kubeconfig file exists
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, fmt.Errorf("kubeconfig file not found at %s", kubeconfig)
	}

	// Build config from kubeconfig file
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to build config from kubeconfig %s: %w", kubeconfig, err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Client{
		Clientset: clientset,
		Config:    config,
	}, nil
}

// Validate checks if client can communicate with the cluster
func (c *Client) Validate() error {
	if c.Clientset == nil {
		return fmt.Errorf("kubernetes clientset is nil")
	}

	// Try to get server version as a simple connectivity test
	_, err := c.Clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("failed to communicate with kubernetes cluster: %w", err)
	}

	return nil
}

// ListNamespaces returns a list of all namespace names in the cluster
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	if c.Clientset == nil {
		return nil, fmt.Errorf("kubernetes clientset is nil")
	}

	namespaces, err := c.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	names := make([]string, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}

	return names, nil
}

// ListPodsInNamespace returns a list of all pods in the specified namespace
func (c *Client) ListPodsInNamespace(ctx context.Context, namespace string) ([]PodInfo, error) {
	if c.Clientset == nil {
		return nil, fmt.Errorf("kubernetes clientset is nil")
	}

	pods, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	podInfos := make([]PodInfo, 0, len(pods.Items))
	for _, pod := range pods.Items {
		podInfo := PodInfo{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Status:    pod.Status,
		}

		// Extract container statuses (both regular and init containers)
		for _, cs := range pod.Status.ContainerStatuses {
			podInfo.ContainerStatuses = append(podInfo.ContainerStatuses, cs)
		}
		for _, cs := range pod.Status.InitContainerStatuses {
			podInfo.ContainerStatuses = append(podInfo.ContainerStatuses, cs)
		}

		podInfos = append(podInfos, podInfo)
	}

	return podInfos, nil
}
