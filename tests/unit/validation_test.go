package unit

import (
	"strings"
	"testing"

	"github.com/yourorg/k8t/pkg/k8s"
)

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name        string
		namespace   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid namespace",
			namespace:   "default",
			expectError: false,
		},
		{
			name:        "Valid namespace with hyphen",
			namespace:   "my-namespace",
			expectError: false,
		},
		{
			name:        "Valid namespace with numbers",
			namespace:   "namespace-123",
			expectError: false,
		},
		{
			name:        "Empty namespace",
			namespace:   "",
			expectError: true,
			errorMsg:    "namespace cannot be empty",
		},
		{
			name:        "Namespace with uppercase",
			namespace:   "MyNamespace",
			expectError: true,
			errorMsg:    "must consist of lowercase",
		},
		{
			name:        "Namespace with slash (injection)",
			namespace:   "default/../../etc/passwd",
			expectError: true,
			errorMsg:    "invalid characters or injection patterns",
		},
		{
			name:        "Namespace with semicolon (injection)",
			namespace:   "default; rm -rf /",
			expectError: true,
			errorMsg:    "invalid characters or injection patterns",
		},
		{
			name:        "Namespace with dollar sign (injection)",
			namespace:   "default$variable",
			expectError: true,
			errorMsg:    "invalid characters or injection patterns",
		},
		{
			name:        "Namespace with backtick (injection)",
			namespace:   "default`whoami`",
			expectError: true,
			errorMsg:    "invalid characters or injection patterns",
		},
		{
			name:        "Namespace starting with hyphen",
			namespace:   "-invalid",
			expectError: true,
			errorMsg:    "must start and end with an alphanumeric",
		},
		{
			name:        "Namespace ending with hyphen",
			namespace:   "invalid-",
			expectError: true,
			errorMsg:    "must start and end with an alphanumeric",
		},
		{
			name:        "Namespace too long",
			namespace:   strings.Repeat("a", 254),
			expectError: true,
			errorMsg:    "exceeds maximum length",
		},
		{
			name:        "Valid namespace at max length",
			namespace:   strings.Repeat("a", 253),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k8s.ValidateNamespace(tt.namespace)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidatePodName(t *testing.T) {
	tests := []struct {
		name        string
		podName     string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid pod name",
			podName:     "my-pod",
			expectError: false,
		},
		{
			name:        "Valid pod name with numbers",
			podName:     "pod-123",
			expectError: false,
		},
		{
			name:        "Valid pod name with dots (subdomain)",
			podName:     "my-pod.example.com",
			expectError: false,
		},
		{
			name:        "Empty pod name",
			podName:     "",
			expectError: true,
			errorMsg:    "pod name cannot be empty",
		},
		{
			name:        "Pod name with uppercase",
			podName:     "MyPod",
			expectError: true,
			errorMsg:    "invalid",
		},
		{
			name:        "Pod name with slash (injection)",
			podName:     "pod/../../../etc/passwd",
			expectError: true,
			errorMsg:    "invalid characters or injection patterns",
		},
		{
			name:        "Pod name with pipe (injection)",
			podName:     "pod | cat /etc/passwd",
			expectError: true,
			errorMsg:    "invalid characters or injection patterns",
		},
		{
			name:        "Pod name starting with hyphen",
			podName:     "-invalid-pod",
			expectError: true,
			errorMsg:    "invalid",
		},
		{
			name:        "Pod name ending with hyphen",
			podName:     "invalid-pod-",
			expectError: true,
			errorMsg:    "invalid",
		},
		{
			name:        "Pod name too long",
			podName:     strings.Repeat("a", 254),
			expectError: true,
			errorMsg:    "exceeds maximum length",
		},
		{
			name:        "Valid generated pod name (deployment)",
			podName:     "nginx-deployment-66b6c48dd5-abcde",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k8s.ValidatePodName(tt.podName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Error message = %v, want to contain %v", err.Error(), tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateWorkloadName(t *testing.T) {
	tests := []struct {
		name         string
		workloadName string
		expectError  bool
	}{
		{
			name:         "Valid deployment name",
			workloadName: "my-deployment",
			expectError:  false,
		},
		{
			name:         "Valid statefulset name",
			workloadName: "postgres-statefulset",
			expectError:  false,
		},
		{
			name:         "Empty workload name",
			workloadName: "",
			expectError:  true,
		},
		{
			name:         "Workload with injection",
			workloadName: "deploy; kubectl delete all",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := k8s.ValidateWorkloadName(tt.workloadName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestInjectionPatternsDetection(t *testing.T) {
	injectionTests := []struct {
		name  string
		input string
	}{
		{"Path traversal", "../../../etc/passwd"},
		{"Command chaining", "pod; rm -rf /"},
		{"Pipe", "pod | cat /etc/passwd"},
		{"Background execution", "pod & rm -rf /"},
		{"Input redirection", "pod < /etc/passwd"},
		{"Output redirection", "pod > /tmp/output"},
		{"Variable expansion", "pod$HOME"},
		{"Command substitution", "pod`whoami`"},
		{"Wildcards", "pod*"},
		{"Brace expansion", "pod{1,2,3}"},
		{"Subshell", "pod(test)"},
		{"Quotes", "pod'test'"},
		{"Double quotes", "pod\"test\""},
		{"Newline", "pod\ntest"},
		{"Null byte", "pod\x00test"},
	}

	for _, tt := range injectionTests {
		t.Run(tt.name, func(t *testing.T) {
			err := k8s.ValidateNamespace(tt.input)
			if err == nil {
				t.Errorf("Expected injection pattern '%s' to be detected in namespace validation", tt.name)
			}

			err = k8s.ValidatePodName(tt.input)
			if err == nil {
				t.Errorf("Expected injection pattern '%s' to be detected in pod name validation", tt.name)
			}
		})
	}
}

func TestSanitizeInput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No whitespace",
			input:    "test",
			expected: "test",
		},
		{
			name:     "Leading whitespace",
			input:    "  test",
			expected: "test",
		},
		{
			name:     "Trailing whitespace",
			input:    "test  ",
			expected: "test",
		},
		{
			name:     "Both sides whitespace",
			input:    "  test  ",
			expected: "test",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only whitespace",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := k8s.SanitizeInput(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeInput() = %v, want %v", result, tt.expected)
			}
		})
	}
}
