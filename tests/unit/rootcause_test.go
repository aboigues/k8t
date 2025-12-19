package unit

import (
	"testing"

	"github.com/aboigues/k8t/pkg/types"
)

func TestRootCause_String(t *testing.T) {
	tests := []struct {
		name     string
		cause    types.RootCause
		expected string
	}{
		{
			name:     "IMAGE_NOT_FOUND",
			cause:    types.RootCauseImageNotFound,
			expected: "Image does not exist in registry",
		},
		{
			name:     "AUTHENTICATION_FAILURE",
			cause:    types.RootCauseAuthFailure,
			expected: "Registry authentication failed",
		},
		{
			name:     "NETWORK_ISSUE",
			cause:    types.RootCauseNetworkIssue,
			expected: "Cannot reach registry",
		},
		{
			name:     "RATE_LIMIT_EXCEEDED",
			cause:    types.RootCauseRateLimit,
			expected: "Registry rate limit exceeded",
		},
		{
			name:     "PERMISSION_DENIED",
			cause:    types.RootCausePermissionDenied,
			expected: "Insufficient permissions to pull image",
		},
		{
			name:     "MANIFEST_ERROR",
			cause:    types.RootCauseManifestError,
			expected: "Image manifest is invalid or corrupted",
		},
		{
			name:     "TRANSIENT_FAILURE",
			cause:    types.RootCauseTransient,
			expected: "Transient failure (may resolve automatically)",
		},
		{
			name:     "UNKNOWN",
			cause:    types.RootCauseUnknown,
			expected: "Unknown failure reason",
		},
		{
			name:     "Invalid value defaults to unknown",
			cause:    types.RootCause("INVALID"),
			expected: "Unknown failure reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cause.String()
			if result != tt.expected {
				t.Errorf("RootCause.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRootCause_Severity(t *testing.T) {
	tests := []struct {
		name     string
		cause    types.RootCause
		expected types.Severity
	}{
		{
			name:     "IMAGE_NOT_FOUND is HIGH severity",
			cause:    types.RootCauseImageNotFound,
			expected: types.SeverityHigh,
		},
		{
			name:     "AUTHENTICATION_FAILURE is HIGH severity",
			cause:    types.RootCauseAuthFailure,
			expected: types.SeverityHigh,
		},
		{
			name:     "PERMISSION_DENIED is HIGH severity",
			cause:    types.RootCausePermissionDenied,
			expected: types.SeverityHigh,
		},
		{
			name:     "NETWORK_ISSUE is MEDIUM severity",
			cause:    types.RootCauseNetworkIssue,
			expected: types.SeverityMedium,
		},
		{
			name:     "RATE_LIMIT_EXCEEDED is MEDIUM severity",
			cause:    types.RootCauseRateLimit,
			expected: types.SeverityMedium,
		},
		{
			name:     "MANIFEST_ERROR is MEDIUM severity",
			cause:    types.RootCauseManifestError,
			expected: types.SeverityMedium,
		},
		{
			name:     "TRANSIENT_FAILURE is LOW severity",
			cause:    types.RootCauseTransient,
			expected: types.SeverityLow,
		},
		{
			name:     "UNKNOWN is MEDIUM severity",
			cause:    types.RootCauseUnknown,
			expected: types.SeverityMedium,
		},
		{
			name:     "Invalid value defaults to MEDIUM severity",
			cause:    types.RootCause("INVALID"),
			expected: types.SeverityMedium,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cause.Severity()
			if result != tt.expected {
				t.Errorf("RootCause.Severity() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSeverity_Color(t *testing.T) {
	tests := []struct {
		name     string
		severity types.Severity
		expected string
	}{
		{
			name:     "HIGH is red",
			severity: types.SeverityHigh,
			expected: "\033[31m",
		},
		{
			name:     "MEDIUM is yellow",
			severity: types.SeverityMedium,
			expected: "\033[33m",
		},
		{
			name:     "LOW is green",
			severity: types.SeverityLow,
			expected: "\033[32m",
		},
		{
			name:     "Invalid defaults to reset",
			severity: types.Severity("INVALID"),
			expected: "\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.severity.Color()
			if result != tt.expected {
				t.Errorf("Severity.Color() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAllRootCauseConstants(t *testing.T) {
	// Ensure all constants are defined correctly
	allCauses := []types.RootCause{
		types.RootCauseImageNotFound,
		types.RootCauseAuthFailure,
		types.RootCauseNetworkIssue,
		types.RootCauseRateLimit,
		types.RootCausePermissionDenied,
		types.RootCauseManifestError,
		types.RootCauseTransient,
		types.RootCauseUnknown,
	}

	for _, cause := range allCauses {
		t.Run(string(cause), func(t *testing.T) {
			// Ensure String() doesn't panic
			str := cause.String()
			if str == "" {
				t.Errorf("RootCause.String() returned empty string for %s", cause)
			}

			// Ensure Severity() doesn't panic and returns valid severity
			severity := cause.Severity()
			if severity != types.SeverityHigh && severity != types.SeverityMedium && severity != types.SeverityLow {
				t.Errorf("RootCause.Severity() returned invalid severity %s for %s", severity, cause)
			}
		})
	}
}
