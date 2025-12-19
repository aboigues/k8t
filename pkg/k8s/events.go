package k8s

import (
	"context"
	"fmt"
	"sort"

	"github.com/aboigues/k8t/pkg/output"
	"github.com/aboigues/k8t/pkg/types"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPodEvents fetches events related to a specific pod
func (c *Client) GetPodEvents(ctx context.Context, namespace, podName string) (*corev1.EventList, error) {
	// Validate inputs
	if err := ValidateNamespace(namespace); err != nil {
		return nil, fmt.Errorf("invalid namespace: %w", err)
	}
	if err := ValidatePodName(podName); err != nil {
		return nil, fmt.Errorf("invalid pod name: %w", err)
	}

	// Create field selector for pod-specific events
	fieldSelector := fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName)

	// Fetch events from Kubernetes API
	eventList, err := c.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		if k8serrors.IsForbidden(err) || k8serrors.IsUnauthorized(err) {
			return nil, fmt.Errorf("insufficient permissions to list events in namespace '%s': %w", namespace, err)
		}
		return nil, fmt.Errorf("failed to list events for pod '%s' in namespace '%s': %w", podName, namespace, err)
	}

	// Sort events by timestamp (oldest first)
	sort.Slice(eventList.Items, func(i, j int) bool {
		return eventList.Items[i].FirstTimestamp.Time.Before(eventList.Items[j].FirstTimestamp.Time)
	})

	return eventList, nil
}

// FilterImagePullEvents filters events related to image pulling failures
func FilterImagePullEvents(events []corev1.Event) []corev1.Event {
	var filtered []corev1.Event

	for _, event := range events {
		// Filter by reason - looking for image pull related events
		reason := event.Reason
		if reason == "Failed" ||
		   reason == "BackOff" ||
		   reason == "ErrImagePull" ||
		   reason == "ImagePullBackOff" ||
		   reason == "FailedPull" ||
		   reason == "InspectFailed" {
			filtered = append(filtered, event)
		}
	}

	return filtered
}

// ConvertToEventSummary converts K8s events to types.EventSummary
// The redact parameter controls whether sensitive information should be redacted from messages
func ConvertToEventSummary(events []corev1.Event, redact bool) []types.EventSummary {
	summaries := make([]types.EventSummary, 0, len(events))

	for _, event := range events {
		message := event.Message

		// Redact sensitive information if requested
		if redact {
			message = output.RedactEventMessage(message)
		}

		summary := types.EventSummary{
			Timestamp: event.LastTimestamp.Time,
			Reason:    event.Reason,
			Message:   message,
			Count:     int(event.Count),
			FirstSeen: event.FirstTimestamp.Time,
			LastSeen:  event.LastTimestamp.Time,
		}

		summaries = append(summaries, summary)
	}

	return summaries
}
