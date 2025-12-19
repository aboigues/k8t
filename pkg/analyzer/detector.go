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
