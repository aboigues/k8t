package analyzer

import (
	"strings"
	"time"

	"github.com/aboigues/k8t/pkg/types"
	corev1 "k8s.io/api/core/v1"
)

// rootCausePatterns defines substring patterns for detecting each root cause type
var rootCausePatterns = map[types.RootCause][]string{
	types.RootCauseImageNotFound: {
		"manifest unknown",
		"manifest not found",
		"not found: manifest unknown",
		"image not found",
		"repository does not exist",
		"404",
	},
	types.RootCauseAuthFailure: {
		"unauthorized",
		"authentication required",
		"authentication failed",
		"authorization failed",
		"401",
		"403",
		"no basic auth credentials",
		"pull access denied",
		"access denied",
		"access forbidden",
		"denied: access forbidden",
	},
	types.RootCauseNetworkIssue: {
		"dial tcp",
		"timeout",
		"i/o timeout",
		"connection refused",
		"no route to host",
		"dns",
		"failed",
		"lookup",
		"no such host",
	},
	types.RootCauseRateLimit: {
		"rate limit",
		"too many requests",
		"429",
		"toomanyrequests",
	},
	types.RootCausePermissionDenied: {
		"forbidden",
		"permission denied",
		"insufficient",
		"permission",
	},
	types.RootCauseManifestError: {
		"manifest invalid",
		"unsupported",
		"platform",
		"no matching manifest",
		"unknown blob",
	},
}

// crashLoopPatterns defines patterns for detecting CrashLoopBackOff root causes
var crashLoopPatterns = map[types.RootCause][]string{
	types.RootCauseConfigError: {
		"failed to load config",
		"configuration error",
		"config file not found",
		"missing required configuration",
		"environment variable",
		"configmap",
		"secret",
		"no such file or directory",
		"cannot open",
	},
	types.RootCauseMissingDependency: {
		"connection refused",
		"dial tcp",
		"cannot connect to",
		"database",
		"redis",
		"mongodb",
		"service not found",
		"unknown host",
		"failed to connect",
	},
	types.RootCausePortConflict: {
		"address already in use",
		"bind: address already in use",
		"port is already allocated",
		"cannot bind to",
		"failed to bind",
	},
	types.RootCausePermissionError: {
		"permission denied",
		"access denied",
		"operation not permitted",
		"forbidden",
		"cannot create directory",
		"cannot write",
		"read-only file system",
	},
	types.RootCauseProbeFailure: {
		"liveness probe failed",
		"readiness probe failed",
		"startup probe failed",
		"probe failed",
		"unhealthy",
	},
	types.RootCauseApplicationError: {
		"panic",
		"fatal error",
		"segmentation fault",
		"stack trace",
		"exception",
		"error:",
		"failed to start",
	},
}

// DetectRootCause determines the root cause from event messages
// Uses priority ordering: IMAGE_NOT_FOUND > AUTH > NETWORK > RATE_LIMIT > PERMISSION > MANIFEST > TRANSIENT > UNKNOWN
func DetectRootCause(events []types.EventSummary, pod *corev1.Pod, analysis *EventAnalysis) types.RootCause {
	// Concatenate all event messages for pattern matching
	var messages strings.Builder
	for _, event := range events {
		messages.WriteString(strings.ToLower(event.Message))
		messages.WriteString(" ")
	}
	combinedMessages := messages.String()

	// Check patterns in priority order
	if matchPatterns(combinedMessages, rootCausePatterns[types.RootCauseImageNotFound]) {
		return types.RootCauseImageNotFound
	}

	if matchPatterns(combinedMessages, rootCausePatterns[types.RootCauseAuthFailure]) {
		return types.RootCauseAuthFailure
	}

	if matchPatterns(combinedMessages, rootCausePatterns[types.RootCauseNetworkIssue]) {
		return types.RootCauseNetworkIssue
	}

	if matchPatterns(combinedMessages, rootCausePatterns[types.RootCauseRateLimit]) {
		return types.RootCauseRateLimit
	}

	if matchPatterns(combinedMessages, rootCausePatterns[types.RootCausePermissionDenied]) {
		return types.RootCausePermissionDenied
	}

	if matchPatterns(combinedMessages, rootCausePatterns[types.RootCauseManifestError]) {
		return types.RootCauseManifestError
	}

	// Check for transient failure (logic-based, not pattern-based)
	if analysis != nil && analysis.IsTransient {
		return types.RootCauseTransient
	}

	// Default to unknown if no patterns match
	return types.RootCauseUnknown
}

// DetectCrashLoopRootCause determines the root cause of CrashLoopBackOff from container status, events, and logs
// Priority: OOM_KILLED > PROBE_FAILURE > CONFIG_ERROR > MISSING_DEPENDENCY > PORT_CONFLICT > PERMISSION_ERROR > APPLICATION_ERROR > EXIT_CODE_ERROR > TRANSIENT > UNKNOWN
func DetectCrashLoopRootCause(events []types.EventSummary, pod *corev1.Pod, containerStatuses []corev1.ContainerStatus, logs string, analysis *EventAnalysis) types.RootCause {
	// Check container termination status for OOMKilled
	for _, status := range containerStatuses {
		if status.LastTerminationState.Terminated != nil {
			terminated := status.LastTerminationState.Terminated
			if terminated.Reason == "OOMKilled" {
				return types.RootCauseOOMKilled
			}
		}
		if status.State.Terminated != nil {
			terminated := status.State.Terminated
			if terminated.Reason == "OOMKilled" {
				return types.RootCauseOOMKilled
			}
		}
	}

	// Concatenate all event messages and logs for pattern matching
	var messages strings.Builder
	for _, event := range events {
		messages.WriteString(strings.ToLower(event.Message))
		messages.WriteString(" ")
	}
	if logs != "" {
		messages.WriteString(strings.ToLower(logs))
		messages.WriteString(" ")
	}
	combinedText := messages.String()

	// Check patterns in priority order
	if matchPatterns(combinedText, crashLoopPatterns[types.RootCauseProbeFailure]) {
		return types.RootCauseProbeFailure
	}

	if matchPatterns(combinedText, crashLoopPatterns[types.RootCauseConfigError]) {
		return types.RootCauseConfigError
	}

	if matchPatterns(combinedText, crashLoopPatterns[types.RootCauseMissingDependency]) {
		return types.RootCauseMissingDependency
	}

	if matchPatterns(combinedText, crashLoopPatterns[types.RootCausePortConflict]) {
		return types.RootCausePortConflict
	}

	if matchPatterns(combinedText, crashLoopPatterns[types.RootCausePermissionError]) {
		return types.RootCausePermissionError
	}

	if matchPatterns(combinedText, crashLoopPatterns[types.RootCauseApplicationError]) {
		return types.RootCauseApplicationError
	}

	// Check exit code from container termination
	for _, status := range containerStatuses {
		var exitCode int32
		if status.LastTerminationState.Terminated != nil {
			exitCode = status.LastTerminationState.Terminated.ExitCode
		} else if status.State.Terminated != nil {
			exitCode = status.State.Terminated.ExitCode
		}

		// Non-zero exit code indicates error
		if exitCode > 0 {
			return types.RootCauseExitCodeError
		}
	}

	// Check for transient failure
	if analysis != nil && analysis.IsTransient {
		return types.RootCauseTransient
	}

	// Default to unknown
	return types.RootCauseUnknown
}

// matchPatterns checks if any of the patterns exist in the text
// Returns true if at least one pattern is found
func matchPatterns(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// matchImageNotFound checks for image not found patterns
func matchImageNotFound(message string) bool {
	msg := strings.ToLower(message)
	return strings.Contains(msg, "not found") ||
		strings.Contains(msg, "manifest unknown") ||
		strings.Contains(msg, "does not exist") ||
		strings.Contains(msg, "404")
}

// matchAuthenticationFailure checks for authentication failure patterns
func matchAuthenticationFailure(message string) bool {
	msg := strings.ToLower(message)
	return strings.Contains(msg, "unauthorized") ||
		strings.Contains(msg, "authentication required") ||
		strings.Contains(msg, "401") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "no basic auth credentials") ||
		strings.Contains(msg, "pull access denied")
}

// matchNetworkIssue checks for network issue patterns
func matchNetworkIssue(message string) bool {
	msg := strings.ToLower(message)
	return (strings.Contains(msg, "dial tcp") && strings.Contains(msg, "timeout")) ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no route to host") ||
		(strings.Contains(msg, "dns") && strings.Contains(msg, "failed")) ||
		(strings.Contains(msg, "lookup") && strings.Contains(msg, "no such host"))
}

// matchRateLimitExceeded checks for rate limit patterns
func matchRateLimitExceeded(message string) bool {
	msg := strings.ToLower(message)
	return strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "too many requests") ||
		strings.Contains(msg, "429") ||
		strings.Contains(msg, "toomanyrequests")
}

// matchPermissionDenied checks for permission denied patterns
func matchPermissionDenied(message string) bool {
	msg := strings.ToLower(message)
	return strings.Contains(msg, "forbidden") ||
		strings.Contains(msg, "permission denied") ||
		(strings.Contains(msg, "insufficient") && strings.Contains(msg, "permission"))
}

// matchManifestError checks for manifest error patterns
func matchManifestError(message string) bool {
	msg := strings.ToLower(message)
	return strings.Contains(msg, "manifest invalid") ||
		(strings.Contains(msg, "unsupported") && strings.Contains(msg, "platform")) ||
		strings.Contains(msg, "no matching manifest") ||
		strings.Contains(msg, "unknown blob")
}

// matchTransientFailure checks if the failure is transient based on failure count and duration
func matchTransientFailure(failureCount int, duration time.Duration) bool {
	return failureCount < 3 || duration < 5*time.Minute
}
