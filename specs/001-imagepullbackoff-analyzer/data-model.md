# Data Model: ImagePullBackOff Analyzer

**Feature**: 001-imagepullbackoff-analyzer
**Date**: 2025-12-18
**Phase**: Phase 1 - Design

## Overview

This document defines the core data structures for the ImagePullBackOff Analyzer. All types are defined in Go and map directly to the domain model extracted from the specification.

## Core Entities

### 1. RootCause

**Purpose**: Enumeration of possible root causes for ImagePullBackOff errors

**Definition**:
```go
package types

// RootCause represents the category of ImagePullBackOff failure
type RootCause string

const (
	RootCauseImageNotFound     RootCause = "IMAGE_NOT_FOUND"
	RootCauseAuthFailure       RootCause = "AUTHENTICATION_FAILURE"
	RootCauseNetworkIssue      RootCause = "NETWORK_ISSUE"
	RootCauseRateLimit         RootCause = "RATE_LIMIT_EXCEEDED"
	RootCausePermissionDenied  RootCause = "PERMISSION_DENIED"
	RootCauseManifestError     RootCause = "MANIFEST_ERROR"
	RootCauseTransient         RootCause = "TRANSIENT_FAILURE"
	RootCauseUnknown           RootCause = "UNKNOWN"
)

// String returns human-readable description
func (r RootCause) String() string {
	switch r {
	case RootCauseImageNotFound:
		return "Image does not exist in registry"
	case RootCauseAuthFailure:
		return "Registry authentication failed"
	case RootCauseNetworkIssue:
		return "Cannot reach registry"
	case RootCauseRateLimit:
		return "Registry rate limit exceeded"
	case RootCausePermissionDenied:
		return "Insufficient permissions to pull image"
	case RootCauseManifestError:
		return "Image manifest is invalid or corrupted"
	case RootCauseTransient:
		return "Transient failure (may resolve automatically)"
	default:
		return "Unknown failure reason"
	}
}

// Severity returns the urgency level
func (r RootCause) Severity() Severity {
	switch r {
	case RootCauseImageNotFound, RootCauseAuthFailure, RootCausePermissionDenied:
		return SeverityHigh // Requires immediate action
	case RootCauseNetworkIssue, RootCauseRateLimit, RootCauseManifestError:
		return SeverityMedium // Needs investigation
	case RootCauseTransient:
		return SeverityLow // May self-resolve
	default:
		return SeverityMedium
	}
}
```

**Validation Rules**:
- RootCause must be one of the defined constants
- Unknown should only be used when no other category matches
- Mapped from K8s event messages via pattern matching (see FR-001)

**Relationships**:
- One RootCause per DiagnosticFinding
- Multiple containers may have same or different RootCauses

---

### 2. Severity

**Purpose**: Indicates urgency of diagnostic finding

**Definition**:
```go
package types

type Severity string

const (
	SeverityHigh   Severity = "HIGH"   // Requires immediate action
	SeverityMedium Severity = "MEDIUM" // Needs investigation
	SeverityLow    Severity = "LOW"    // Informational or may self-resolve
)

// Color returns ANSI color code for terminal output
func (s Severity) Color() string {
	switch s {
	case SeverityHigh:
		return "\033[31m" // Red
	case SeverityMedium:
		return "\033[33m" // Yellow
	case SeverityLow:
		return "\033[32m" // Green
	default:
		return "\033[0m" // Reset
	}
}
```

---

### 3. DiagnosticFinding

**Purpose**: Represents a single diagnostic finding for one or more containers

**Definition**:
```go
package types

import (
	"time"
)

// DiagnosticFinding represents analysis results for container image pull issues
type DiagnosticFinding struct {
	// Core identification
	RootCause         RootCause       `json:"root_cause" yaml:"root_cause"`
	Severity          Severity        `json:"severity" yaml:"severity"`

	// Affected resources
	PodName           string          `json:"pod_name" yaml:"pod_name"`
	PodNamespace      string          `json:"pod_namespace" yaml:"pod_namespace"`
	AffectedContainers []string       `json:"affected_containers" yaml:"affected_containers"` // Grouped by root cause (clarification Q4)

	// Diagnostic details
	Summary           string          `json:"summary" yaml:"summary"`
	Details           string          `json:"details" yaml:"details"`
	RemediationSteps  []string        `json:"remediation_steps" yaml:"remediation_steps"`

	// Context
	ImageReferences   []ImageReference `json:"image_references" yaml:"image_references"`
	Events            []EventSummary   `json:"events,omitempty" yaml:"events,omitempty"`

	// Failure analysis
	IsTransient       bool            `json:"is_transient" yaml:"is_transient"`
	FailureCount      int             `json:"failure_count" yaml:"failure_count"`
	FirstFailureTime  *time.Time      `json:"first_failure_time,omitempty" yaml:"first_failure_time,omitempty"`
	LastFailureTime   *time.Time      `json:"last_failure_time,omitempty" yaml:"last_failure_time,omitempty"`
	FailureDuration   string          `json:"failure_duration,omitempty" yaml:"failure_duration,omitempty"` // Human-readable

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
		return errors.New("persistent failure requires 3+ failures")
	}
	return nil
}
```

**Validation Rules** (from spec):
- FR-002: RemediationSteps must not be empty (SC-003: 100% actionable)
- FR-013: IsTransient determined by failure count and duration (clarification: 3+ failures over 5+ minutes = persistent)
- FR-015: AffectedContainers grouped by RootCause (clarification Q4)
- SR-007: All fields safe to share (no credentials)

**Relationships**:
- Contains 1+ ImageReference
- May contain 0+ EventSummary
- May contain 1 NetworkDiagnostics (conditional)

---

### 4. ImageReference

**Purpose**: Parsed container image specification

**Definition**:
```go
package types

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

// Parse parses image reference string into components
func ParseImageReference(containerName, imageRef string) (*ImageReference, error) {
	// Parsing logic for formats:
	// - nginx
	// - nginx:1.21
	// - docker.io/library/nginx:1.21
	// - gcr.io/project/image@sha256:abc123...
	// Handles FR-014 requirement
}
```

**Validation Rules**:
- FR-014: Must support both tag-based and digest-based references
- Registry defaults to "docker.io" if not specified
- Either Tag or Digest must be present (not both)

---

### 5. EventSummary

**Purpose**: Simplified representation of relevant Kubernetes events

**Definition**:
```go
package types

import (
	"time"
)

// EventSummary represents key information from a Kubernetes Event
type EventSummary struct {
	Timestamp     time.Time `json:"timestamp" yaml:"timestamp"`
	Reason        string    `json:"reason" yaml:"reason"`         // e.g., "Failed", "BackOff"
	Message       string    `json:"message" yaml:"message"`       // Event message (SR-007: credentials redacted)
	Count         int       `json:"count" yaml:"count"`           // Event count (repeated events)
	FirstSeen     time.Time `json:"first_seen" yaml:"first_seen"`
	LastSeen      time.Time `json:"last_seen" yaml:"last_seen"`
}
```

**Validation Rules**:
- SR-007: Message must be sanitized (redact secrets, tokens)
- FR-012: Handle missing events gracefully (count may be 0)
- Events sorted by Timestamp descending (most recent first)

---

### 6. NetworkDiagnostics

**Purpose**: Detailed network connectivity test results

**Definition**:
```go
package types

// NetworkDiagnostics contains results of network connectivity tests
type NetworkDiagnostics struct {
	RegistryHost    string         `json:"registry_host" yaml:"registry_host"`
	DNSResolution   *DNSResult     `json:"dns_resolution" yaml:"dns_resolution"`
	TCPConnection   *TCPResult     `json:"tcp_connection" yaml:"tcp_connection"`
	HTTPCheck       *HTTPResult    `json:"http_check" yaml:"http_check"`
}

// DNSResult represents DNS lookup results
type DNSResult struct {
	Success       bool     `json:"success" yaml:"success"`
	ResolvedIPs   []string `json:"resolved_ips,omitempty" yaml:"resolved_ips,omitempty"`
	ErrorMessage  string   `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	DurationMs    int64    `json:"duration_ms" yaml:"duration_ms"`
}

// TCPResult represents TCP connection test results
type TCPResult struct {
	Success       bool   `json:"success" yaml:"success"`
	Port          int    `json:"port" yaml:"port"`           // Typically 443 for HTTPS registries
	ErrorMessage  string `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	DurationMs    int64  `json:"duration_ms" yaml:"duration_ms"`
}

// HTTPResult represents HTTP HEAD request results
type HTTPResult struct {
	Success       bool   `json:"success" yaml:"success"`
	StatusCode    int    `json:"status_code,omitempty" yaml:"status_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty" yaml:"error_message,omitempty"`
	DurationMs    int64  `json:"duration_ms" yaml:"duration_ms"`
}
```

**Validation Rules** (from clarification Q3):
- FR-005: Must perform DNS + TCP + HTTP HEAD (full connectivity check)
- All duration measurements in milliseconds
- ErrorMessage sanitized (no sensitive data)

---

### 7. AnalysisReport

**Purpose**: Top-level container for analysis results

**Definition**:
```go
package types

import (
	"time"
)

// AnalysisReport contains complete analysis results for one or more pods
type AnalysisReport struct {
	// Metadata
	GeneratedAt   time.Time  `json:"generated_at" yaml:"generated_at"`
	ToolVersion   string     `json:"tool_version" yaml:"tool_version"`

	// Scope
	TargetType    TargetType `json:"target_type" yaml:"target_type"` // pod, deployment, namespace
	TargetName    string     `json:"target_name" yaml:"target_name"`
	Namespace     string     `json:"namespace" yaml:"namespace"`

	// Results
	Summary       ReportSummary        `json:"summary" yaml:"summary"`
	Findings      []DiagnosticFinding  `json:"findings" yaml:"findings"`

	// Audit trail (SR-004)
	AuditLog      []AuditEntry         `json:"audit_log,omitempty" yaml:"audit_log,omitempty"`
}

// TargetType indicates analysis scope
type TargetType string

const (
	TargetTypePod         TargetType = "pod"         // FR-008: Single pod
	TargetTypeWorkload    TargetType = "workload"    // FR-009: Deployment/StatefulSet/etc
	TargetTypeNamespace   TargetType = "namespace"   // FR-010: All pods in namespace
)

// ReportSummary provides high-level overview
type ReportSummary struct {
	TotalPodsAnalyzed     int `json:"total_pods_analyzed" yaml:"total_pods_analyzed"`
	PodsWithIssues        int `json:"pods_with_issues" yaml:"pods_with_issues"`
	TotalContainers       int `json:"total_containers" yaml:"total_containers"`
	ContainersWithIssues  int `json:"containers_with_issues" yaml:"containers_with_issues"`

	// Root cause breakdown
	RootCauseBreakdown    map[RootCause]int `json:"root_cause_breakdown" yaml:"root_cause_breakdown"`

	// Severity breakdown
	HighSeverityCount     int `json:"high_severity_count" yaml:"high_severity_count"`
	MediumSeverityCount   int `json:"medium_severity_count" yaml:"medium_severity_count"`
	LowSeverityCount      int `json:"low_severity_count" yaml:"low_severity_count"`
}

// AuditEntry records cluster access for audit trail
type AuditEntry struct {
	Timestamp    time.Time `json:"timestamp" yaml:"timestamp"`
	ResourceType string    `json:"resource_type" yaml:"resource_type"` // "pods", "events", "secrets"
	ResourceName string    `json:"resource_name" yaml:"resource_name"`
	Namespace    string    `json:"namespace" yaml:"namespace"`
	Operation    string    `json:"operation" yaml:"operation"`         // "get", "list"
}
```

**Validation Rules**:
- SR-004: AuditLog captures all cluster access (stdout/stderr simple format per clarification Q5)
- FR-015: Findings grouped by RootCause for multi-container pods
- Summary aggregates across all findings

---

## State Transitions

### ImagePullBackOff Failure States

```
                                    ┌─────────────────┐
                                    │   Pod Created   │
                                    └────────┬────────┘
                                             │
                                             ▼
                           ┌─────────────────────────────────┐
                           │   Image Pull Attempted          │
                           └──────┬──────────────────────────┘
                                  │
                ┌─────────────────┼─────────────────┐
                │                 │                 │
                ▼                 ▼                 ▼
        ┌────────────┐    ┌──────────────┐   ┌─────────────┐
        │  Success   │    │   Transient  │   │ Persistent  │
        │  (ignore)  │    │   Failure    │   │   Failure   │
        └────────────┘    └──────┬───────┘   └──────┬──────┘
                                  │                  │
                     ┌────────────┴──────────────────┘
                     │
                     ▼
          ┌──────────────────────┐
          │  Root Cause Analysis │
          └──────────┬───────────┘
                     │
       ┌─────────────┼─────────────┬─────────────────┬──────────────┐
       │             │             │                 │              │
       ▼             ▼             ▼                 ▼              ▼
 ┌──────────┐ ┌───────────┐ ┌──────────────┐ ┌──────────┐  ┌───────────┐
 │ Not Found│ │Auth Failure│ │Network Issue│ │Rate Limit│  │  Other    │
 └────┬─────┘ └─────┬─────┘ └──────┬───────┘ └────┬─────┘  └─────┬─────┘
      │             │               │              │              │
      └─────────────┴───────────────┴──────────────┴──────────────┘
                                    │
                                    ▼
                    ┌──────────────────────────────┐
                    │  Generate Remediation Steps  │
                    └──────────────────────────────┘
```

**Transient vs Persistent** (clarification Q2):
- **Transient**: <3 failures OR <5 minutes duration
- **Persistent**: ≥3 consecutive failures over ≥5 minutes

---

## Data Flow

### Analysis Pipeline

```
Input (Pod/Workload/Namespace)
         │
         ▼
    ┌────────────────┐
    │  Fetch Pods    │ ← K8s API (pods/list)
    └────────┬───────┘
             │
             ▼
    ┌────────────────┐
    │  Fetch Events  │ ← K8s API (events/list)
    └────────┬───────┘
             │
             ▼
    ┌────────────────────────────┐
    │  Parse Events & Containers │
    └────────┬───────────────────┘
             │
             ▼
    ┌─────────────────────────┐
    │  Root Cause Detection   │
    └────────┬────────────────┘
             │
     ┌───────┴────────┐
     │                │
     ▼                ▼
┌──────────┐   ┌───────────────┐
│Auth Check│   │Network Testing│ (if needed)
│(Secrets) │   │(DNS/TCP/HTTP) │
└────┬─────┘   └───────┬───────┘
     │                 │
     └────────┬────────┘
              │
              ▼
    ┌──────────────────────┐
    │Generate Findings +   │
    │Remediation Steps     │
    └────────┬─────────────┘
             │
             ▼
    ┌──────────────────────┐
    │  Format Output       │
    │ (Text/JSON/YAML)     │
    └──────────────────────┘
```

---

## Persistence

**None** - Tool is stateless (per Technical Context):
- No database
- No file storage
- All data in-memory during analysis
- Results output to stdout/stderr

---

## Security Considerations

### Data Sanitization (SR-007)

All types must ensure:
- **No credential leaks**: Redact passwords, tokens, secrets
- **No sensitive ConfigMap data**: Unless explicitly requested
- **Safe to share**: Output can be shared in public channels

**Redaction Rules**:
```go
// Example redaction patterns
var sensitivePatterns = []string{
	`password[=:]\s*\S+`,
	`token[=:]\s*\S+`,
	`secret[=:]\s*\S+`,
	`Authorization:\s*\S+`,
}

func Redact(message string) string {
	// Replace sensitive patterns with "[REDACTED]"
}
```

Applied to:
- EventSummary.Message
- DiagnosticFinding.Details
- All error messages

---

## Testing Considerations

### Mock Data

Test data structures required:
1. **Sample Pods**: With various ImagePullBackOff scenarios
2. **Sample Events**: Covering all root cause categories
3. **Sample Findings**: Expected outputs for each scenario

### Test Coverage

Required test scenarios per data type:
- **RootCause**: All enum values, severity mapping
- **DiagnosticFinding**: Validation rules, transient/persistent logic
- **NetworkDiagnostics**: All three check types (DNS/TCP/HTTP)
- **AnalysisReport**: Aggregation logic, summary calculations

---

## Performance Constraints

Per Technical Context and Success Criteria:

| Metric | Target | Data Model Impact |
|--------|--------|-------------------|
| Memory footprint | <50MB | Keep findings in slice, no caching |
| Analysis time | <10s P95 | Efficient event parsing, minimal allocations |
| Root cause accuracy | >90% | Comprehensive pattern matching in RootCause detection |

**Optimization Strategies**:
- Pre-compile regex patterns for event parsing
- Reuse buffers for string operations
- Limit event history processed (last 100 events per pod)
- Use pointer receivers for large structs

---

## Future Extensions

Potential additions (post-MVP):
- **Trend Analysis**: Track failures over time (requires persistence)
- **ML-based Root Cause**: Enhanced pattern detection
- **Remediation Automation**: Auto-fix common issues (requires write permissions)
- **Multi-Cluster Support**: Aggregate findings across clusters

None of these are in current scope (Out of Scope section in spec).
