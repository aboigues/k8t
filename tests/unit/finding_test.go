package unit

import (
	"testing"
	"time"

	"github.com/yourorg/k8t/pkg/types"
)

func TestParseImageReference(t *testing.T) {
	tests := []struct {
		name          string
		containerName string
		imageRef      string
		expected      *types.ImageReference
		expectError   bool
	}{
		{
			name:          "Simple image name only",
			containerName: "nginx-container",
			imageRef:      "nginx",
			expected: &types.ImageReference{
				ContainerName: "nginx-container",
				FullReference: "nginx",
				Registry:      "docker.io",
				Repository:    "library/nginx",
				Tag:           "latest",
				Digest:        "",
				IsDigest:      false,
			},
		},
		{
			name:          "Image with tag",
			containerName: "app",
			imageRef:      "nginx:1.21",
			expected: &types.ImageReference{
				ContainerName: "app",
				FullReference: "nginx:1.21",
				Registry:      "docker.io",
				Repository:    "library/nginx",
				Tag:           "1.21",
				Digest:        "",
				IsDigest:      false,
			},
		},
		{
			name:          "User image on Docker Hub",
			containerName: "app",
			imageRef:      "myuser/myapp:v1.0",
			expected: &types.ImageReference{
				ContainerName: "app",
				FullReference: "myuser/myapp:v1.0",
				Registry:      "docker.io",
				Repository:    "myuser/myapp",
				Tag:           "v1.0",
				Digest:        "",
				IsDigest:      false,
			},
		},
		{
			name:          "Full registry path with tag",
			containerName: "app",
			imageRef:      "gcr.io/my-project/my-app:v2.0",
			expected: &types.ImageReference{
				ContainerName: "app",
				FullReference: "gcr.io/my-project/my-app:v2.0",
				Registry:      "gcr.io",
				Repository:    "my-project/my-app",
				Tag:           "v2.0",
				Digest:        "",
				IsDigest:      false,
			},
		},
		{
			name:          "Image with digest",
			containerName: "app",
			imageRef:      "nginx@sha256:abc123def456",
			expected: &types.ImageReference{
				ContainerName: "app",
				FullReference: "nginx@sha256:abc123def456",
				Registry:      "docker.io",
				Repository:    "library/nginx",
				Tag:           "",
				Digest:        "sha256:abc123def456",
				IsDigest:      true,
			},
		},
		{
			name:          "Full path with digest",
			containerName: "app",
			imageRef:      "gcr.io/my-project/my-app@sha256:xyz789",
			expected: &types.ImageReference{
				ContainerName: "app",
				FullReference: "gcr.io/my-project/my-app@sha256:xyz789",
				Registry:      "gcr.io",
				Repository:    "my-project/my-app",
				Tag:           "",
				Digest:        "sha256:xyz789",
				IsDigest:      true,
			},
		},
		{
			name:          "Private registry with port",
			containerName: "app",
			imageRef:      "registry.example.com:5000/my-app:latest",
			expected: &types.ImageReference{
				ContainerName: "app",
				FullReference: "registry.example.com:5000/my-app:latest",
				Registry:      "registry.example.com:5000",
				Repository:    "my-app",
				Tag:           "latest",
				Digest:        "",
				IsDigest:      false,
			},
		},
		{
			name:          "Deep repository path",
			containerName: "app",
			imageRef:      "quay.io/organization/team/project/app:v1",
			expected: &types.ImageReference{
				ContainerName: "app",
				FullReference: "quay.io/organization/team/project/app:v1",
				Registry:      "quay.io",
				Repository:    "organization/team/project/app",
				Tag:           "v1",
				Digest:        "",
				IsDigest:      false,
			},
		},
		{
			name:          "Empty image reference",
			containerName: "app",
			imageRef:      "",
			expected:      nil,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := types.ParseImageReference(tt.containerName, tt.imageRef)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.ContainerName != tt.expected.ContainerName {
				t.Errorf("ContainerName = %v, want %v", result.ContainerName, tt.expected.ContainerName)
			}
			if result.FullReference != tt.expected.FullReference {
				t.Errorf("FullReference = %v, want %v", result.FullReference, tt.expected.FullReference)
			}
			if result.Registry != tt.expected.Registry {
				t.Errorf("Registry = %v, want %v", result.Registry, tt.expected.Registry)
			}
			if result.Repository != tt.expected.Repository {
				t.Errorf("Repository = %v, want %v", result.Repository, tt.expected.Repository)
			}
			if result.Tag != tt.expected.Tag {
				t.Errorf("Tag = %v, want %v", result.Tag, tt.expected.Tag)
			}
			if result.Digest != tt.expected.Digest {
				t.Errorf("Digest = %v, want %v", result.Digest, tt.expected.Digest)
			}
			if result.IsDigest != tt.expected.IsDigest {
				t.Errorf("IsDigest = %v, want %v", result.IsDigest, tt.expected.IsDigest)
			}
		})
	}
}

func TestDiagnosticFinding_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		finding     *types.DiagnosticFinding
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid persistent finding",
			finding: &types.DiagnosticFinding{
				PodName:            "test-pod",
				PodNamespace:       "default",
				AffectedContainers: []string{"container1"},
				RemediationSteps:   []string{"Step 1", "Step 2"},
				IsTransient:        false,
				FailureCount:       3,
			},
			expectError: false,
		},
		{
			name: "Valid transient finding",
			finding: &types.DiagnosticFinding{
				PodName:            "test-pod",
				PodNamespace:       "default",
				AffectedContainers: []string{"container1"},
				RemediationSteps:   []string{"Step 1"},
				IsTransient:        true,
				FailureCount:       2,
			},
			expectError: false,
		},
		{
			name: "Missing pod name",
			finding: &types.DiagnosticFinding{
				PodName:            "",
				PodNamespace:       "default",
				AffectedContainers: []string{"container1"},
				RemediationSteps:   []string{"Step 1"},
			},
			expectError: true,
			errorMsg:    "pod_name is required",
		},
		{
			name: "Missing pod namespace",
			finding: &types.DiagnosticFinding{
				PodName:            "test-pod",
				PodNamespace:       "",
				AffectedContainers: []string{"container1"},
				RemediationSteps:   []string{"Step 1"},
			},
			expectError: true,
			errorMsg:    "pod_namespace is required",
		},
		{
			name: "No affected containers",
			finding: &types.DiagnosticFinding{
				PodName:            "test-pod",
				PodNamespace:       "default",
				AffectedContainers: []string{},
				RemediationSteps:   []string{"Step 1"},
			},
			expectError: true,
			errorMsg:    "at least one affected container required",
		},
		{
			name: "No remediation steps",
			finding: &types.DiagnosticFinding{
				PodName:            "test-pod",
				PodNamespace:       "default",
				AffectedContainers: []string{"container1"},
				RemediationSteps:   []string{},
			},
			expectError: true,
			errorMsg:    "remediation_steps required",
		},
		{
			name: "Persistent with insufficient failures",
			finding: &types.DiagnosticFinding{
				PodName:            "test-pod",
				PodNamespace:       "default",
				AffectedContainers: []string{"container1"},
				RemediationSteps:   []string{"Step 1"},
				IsTransient:        false,
				FailureCount:       2, // Less than 3
			},
			expectError: true,
			errorMsg:    "persistent failure requires 3+ failures",
		},
		{
			name: "Complete finding with all fields",
			finding: &types.DiagnosticFinding{
				RootCause:          types.RootCauseImageNotFound,
				Severity:           types.SeverityHigh,
				PodName:            "test-pod",
				PodNamespace:       "default",
				AffectedContainers: []string{"nginx", "sidecar"},
				Summary:            "Image not found",
				Details:            "The specified image does not exist",
				RemediationSteps:   []string{"Check image name", "Verify registry access"},
				ImageReferences: []types.ImageReference{
					{ContainerName: "nginx", FullReference: "nginx:invalid"},
				},
				IsTransient:      false,
				FailureCount:     5,
				FirstFailureTime: &now,
				LastFailureTime:  &now,
				FailureDuration:  "5m",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.finding.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					// Check if error message contains expected substring
					if !contains(err.Error(), tt.errorMsg) {
						t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errorMsg)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && s[:len(substr)] == substr) ||
		(len(s) > len(substr) && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
