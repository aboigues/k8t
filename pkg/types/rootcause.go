package types

// RootCause represents the category of pod failure (ImagePullBackOff, CrashLoopBackOff, etc.)
type RootCause string

const (
	// ImagePullBackOff root causes
	RootCauseImageNotFound    RootCause = "IMAGE_NOT_FOUND"
	RootCauseAuthFailure      RootCause = "AUTHENTICATION_FAILURE"
	RootCauseNetworkIssue     RootCause = "NETWORK_ISSUE"
	RootCauseRateLimit        RootCause = "RATE_LIMIT_EXCEEDED"
	RootCausePermissionDenied RootCause = "PERMISSION_DENIED"
	RootCauseManifestError    RootCause = "MANIFEST_ERROR"
	RootCauseTransient        RootCause = "TRANSIENT_FAILURE"

	// CrashLoopBackOff root causes
	RootCauseOOMKilled           RootCause = "OOM_KILLED"
	RootCauseApplicationError    RootCause = "APPLICATION_ERROR"
	RootCauseConfigError         RootCause = "CONFIG_ERROR"
	RootCauseMissingDependency   RootCause = "MISSING_DEPENDENCY"
	RootCauseProbeFailure        RootCause = "PROBE_FAILURE"
	RootCausePermissionError     RootCause = "PERMISSION_ERROR"
	RootCausePortConflict        RootCause = "PORT_CONFLICT"
	RootCauseExitCodeError       RootCause = "EXIT_CODE_ERROR"

	// Generic root causes
	RootCauseUnknown RootCause = "UNKNOWN"
)

// String returns human-readable description
func (r RootCause) String() string {
	switch r {
	// ImagePullBackOff root causes
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

	// CrashLoopBackOff root causes
	case RootCauseOOMKilled:
		return "Container exceeded memory limits (OOMKilled)"
	case RootCauseApplicationError:
		return "Application crashed due to internal error"
	case RootCauseConfigError:
		return "Missing or invalid configuration"
	case RootCauseMissingDependency:
		return "Required service or resource unavailable"
	case RootCauseProbeFailure:
		return "Liveness or readiness probe failing"
	case RootCausePermissionError:
		return "Filesystem or security context permissions issue"
	case RootCausePortConflict:
		return "Port already in use or binding failed"
	case RootCauseExitCodeError:
		return "Container exited with non-zero exit code"

	default:
		return "Unknown failure reason"
	}
}

// Severity returns the urgency level
func (r RootCause) Severity() Severity {
	switch r {
	// High severity - requires immediate action
	case RootCauseImageNotFound, RootCauseAuthFailure, RootCausePermissionDenied,
		RootCauseOOMKilled, RootCauseConfigError, RootCausePermissionError:
		return SeverityHigh

	// Medium severity - needs investigation
	case RootCauseNetworkIssue, RootCauseRateLimit, RootCauseManifestError,
		RootCauseApplicationError, RootCauseMissingDependency, RootCauseProbeFailure,
		RootCausePortConflict, RootCauseExitCodeError:
		return SeverityMedium

	// Low severity - may self-resolve
	case RootCauseTransient:
		return SeverityLow

	default:
		return SeverityMedium
	}
}

// Severity indicates urgency of diagnostic finding
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
