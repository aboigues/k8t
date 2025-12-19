package analyzer

import (
	"time"

	"github.com/aboigues/k8t/pkg/types"
	corev1 "k8s.io/api/core/v1"
)

// EventAnalysis contains parsed event data for diagnostic analysis
type EventAnalysis struct {
	FailureCount     int
	FirstFailureTime time.Time
	LastFailureTime  time.Time
	ErrorMessages    []string
	IsTransient      bool // <3 failures AND <5 minutes
}

// ParseEvents extracts diagnostic information from Kubernetes events
func ParseEvents(events []types.EventSummary) *EventAnalysis {
	analysis := &EventAnalysis{
		ErrorMessages: make([]string, 0, len(events)),
	}

	if len(events) == 0 {
		return analysis
	}

	// Process events to extract failure information
	for _, event := range events {
		// Count failures
		if event.Count > 0 {
			analysis.FailureCount += event.Count
		} else {
			analysis.FailureCount++
		}

		// Track first and last failure times
		if analysis.FirstFailureTime.IsZero() || event.FirstSeen.Before(analysis.FirstFailureTime) {
			analysis.FirstFailureTime = event.FirstSeen
		}
		if analysis.LastFailureTime.IsZero() || event.LastSeen.After(analysis.LastFailureTime) {
			analysis.LastFailureTime = event.LastSeen
		}

		// Extract error message
		if event.Message != "" {
			analysis.ErrorMessages = append(analysis.ErrorMessages, event.Message)
		}
	}

	// Determine if failure is transient
	// Transient: < 3 failures AND duration < 5 minutes
	duration := analysis.LastFailureTime.Sub(analysis.FirstFailureTime)
	analysis.IsTransient = analysis.FailureCount < 3 && duration < 5*time.Minute

	return analysis
}

// ParseEventsFromK8s parses Kubernetes Event objects directly
func ParseEventsFromK8s(events []corev1.Event) *EventAnalysis {
	analysis := &EventAnalysis{
		ErrorMessages: make([]string, 0, len(events)),
	}

	if len(events) == 0 {
		return analysis
	}

	// Process events to extract failure information
	for _, event := range events {
		// Only count failure-related events
		if !isFailureEvent(&event) {
			continue
		}

		// Count failures
		if event.Count > 0 {
			analysis.FailureCount += int(event.Count)
		} else {
			analysis.FailureCount++
		}

		// Track first and last failure times
		if analysis.FirstFailureTime.IsZero() || event.FirstTimestamp.Time.Before(analysis.FirstFailureTime) {
			analysis.FirstFailureTime = event.FirstTimestamp.Time
		}
		if analysis.LastFailureTime.IsZero() || event.LastTimestamp.Time.After(analysis.LastFailureTime) {
			analysis.LastFailureTime = event.LastTimestamp.Time
		}

		// Extract error message
		if event.Message != "" {
			analysis.ErrorMessages = append(analysis.ErrorMessages, event.Message)
		}
	}

	// Determine if failure is transient
	// Transient: < 3 failures AND duration < 5 minutes
	if analysis.FailureCount > 0 {
		duration := analysis.LastFailureTime.Sub(analysis.FirstFailureTime)
		analysis.IsTransient = analysis.FailureCount < 3 && duration < 5*time.Minute
	}

	return analysis
}

// ExtractErrorMessage gets the primary error message from an event
func ExtractErrorMessage(event *corev1.Event) string {
	if event == nil {
		return ""
	}

	// The Message field contains the primary error information
	return event.Message
}

// isFailureEvent checks if an event represents a failure
func isFailureEvent(event *corev1.Event) bool {
	reason := event.Reason
	return reason == "Failed" ||
		reason == "BackOff" ||
		reason == "ErrImagePull" ||
		reason == "ImagePullBackOff" ||
		reason == "FailedPull" ||
		reason == "InspectFailed"
}
