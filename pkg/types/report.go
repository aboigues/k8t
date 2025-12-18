package types

import (
	"time"
)

// AnalysisReport contains complete analysis results for one or more pods
type AnalysisReport struct {
	// Metadata
	GeneratedAt time.Time `json:"generated_at" yaml:"generated_at"`
	ToolVersion string    `json:"tool_version" yaml:"tool_version"`

	// Scope
	TargetType TargetType `json:"target_type" yaml:"target_type"` // pod, deployment, namespace
	TargetName string     `json:"target_name" yaml:"target_name"`
	Namespace  string     `json:"namespace" yaml:"namespace"`

	// Results
	Summary  ReportSummary       `json:"summary" yaml:"summary"`
	Findings []DiagnosticFinding `json:"findings" yaml:"findings"`

	// Audit trail (SR-004)
	AuditLog []AuditEntry `json:"audit_log,omitempty" yaml:"audit_log,omitempty"`
}

// TargetType indicates analysis scope
type TargetType string

const (
	TargetTypePod       TargetType = "pod"       // FR-008: Single pod
	TargetTypeWorkload  TargetType = "workload"  // FR-009: Deployment/StatefulSet/etc
	TargetTypeNamespace TargetType = "namespace" // FR-010: All pods in namespace
)

// ReportSummary provides high-level overview
type ReportSummary struct {
	TotalPodsAnalyzed    int `json:"total_pods_analyzed" yaml:"total_pods_analyzed"`
	PodsWithIssues       int `json:"pods_with_issues" yaml:"pods_with_issues"`
	TotalContainers      int `json:"total_containers" yaml:"total_containers"`
	ContainersWithIssues int `json:"containers_with_issues" yaml:"containers_with_issues"`

	// Root cause breakdown
	RootCauseBreakdown map[RootCause]int `json:"root_cause_breakdown" yaml:"root_cause_breakdown"`

	// Severity breakdown
	HighSeverityCount   int `json:"high_severity_count" yaml:"high_severity_count"`
	MediumSeverityCount int `json:"medium_severity_count" yaml:"medium_severity_count"`
	LowSeverityCount    int `json:"low_severity_count" yaml:"low_severity_count"`
}

// AuditEntry records cluster access for audit trail
type AuditEntry struct {
	Timestamp    time.Time `json:"timestamp" yaml:"timestamp"`
	ResourceType string    `json:"resource_type" yaml:"resource_type"` // "pods", "events", "secrets"
	ResourceName string    `json:"resource_name" yaml:"resource_name"`
	Namespace    string    `json:"namespace" yaml:"namespace"`
	Operation    string    `json:"operation" yaml:"operation"` // "get", "list"
}
