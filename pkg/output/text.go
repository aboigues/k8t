package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/aboigues/k8t/pkg/types"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

// formatTextOutput renders report as human-readable text with ANSI colors
func formatTextOutput(report *types.AnalysisReport, noColor bool, w io.Writer) error {
	if report == nil {
		return fmt.Errorf("report cannot be nil")
	}

	var b strings.Builder

	// Header
	b.WriteString(formatHeader("IMAGEPULLBACKOFF ANALYSIS REPORT", noColor))
	b.WriteString("\n")
	b.WriteString(formatField("Target", fmt.Sprintf("%s/%s", report.Namespace, report.TargetName), noColor))
	b.WriteString(formatField("Type", string(report.TargetType), noColor))
	b.WriteString(formatField("Generated At", report.GeneratedAt.Format("2006-01-02 15:04:05 MST"), noColor))
	b.WriteString("\n")

	// Summary
	b.WriteString(formatSection("SUMMARY", noColor))
	b.WriteString(formatField("Pods Analyzed", fmt.Sprintf("%d", report.Summary.TotalPodsAnalyzed), noColor))
	b.WriteString(formatField("Pods with Issues", fmt.Sprintf("%d", report.Summary.PodsWithIssues), noColor))

	if report.Summary.PodsWithIssues == 0 {
		b.WriteString("\n")
		b.WriteString(colorize("No ImagePullBackOff issues found.", colorGreen, noColor))
		b.WriteString("\n")
		_, err := w.Write([]byte(b.String()))
		return err
	}

	// By Root Cause
	if len(report.Summary.RootCauseBreakdown) > 0 {
		b.WriteString(formatField("By Root Cause", "", noColor))
		for cause, count := range report.Summary.RootCauseBreakdown {
			b.WriteString(fmt.Sprintf("  - %s: %d\n", cause, count))
		}
	}

	// By Severity
	totalSeverity := report.Summary.HighSeverityCount + report.Summary.MediumSeverityCount + report.Summary.LowSeverityCount
	if totalSeverity > 0 {
		b.WriteString(formatField("By Severity", "", noColor))
		if report.Summary.HighSeverityCount > 0 {
			b.WriteString(fmt.Sprintf("  - %s: %d\n", colorize("HIGH", colorRed, noColor), report.Summary.HighSeverityCount))
		}
		if report.Summary.MediumSeverityCount > 0 {
			b.WriteString(fmt.Sprintf("  - %s: %d\n", colorize("MEDIUM", colorYellow, noColor), report.Summary.MediumSeverityCount))
		}
		if report.Summary.LowSeverityCount > 0 {
			b.WriteString(fmt.Sprintf("  - %s: %d\n", colorize("LOW", colorGreen, noColor), report.Summary.LowSeverityCount))
		}
	}

	b.WriteString("\n")

	// Findings
	for i, finding := range report.Findings {
		b.WriteString(formatSection(fmt.Sprintf("FINDING #%d", i+1), noColor))

		// Root Cause and Severity
		severityColor := getSeverityColor(finding.Severity)
		b.WriteString(formatField("Root Cause", string(finding.RootCause), noColor))
		b.WriteString(formatField("Severity", colorize(string(finding.Severity), severityColor, noColor), noColor))
		b.WriteString(formatField("Pod", fmt.Sprintf("%s/%s", finding.PodNamespace, finding.PodName), noColor))

		// Affected Containers
		if len(finding.AffectedContainers) > 0 {
			b.WriteString(formatField("Affected Containers", strings.Join(finding.AffectedContainers, ", "), noColor))
		}

		// Summary and Details
		b.WriteString(formatField("Summary", finding.Summary, noColor))
		if finding.Details != "" {
			b.WriteString(formatField("Details", finding.Details, noColor))
		}

		// Failure Information
		if finding.FailureCount > 0 {
			b.WriteString(formatField("Failure Count", fmt.Sprintf("%d", finding.FailureCount), noColor))
		}
		if finding.FailureDuration != "" {
			b.WriteString(formatField("Failure Duration", finding.FailureDuration, noColor))
		}
		if finding.IsTransient {
			b.WriteString(formatField("Status", colorize("TRANSIENT (may self-resolve)", colorYellow, noColor), noColor))
		} else {
			b.WriteString(formatField("Status", colorize("PERSISTENT (requires action)", colorRed, noColor), noColor))
		}

		// Image References
		if len(finding.ImageReferences) > 0 {
			b.WriteString("\n")
			b.WriteString(colorize("IMAGE REFERENCES:", colorBold, noColor))
			b.WriteString("\n")
			for _, img := range finding.ImageReferences {
				b.WriteString(fmt.Sprintf("  Container: %s\n", img.ContainerName))
				b.WriteString(fmt.Sprintf("    Image: %s\n", img.FullReference))
				b.WriteString(fmt.Sprintf("    Registry: %s\n", img.Registry))
				b.WriteString(fmt.Sprintf("    Repository: %s\n", img.Repository))
				if img.Tag != "" {
					b.WriteString(fmt.Sprintf("    Tag: %s\n", img.Tag))
				}
				if img.Digest != "" {
					b.WriteString(fmt.Sprintf("    Digest: %s\n", img.Digest))
				}
			}
		}

		// Remediation Steps
		if len(finding.RemediationSteps) > 0 {
			b.WriteString("\n")
			b.WriteString(colorize("REMEDIATION STEPS:", colorBold, noColor))
			b.WriteString("\n")
			for i, step := range finding.RemediationSteps {
				b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step))
			}
		}

		// Events (condensed)
		if len(finding.Events) > 0 {
			b.WriteString("\n")
			b.WriteString(colorize("RECENT EVENTS:", colorBold, noColor))
			b.WriteString(fmt.Sprintf(" (showing last %d)\n", min(len(finding.Events), 5)))
			displayCount := min(len(finding.Events), 5)
			for i := len(finding.Events) - displayCount; i < len(finding.Events); i++ {
				event := finding.Events[i]
				b.WriteString(fmt.Sprintf("  [%s] %s: %s\n",
					event.Timestamp.Format("15:04:05"),
					event.Reason,
					truncate(event.Message, 100)))
			}
		}

		b.WriteString("\n")
	}

	// Footer
	b.WriteString(formatDivider(noColor))
	b.WriteString(colorize("For more information, visit: https://kubernetes.io/docs/concepts/containers/images/", colorGray, noColor))
	b.WriteString("\n")

	_, err := w.Write([]byte(b.String()))
	return err
}

// Helper functions

func formatHeader(title string, noColor bool) string {
	divider := strings.Repeat("=", len(title))
	return fmt.Sprintf("%s\n%s\n%s\n",
		colorize(divider, colorBold, noColor),
		colorize(title, colorBold, noColor),
		colorize(divider, colorBold, noColor))
}

func formatSection(title string, noColor bool) string {
	return fmt.Sprintf("%s\n%s\n",
		colorize(title, colorBold, noColor),
		colorize(strings.Repeat("-", len(title)), colorGray, noColor))
}

func formatField(label, value string, noColor bool) string {
	if value == "" {
		return fmt.Sprintf("%s:\n", colorize(label, colorBlue, noColor))
	}
	return fmt.Sprintf("%s: %s\n", colorize(label, colorBlue, noColor), value)
}

func formatDivider(noColor bool) string {
	return colorize(strings.Repeat("=", 80), colorGray, noColor) + "\n"
}

func colorize(text string, color string, noColor bool) string {
	if noColor {
		return text
	}
	return color + text + colorReset
}

func getSeverityColor(severity types.Severity) string {
	switch severity {
	case types.SeverityHigh:
		return colorRed
	case types.SeverityMedium:
		return colorYellow
	case types.SeverityLow:
		return colorGreen
	default:
		return colorReset
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
