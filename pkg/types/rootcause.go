package types

// RootCause represents the category of ImagePullBackOff failure
type RootCause string

const (
	RootCauseImageNotFound    RootCause = "IMAGE_NOT_FOUND"
	RootCauseAuthFailure      RootCause = "AUTHENTICATION_FAILURE"
	RootCauseNetworkIssue     RootCause = "NETWORK_ISSUE"
	RootCauseRateLimit        RootCause = "RATE_LIMIT_EXCEEDED"
	RootCausePermissionDenied RootCause = "PERMISSION_DENIED"
	RootCauseManifestError    RootCause = "MANIFEST_ERROR"
	RootCauseTransient        RootCause = "TRANSIENT_FAILURE"
	RootCauseUnknown          RootCause = "UNKNOWN"
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
