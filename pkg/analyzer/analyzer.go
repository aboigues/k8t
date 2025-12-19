package analyzer

import (
	"context"
	"fmt"
	"time"

	"github.com/aboigues/k8t/pkg/k8s"
	"github.com/aboigues/k8t/pkg/output"
	"github.com/aboigues/k8t/pkg/types"
)

// Analyzer coordinates diagnostic analysis for ImagePullBackOff issues
type Analyzer struct {
	k8sClient   *k8s.Client
	auditLogger *output.AuditLogger
	timeout     time.Duration
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(client *k8s.Client, logger *output.AuditLogger, timeout time.Duration) *Analyzer {
	return &Analyzer{
		k8sClient:   client,
		auditLogger: logger,
		timeout:     timeout,
	}
}

// AnalyzePod performs complete analysis on a single pod
func (a *Analyzer) AnalyzePod(ctx context.Context, namespace, podName string) (*types.AnalysisReport, error) {
	// Set timeout for the entire analysis operation
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// Log analysis start
	a.auditLogger.LogAnalysisStart(types.TargetTypePod, podName, namespace)
	startTime := time.Now()

	// Fetch pod
	a.auditLogger.LogPodGet(podName, namespace)
	pod, err := a.k8sClient.GetPod(ctx, namespace, podName)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, NewTimeoutError("GetPod", a.timeout)
		}
		// Check if it's a not found error
		if isNotFoundError(err) {
			return nil, NewPodNotFoundError(namespace, podName)
		}
		return nil, fmt.Errorf("failed to fetch pod: %w", err)
	}

	// Check if pod has ImagePullBackOff status
	affectedContainers := k8s.GetAffectedContainers(pod)
	if len(affectedContainers) == 0 {
		// No ImagePullBackOff issue found
		report := &types.AnalysisReport{
			TargetType:   types.TargetTypePod,
			TargetName:   podName,
			Namespace:    namespace,
			GeneratedAt:  time.Now(),
			Findings:     []types.DiagnosticFinding{},
			Summary: types.ReportSummary{
				TotalPodsAnalyzed:    1,
				PodsWithIssues:       0,
				RootCauseBreakdown:   make(map[types.RootCause]int),
				HighSeverityCount:    0,
				MediumSeverityCount:  0,
				LowSeverityCount:     0,
			},
		}
		a.auditLogger.LogAnalysisComplete(types.TargetTypePod, podName, namespace, 0)
		return report, nil
	}

	// Fetch events
	a.auditLogger.LogEventList(namespace)
	eventList, err := a.k8sClient.GetPodEvents(ctx, namespace, podName)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, NewTimeoutError("GetPodEvents", a.timeout)
		}
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	// Filter to image pull events
	imagePullEvents := k8s.FilterImagePullEvents(eventList.Items)

	// Convert to EventSummary (with redaction)
	eventSummaries := k8s.ConvertToEventSummary(imagePullEvents, true)

	// Parse events for analysis
	eventAnalysis := ParseEvents(eventSummaries)

	// Detect root cause
	rootCause := DetectRootCause(eventSummaries, pod, eventAnalysis)

	// Extract image references
	imageRefs := k8s.GetContainerImages(pod)

	// Find the primary image reference (first affected container's image)
	var primaryImageRef *types.ImageReference
	if len(imageRefs) > 0 {
		// Find the image ref for the first affected container
		for _, imgRef := range imageRefs {
			for _, affectedContainer := range affectedContainers {
				if imgRef.ContainerName == affectedContainer {
					primaryImageRef = &imgRef
					break
				}
			}
			if primaryImageRef != nil {
				break
			}
		}
		// Fall back to first image if not found
		if primaryImageRef == nil {
			primaryImageRef = &imageRefs[0]
		}
	}

	// Generate remediation steps
	remediationSteps := GenerateRemediationSteps(rootCause, primaryImageRef)

	// Build diagnostic finding
	finding := a.buildFinding(pod, eventSummaries, rootCause, affectedContainers, imageRefs, remediationSteps, eventAnalysis)

	// Count severity
	highCount, mediumCount, lowCount := 0, 0, 0
	switch finding.Severity {
	case types.SeverityHigh:
		highCount = 1
	case types.SeverityMedium:
		mediumCount = 1
	case types.SeverityLow:
		lowCount = 1
	}

	// Build analysis report
	report := &types.AnalysisReport{
		TargetType:  types.TargetTypePod,
		TargetName:  podName,
		Namespace:   namespace,
		GeneratedAt: time.Now(),
		Findings:    []types.DiagnosticFinding{finding},
		Summary: types.ReportSummary{
			TotalPodsAnalyzed:   1,
			PodsWithIssues:      1,
			TotalContainers:     len(imageRefs),
			ContainersWithIssues: len(affectedContainers),
			RootCauseBreakdown: map[types.RootCause]int{
				rootCause: 1,
			},
			HighSeverityCount:   highCount,
			MediumSeverityCount: mediumCount,
			LowSeverityCount:    lowCount,
		},
		AuditLog: []types.AuditEntry{},
	}

	// Log analysis complete
	duration := time.Since(startTime)
	a.auditLogger.LogAnalysisComplete(types.TargetTypePod, podName, namespace, len(report.Findings))

	// Add duration to report if needed
	_ = duration

	return report, nil
}

// buildFinding creates DiagnosticFinding from analysis data
func (a *Analyzer) buildFinding(
	pod interface{},
	events []types.EventSummary,
	rootCause types.RootCause,
	affectedContainers []string,
	imageRefs []types.ImageReference,
	remediationSteps []string,
	analysis *EventAnalysis,
) types.DiagnosticFinding {
	// Get pod metadata
	podName, namespace := getPodMetadata(pod)

	// Build summary
	summary := fmt.Sprintf("%s: %s", rootCause, rootCause.String())

	// Build details from error messages
	details := "Image pull failures detected."
	if len(analysis.ErrorMessages) > 0 {
		details = analysis.ErrorMessages[0]
		if len(analysis.ErrorMessages) > 1 {
			details += fmt.Sprintf(" (and %d more events)", len(analysis.ErrorMessages)-1)
		}
	}

	// Calculate failure duration
	var failureDuration string
	if !analysis.FirstFailureTime.IsZero() && !analysis.LastFailureTime.IsZero() {
		duration := analysis.LastFailureTime.Sub(analysis.FirstFailureTime)
		failureDuration = formatDuration(duration)
	}

	finding := types.DiagnosticFinding{
		RootCause:          rootCause,
		Severity:           rootCause.Severity(),
		PodName:            podName,
		PodNamespace:       namespace,
		AffectedContainers: affectedContainers,
		Summary:            summary,
		Details:            details,
		RemediationSteps:   remediationSteps,
		ImageReferences:    imageRefs,
		Events:             events,
		IsTransient:        analysis.IsTransient,
		FailureCount:       analysis.FailureCount,
		FirstFailureTime:   &analysis.FirstFailureTime,
		LastFailureTime:    &analysis.LastFailureTime,
		FailureDuration:    failureDuration,
	}

	return finding
}

// getPodMetadata extracts pod name and namespace from pod object
func getPodMetadata(pod interface{}) (string, string) {
	// Type assertion to get pod metadata
	// Handle any object with GetName() and GetNamespace() methods
	if p, ok := pod.(interface{ GetName() string; GetNamespace() string }); ok {
		return p.GetName(), p.GetNamespace()
	}
	return "", ""
}

// isNotFoundError checks if an error indicates resource not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	// Simple string matching for "not found" errors
	errMsg := err.Error()
	return contains(errMsg, "not found")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && hasSubstring(s, substr))
}

// hasSubstring checks for substring presence
func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// formatDuration formats a duration in human-readable format
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%d minutes %d seconds", minutes, seconds)
	}
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%d hours %d minutes", hours, minutes)
}
