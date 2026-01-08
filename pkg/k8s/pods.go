package k8s

import (
	"context"
	"fmt"

	"github.com/aboigues/k8t/pkg/types"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodInfo contains essential information about a pod for health checks
type PodInfo struct {
	Name              string
	Namespace         string
	Status            corev1.PodStatus
	ContainerStatuses []corev1.ContainerStatus
}

// GetPod fetches a single pod by name in a namespace
func (c *Client) GetPod(ctx context.Context, namespace, podName string) (*corev1.Pod, error) {
	// Validate inputs using existing validation functions
	if err := ValidateNamespace(namespace); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := ValidatePodName(podName); err != nil {
		return nil, fmt.Errorf("invalid pod name: %w", err)
	}

	// Fetch pod from Kubernetes API
	pod, err := c.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("pod '%s' not found in namespace '%s'", podName, namespace)
		}
		if k8serrors.IsForbidden(err) || k8serrors.IsUnauthorized(err) {
			return nil, fmt.Errorf("insufficient permissions to get pod '%s' in namespace '%s': %w", podName, namespace, err)
		}
		return nil, fmt.Errorf("failed to get pod '%s' in namespace '%s': %w", podName, namespace, err)
	}

	return pod, nil
}

// ListPods lists all pods in a namespace
func (c *Client) ListPods(ctx context.Context, namespace string) (*corev1.PodList, error) {
	// Validate namespace
	if err := ValidateNamespace(namespace); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}

	// List pods from Kubernetes API
	podList, err := c.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if k8serrors.IsForbidden(err) || k8serrors.IsUnauthorized(err) {
			return nil, fmt.Errorf("insufficient permissions to list pods in namespace '%s': %w", namespace, err)
		}
		return nil, fmt.Errorf("failed to list pods in namespace '%s': %w", namespace, err)
	}

	return podList, nil
}

// FilterPodsWithImagePullBackOff filters pods with ImagePullBackOff status
func FilterPodsWithImagePullBackOff(pods []corev1.Pod) []corev1.Pod {
	var filtered []corev1.Pod

	for _, pod := range pods {
		hasImagePullBackOff := false

		// Check all container statuses (including init containers)
		allStatuses := append([]corev1.ContainerStatus{}, pod.Status.ContainerStatuses...)
		allStatuses = append(allStatuses, pod.Status.InitContainerStatuses...)

		for _, containerStatus := range allStatuses {
			// Check waiting state for ImagePullBackOff or ErrImagePull
			if containerStatus.State.Waiting != nil {
				reason := containerStatus.State.Waiting.Reason
				if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
					hasImagePullBackOff = true
					break
				}
			}

			// Also check last terminated state (could have failed before)
			if containerStatus.LastTerminationState.Waiting != nil {
				reason := containerStatus.LastTerminationState.Waiting.Reason
				if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
					hasImagePullBackOff = true
					break
				}
			}
		}

		if hasImagePullBackOff {
			filtered = append(filtered, pod)
		}
	}

	return filtered
}

// GetContainerImages extracts container image references from pod spec
func GetContainerImages(pod *corev1.Pod) []types.ImageReference {
	var imageRefs []types.ImageReference

	// Extract images from regular containers
	for _, container := range pod.Spec.Containers {
		imgRef, err := types.ParseImageReference(container.Name, container.Image)
		if err != nil {
			// Log error but continue with other containers
			// In production, this would use the audit logger
			continue
		}
		imageRefs = append(imageRefs, *imgRef)
	}

	// Extract images from init containers
	for _, container := range pod.Spec.InitContainers {
		imgRef, err := types.ParseImageReference(container.Name, container.Image)
		if err != nil {
			// Log error but continue with other containers
			continue
		}
		imageRefs = append(imageRefs, *imgRef)
	}

	return imageRefs
}

// GetAffectedContainers returns names of containers with ImagePullBackOff
func GetAffectedContainers(pod *corev1.Pod) []string {
	var affectedContainers []string

	// Check all container statuses
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
				affectedContainers = append(affectedContainers, containerStatus.Name)
			}
		}
	}

	// Check init container statuses
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			if reason == "ImagePullBackOff" || reason == "ErrImagePull" {
				affectedContainers = append(affectedContainers, containerStatus.Name)
			}
		}
	}

	return affectedContainers
}

// GetAffectedContainersForCrashLoop returns names of containers with CrashLoopBackOff
func GetAffectedContainersForCrashLoop(pod *corev1.Pod) []string {
	var affectedContainers []string

	// Check all container statuses
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// Check waiting state for CrashLoopBackOff
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			if reason == "CrashLoopBackOff" {
				affectedContainers = append(affectedContainers, containerStatus.Name)
			}
		}
		// Also check if container has recently terminated (might not be in waiting state yet)
		if containerStatus.LastTerminationState.Terminated != nil && containerStatus.RestartCount > 0 {
			affectedContainers = append(affectedContainers, containerStatus.Name)
		}
	}

	// Check init container statuses
	for _, containerStatus := range pod.Status.InitContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			if reason == "CrashLoopBackOff" {
				affectedContainers = append(affectedContainers, containerStatus.Name)
			}
		}
		if containerStatus.LastTerminationState.Terminated != nil && containerStatus.RestartCount > 0 {
			affectedContainers = append(affectedContainers, containerStatus.Name)
		}
	}

	return affectedContainers
}

// FilterPodsWithCrashLoopBackOff filters pods with CrashLoopBackOff status
func FilterPodsWithCrashLoopBackOff(pods []corev1.Pod) []corev1.Pod {
	var filtered []corev1.Pod

	for _, pod := range pods {
		hasCrashLoopBackOff := false

		// Check all container statuses (including init containers)
		allStatuses := append([]corev1.ContainerStatus{}, pod.Status.ContainerStatuses...)
		allStatuses = append(allStatuses, pod.Status.InitContainerStatuses...)

		for _, containerStatus := range allStatuses {
			// Check waiting state for CrashLoopBackOff
			if containerStatus.State.Waiting != nil {
				reason := containerStatus.State.Waiting.Reason
				if reason == "CrashLoopBackOff" {
					hasCrashLoopBackOff = true
					break
				}
			}
		}

		if hasCrashLoopBackOff {
			filtered = append(filtered, pod)
		}
	}

	return filtered
}

// GetContainerLogs retrieves logs from a specific container in a pod
func (c *Client) GetContainerLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64) (string, error) {
	// Validate inputs
	if err := ValidateNamespace(namespace); err != nil {
		return "", fmt.Errorf("invalid namespace: %w", err)
	}
	if err := ValidatePodName(podName); err != nil {
		return "", fmt.Errorf("invalid pod name: %w", err)
	}

	// Prepare log options
	logOptions := &corev1.PodLogOptions{
		Container: containerName,
		Previous:  true, // Get logs from previous (crashed) container
	}
	if tailLines > 0 {
		logOptions.TailLines = &tailLines
	}

	// Get logs
	req := c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
	logStream, err := req.Stream(ctx)
	if err != nil {
		// If previous logs don't exist, try current logs
		logOptions.Previous = false
		req = c.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
		logStream, err = req.Stream(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get logs for container '%s' in pod '%s': %w", containerName, podName, err)
		}
	}
	defer logStream.Close()

	// Read logs from stream
	buf := make([]byte, 2048)
	var logs string
	for {
		n, err := logStream.Read(buf)
		if n > 0 {
			logs += string(buf[:n])
		}
		if err != nil {
			break
		}
	}

	return logs, nil
}
