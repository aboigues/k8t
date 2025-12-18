package types

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// ImageReference represents a parsed container image reference
type ImageReference struct {
	ContainerName string `json:"container_name" yaml:"container_name"`
	FullReference string `json:"full_reference" yaml:"full_reference"` // Full image string from pod spec
	Registry      string `json:"registry" yaml:"registry"`             // e.g., "docker.io", "gcr.io"
	Repository    string `json:"repository" yaml:"repository"`         // e.g., "library/nginx"
	Tag           string `json:"tag,omitempty" yaml:"tag,omitempty"`   // e.g., "1.21"
	Digest        string `json:"digest,omitempty" yaml:"digest,omitempty"` // e.g., "sha256:abc123..."
	IsDigest      bool   `json:"is_digest" yaml:"is_digest"`           // FR-014: tag-based vs digest-based
}

// ParseImageReference parses image reference string into components
// Handles formats:
//   - nginx
//   - nginx:1.21
//   - docker.io/library/nginx:1.21
//   - gcr.io/project/image@sha256:abc123...
func ParseImageReference(containerName, imageRef string) (*ImageReference, error) {
	if imageRef == "" {
		return nil, errors.New("image reference cannot be empty")
	}

	ref := &ImageReference{
		ContainerName: containerName,
		FullReference: imageRef,
	}

	// Split by @ for digest-based references
	parts := strings.Split(imageRef, "@")
	if len(parts) == 2 {
		ref.IsDigest = true
		ref.Digest = parts[1]
		imageRef = parts[0] // Process the rest as normal
	}

	// Split by : for tag or port
	tagParts := strings.Split(imageRef, ":")

	// Check if we have registry/repo:tag or just repo:tag
	var registryAndRepo string
	if len(tagParts) >= 2 && !ref.IsDigest {
		// Last part might be a tag
		possibleTag := tagParts[len(tagParts)-1]
		// Check if it looks like a tag (no slashes) and not a port number
		if !strings.Contains(possibleTag, "/") {
			ref.Tag = possibleTag
			registryAndRepo = strings.Join(tagParts[:len(tagParts)-1], ":")
		} else {
			registryAndRepo = imageRef
		}
	} else {
		registryAndRepo = imageRef
	}

	// Parse registry and repository
	repoParts := strings.Split(registryAndRepo, "/")

	switch len(repoParts) {
	case 1:
		// Just image name (e.g., "nginx")
		ref.Registry = "docker.io"
		ref.Repository = "library/" + repoParts[0]
		if ref.Tag == "" && !ref.IsDigest {
			ref.Tag = "latest"
		}
	case 2:
		// Two parts: could be registry/image or user/image
		if strings.Contains(repoParts[0], ".") || strings.Contains(repoParts[0], ":") {
			// Contains dot or colon, treat as registry
			ref.Registry = repoParts[0]
			ref.Repository = repoParts[1]
		} else {
			// No dot, treat as docker.io/user/image
			ref.Registry = "docker.io"
			ref.Repository = strings.Join(repoParts, "/")
		}
		if ref.Tag == "" && !ref.IsDigest {
			ref.Tag = "latest"
		}
	default:
		// Three or more parts: registry/path/to/image
		ref.Registry = repoParts[0]
		ref.Repository = strings.Join(repoParts[1:], "/")
		if ref.Tag == "" && !ref.IsDigest {
			ref.Tag = "latest"
		}
	}

	return ref, nil
}

// DiagnosticFinding represents analysis results for container image pull issues
type DiagnosticFinding struct {
	// Core identification
	RootCause RootCause `json:"root_cause" yaml:"root_cause"`
	Severity  Severity  `json:"severity" yaml:"severity"`

	// Affected resources
	PodName            string   `json:"pod_name" yaml:"pod_name"`
	PodNamespace       string   `json:"pod_namespace" yaml:"pod_namespace"`
	AffectedContainers []string `json:"affected_containers" yaml:"affected_containers"` // Grouped by root cause (clarification Q4)

	// Diagnostic details
	Summary          string   `json:"summary" yaml:"summary"`
	Details          string   `json:"details" yaml:"details"`
	RemediationSteps []string `json:"remediation_steps" yaml:"remediation_steps"`

	// Context
	ImageReferences []ImageReference `json:"image_references" yaml:"image_references"`
	Events          []EventSummary   `json:"events,omitempty" yaml:"events,omitempty"`

	// Failure analysis
	IsTransient      bool       `json:"is_transient" yaml:"is_transient"`
	FailureCount     int        `json:"failure_count" yaml:"failure_count"`
	FirstFailureTime *time.Time `json:"first_failure_time,omitempty" yaml:"first_failure_time,omitempty"`
	LastFailureTime  *time.Time `json:"last_failure_time,omitempty" yaml:"last_failure_time,omitempty"`
	FailureDuration  string     `json:"failure_duration,omitempty" yaml:"failure_duration,omitempty"` // Human-readable

	// Network diagnostics (when RootCause = NETWORK_ISSUE)
	NetworkDiagnostics *NetworkDiagnostics `json:"network_diagnostics,omitempty" yaml:"network_diagnostics,omitempty"`
}

// Validate checks if finding is well-formed
func (f *DiagnosticFinding) Validate() error {
	if f.PodName == "" {
		return errors.New("pod_name is required")
	}
	if f.PodNamespace == "" {
		return errors.New("pod_namespace is required")
	}
	if len(f.AffectedContainers) == 0 {
		return errors.New("at least one affected container required")
	}
	if len(f.RemediationSteps) == 0 {
		return errors.New("remediation_steps required (FR-002, SC-003)")
	}
	// Persistent failure check (clarification Q2: 3+ failures over 5+ minutes)
	if !f.IsTransient && f.FailureCount < 3 {
		return fmt.Errorf("persistent failure requires 3+ failures, got %d", f.FailureCount)
	}
	return nil
}

// EventSummary represents key information from a Kubernetes Event
type EventSummary struct {
	Timestamp time.Time `json:"timestamp" yaml:"timestamp"`
	Reason    string    `json:"reason" yaml:"reason"`         // e.g., "Failed", "BackOff"
	Message   string    `json:"message" yaml:"message"`       // Event message (SR-007: credentials redacted)
	Count     int       `json:"count" yaml:"count"`           // Event count (repeated events)
	FirstSeen time.Time `json:"first_seen" yaml:"first_seen"`
	LastSeen  time.Time `json:"last_seen" yaml:"last_seen"`
}

// NetworkDiagnostics contains results of network connectivity tests
type NetworkDiagnostics struct {
	RegistryHost  string      `json:"registry_host" yaml:"registry_host"`
	DNSResolution *DNSResult  `json:"dns_resolution" yaml:"dns_resolution"`
	TCPConnection *TCPResult  `json:"tcp_connection" yaml:"tcp_connection"`
	HTTPCheck     *HTTPResult `json:"http_check" yaml:"http_check"`
}

// DNSResult represents DNS lookup results
type DNSResult struct {
	Success      bool     `json:"success" yaml:"success"`
	ResolvedIPs  []string `json:"resolved_ips,omitempty" yaml:"resolved_ips,omitempty"`
	ErrorMessage string   `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	DurationMs   int64    `json:"duration_ms" yaml:"duration_ms"`
}

// TCPResult represents TCP connection test results
type TCPResult struct {
	Success      bool   `json:"success" yaml:"success"`
	Port         int    `json:"port" yaml:"port"` // Typically 443 for HTTPS registries
	ErrorMessage string `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	DurationMs   int64  `json:"duration_ms" yaml:"duration_ms"`
}

// HTTPResult represents HTTP HEAD request results
type HTTPResult struct {
	Success      bool   `json:"success" yaml:"success"`
	StatusCode   int    `json:"status_code,omitempty" yaml:"status_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	DurationMs   int64  `json:"duration_ms" yaml:"duration_ms"`
}
