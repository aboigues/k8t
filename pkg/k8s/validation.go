package k8s

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// DNS-1123 label format for Kubernetes names (RFC 1123)
	// Must consist of lower case alphanumeric characters or '-',
	// start and end with alphanumeric, and be at most 63 characters
	dns1123LabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

	// Maximum length for Kubernetes resource names
	maxNameLength = 253
)

// ValidateNamespace validates a Kubernetes namespace name
// Prevents injection attacks and ensures RFC 1123 compliance (SR-005)
func ValidateNamespace(namespace string) error {
	if namespace == "" {
		return errors.New("namespace cannot be empty")
	}

	if len(namespace) > maxNameLength {
		return fmt.Errorf("namespace exceeds maximum length of %d characters", maxNameLength)
	}

	// Check for injection patterns
	if containsInjectionPatterns(namespace) {
		return errors.New("namespace contains invalid characters or injection patterns")
	}

	// Validate against DNS-1123 label format
	if !dns1123LabelRegex.MatchString(namespace) {
		return errors.New("namespace must consist of lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character")
	}

	return nil
}

// ValidatePodName validates a Kubernetes pod name
// Prevents injection attacks and ensures RFC 1123 compliance (SR-005)
func ValidatePodName(podName string) error {
	if podName == "" {
		return errors.New("pod name cannot be empty")
	}

	if len(podName) > maxNameLength {
		return fmt.Errorf("pod name exceeds maximum length of %d characters", maxNameLength)
	}

	// Check for injection patterns
	if containsInjectionPatterns(podName) {
		return errors.New("pod name contains invalid characters or injection patterns")
	}

	// Validate against DNS-1123 subdomain format
	// Pod names can contain dots, so we split and validate each label
	labels := strings.Split(podName, ".")
	for _, label := range labels {
		if !dns1123LabelRegex.MatchString(label) {
			return fmt.Errorf("pod name label '%s' is invalid: must consist of lowercase alphanumeric characters or '-', and must start and end with an alphanumeric character", label)
		}
	}

	return nil
}

// ValidateWorkloadName validates a Kubernetes workload name (deployment, statefulset, etc.)
func ValidateWorkloadName(workloadName string) error {
	// Same validation as pod names
	if workloadName == "" {
		return errors.New("workload name cannot be empty")
	}

	if len(workloadName) > maxNameLength {
		return fmt.Errorf("workload name exceeds maximum length of %d characters", maxNameLength)
	}

	if containsInjectionPatterns(workloadName) {
		return errors.New("workload name contains invalid characters or injection patterns")
	}

	labels := strings.Split(workloadName, ".")
	for _, label := range labels {
		if !dns1123LabelRegex.MatchString(label) {
			return fmt.Errorf("workload name label '%s' is invalid", label)
		}
	}

	return nil
}

// containsInjectionPatterns checks for common injection attack patterns
// Prevents command injection, path traversal, and other attacks (SR-005)
func containsInjectionPatterns(input string) bool {
	// Check for common injection patterns
	injectionPatterns := []string{
		"..",          // Path traversal
		"/",           // Absolute paths (Kubernetes names don't contain /)
		"\\",          // Windows paths / escape sequences
		"$",           // Shell variable expansion
		"`",           // Command substitution
		";",           // Command chaining
		"|",           // Pipe
		"&",           // Background execution
		"<",           // Input redirection
		">",           // Output redirection
		"*",           // Wildcards
		"?",           // Wildcards
		"[",           // Character classes
		"]",           // Character classes
		"{",           // Brace expansion
		"}",           // Brace expansion
		"(",           // Subshell
		")",           // Subshell
		"'",           // Quote
		"\"",          // Quote
		"\n",          // Newline
		"\r",          // Carriage return
		"\t",          // Tab (except in normal Kubernetes names)
		"\x00",        // Null byte
	}

	for _, pattern := range injectionPatterns {
		if strings.Contains(input, pattern) {
			return true
		}
	}

	return false
}

// SanitizeInput sanitizes user input by trimming whitespace
// This is a helper for non-critical inputs
func SanitizeInput(input string) string {
	return strings.TrimSpace(input)
}
