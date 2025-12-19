package analyzer

import (
	"fmt"
	"time"
)

// PodNotFoundError indicates that the specified pod does not exist
type PodNotFoundError struct {
	Namespace string
	PodName   string
}

func (e *PodNotFoundError) Error() string {
	return fmt.Sprintf("pod '%s' not found in namespace '%s'", e.PodName, e.Namespace)
}

// NewPodNotFoundError creates a new PodNotFoundError
func NewPodNotFoundError(namespace, podName string) *PodNotFoundError {
	return &PodNotFoundError{
		Namespace: namespace,
		PodName:   podName,
	}
}

// PermissionError indicates insufficient RBAC permissions
type PermissionError struct {
	Resource  string // e.g., "pods", "events"
	Verb      string // e.g., "get", "list"
	Namespace string
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("insufficient permissions: %s/%s in namespace '%s'", e.Resource, e.Verb, e.Namespace)
}

// NewPermissionError creates a new PermissionError
func NewPermissionError(resource, verb, namespace string) *PermissionError {
	return &PermissionError{
		Resource:  resource,
		Verb:      verb,
		Namespace: namespace,
	}
}

// TimeoutError indicates that an operation timed out
type TimeoutError struct {
	Operation string
	Timeout   time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("operation '%s' timed out after %v", e.Operation, e.Timeout)
}

// NewTimeoutError creates a new TimeoutError
func NewTimeoutError(operation string, timeout time.Duration) *TimeoutError {
	return &TimeoutError{
		Operation: operation,
		Timeout:   timeout,
	}
}

// NoImagePullBackOffError indicates the pod does not have ImagePullBackOff status
type NoImagePullBackOffError struct {
	PodName   string
	Namespace string
}

func (e *NoImagePullBackOffError) Error() string {
	return fmt.Sprintf("pod '%s' in namespace '%s' does not have ImagePullBackOff status", e.PodName, e.Namespace)
}

// NewNoImagePullBackOffError creates a new NoImagePullBackOffError
func NewNoImagePullBackOffError(namespace, podName string) *NoImagePullBackOffError {
	return &NoImagePullBackOffError{
		PodName:   podName,
		Namespace: namespace,
	}
}

// ValidationError indicates input validation failed
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}
