# Quickstart: ImagePullBackOff Analyzer Development

**Feature**: 001-imagepullbackoff-analyzer
**Date**: 2025-12-18
**Phase**: Phase 1 - Developer Guide

## Overview

This quickstart guide helps developers get started implementing the ImagePullBackOff Analyzer. It covers setup, development workflow, testing, and contribution guidelines.

## Prerequisites

### Required Tools

- **Go 1.21+**: [Download](https://go.dev/dl/)
- **kubectl**: Kubernetes command-line tool
- **Docker**: For running kind clusters
- **kind v0.20+**: Kubernetes IN Docker for testing
- **make**: Build automation (optional but recommended)

### Kubernetes Access

You need access to a Kubernetes cluster for development and testing:

**Option 1: kind (Recommended for local development)**
```bash
# Install kind
GO111MODULE="on" go install sigs.k8s.io/kind@v0.20.0

# Create test cluster
kind create cluster --name k8t-dev

# Verify cluster
kubectl cluster-info --context kind-k8t-dev
```

**Option 2: Existing Cluster**
```bash
# Verify access
kubectl get pods -A

# Ensure you have required RBAC permissions
kubectl auth can-i get pods
kubectl auth can-i list events
```

## Project Setup

### 1. Clone Repository

```bash
git clone https://github.com/yourorg/k8t.git
cd k8t
```

### 2. Initialize Go Module

```bash
go mod init github.com/yourorg/k8t
go mod tidy
```

### 3. Install Dependencies

```bash
# Core dependencies
go get k8s.io/client-go@v0.29.0
go get k8s.io/api@v0.29.0
go get k8s.io/apimachinery@v0.29.0
go get github.com/spf13/cobra@v1.8.0
go get github.com/spf13/viper@v1.18.2
go get gopkg.in/yaml.v3@v3.0.1
go get github.com/fatih/color@v1.16.0
go get go.uber.org/zap@v1.26.0

# Test dependencies
go get github.com/stretchr/testify@v1.8.4
go get sigs.k8s.io/kind@v0.20.0

# Development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### 4. Verify Setup

```bash
# Build project
go build -o k8t ./cmd/k8t

# Run basic command
./k8t --version

# Run tests
go test ./...
```

## Project Structure

```
k8t/
├── cmd/
│   └── k8t/
│       └── main.go                 # Start here: CLI entry point
├── pkg/
│   ├── analyzer/                   # Core analysis logic
│   │   ├── analyzer.go            # Main orchestrator
│   │   ├── events.go              # Event parsing
│   │   ├── imagepull.go           # Root cause detection
│   │   ├── network.go             # Network diagnostics
│   │   └── secrets.go             # ImagePullSecrets analysis
│   ├── k8s/                        # Kubernetes API wrappers
│   │   ├── client.go              # K8s client initialization
│   │   ├── pods.go                # Pod fetching
│   │   └── events.go              # Event fetching
│   ├── output/                     # Output formatting
│   │   ├── formatter.go           # Format dispatcher
│   │   ├── json.go                # JSON marshaling
│   │   ├── yaml.go                # YAML marshaling
│   │   └── table.go               # Human-readable tables
│   └── types/                      # Data models
│       ├── finding.go             # DiagnosticFinding
│       ├── rootcause.go           # RootCause enum
│       └── report.go              # AnalysisReport
├── tests/
│   ├── unit/                       # Unit tests (fast, mocked)
│   ├── integration/                # kind-based integration tests
│   └── contract/                   # K8s API contract tests
├── go.mod                          # Go dependencies
├── Makefile                        # Build targets
└── README.md                       # Project documentation
```

## Development Workflow

### Phase 1: User Story 1 (P1) - MVP

**Goal**: Basic root cause identification for single pod

**Implementation Order**:

1. **Types (pkg/types/)** - Start here
   ```bash
   # Define data structures first
   vim pkg/types/rootcause.go
   vim pkg/types/finding.go
   vim pkg/types/report.go

   # Write tests
   vim pkg/types/rootcause_test.go

   # Run tests
   go test ./pkg/types/...
   ```

2. **K8s Client Wrapper (pkg/k8s/)**
   ```bash
   # Implement K8s API interactions
   vim pkg/k8s/client.go
   vim pkg/k8s/pods.go
   vim pkg/k8s/events.go

   # Write unit tests with mocked client
   vim pkg/k8s/pods_test.go

   # Run tests
   go test ./pkg/k8s/...
   ```

3. **Event Analysis (pkg/analyzer/events.go)**
   ```bash
   # Parse K8s events to extract image pull errors
   vim pkg/analyzer/events.go

   # Test with sample event data
   vim tests/unit/events_test.go

   # Run tests
   go test ./tests/unit/...
   ```

4. **Root Cause Detection (pkg/analyzer/imagepull.go)**
   ```bash
   # Implement root cause categorization logic
   vim pkg/analyzer/imagepull.go

   # Test all root cause scenarios
   vim tests/unit/imagepull_test.go

   # Run tests
   go test ./tests/unit/...
   ```

5. **Main Analyzer (pkg/analyzer/analyzer.go)**
   ```bash
   # Orchestrate analysis workflow
   vim pkg/analyzer/analyzer.go

   # Integration test
   vim tests/integration/analyzer_test.go

   # Run integration tests (requires kind cluster)
   make test-integration
   ```

6. **Output Formatting (pkg/output/)**
   ```bash
   # Implement text/JSON/YAML formatters
   vim pkg/output/table.go
   vim pkg/output/json.go
   vim pkg/output/yaml.go

   # Test output formats
   vim pkg/output/formatter_test.go

   # Run tests
   go test ./pkg/output/...
   ```

7. **CLI (cmd/k8t/main.go)**
   ```bash
   # Wire up cobra commands
   vim cmd/k8t/main.go
   vim cmd/k8t/analyze.go

   # Build and test
   go build -o k8t ./cmd/k8t
   ./k8t analyze imagepullbackoff pod test-pod
   ```

### Quick Commands

```bash
# Build
make build

# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests (requires kind)
make test-integration

# Lint code
make lint

# Security scan
make security

# Format code
make fmt

# Generate mocks
make mocks
```

## Testing Guide

### Unit Tests

**Fast tests with mocked dependencies**

```go
// Example: tests/unit/events_test.go
package unit

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/yourorg/k8t/pkg/analyzer"
)

func TestParseImagePullEvent(t *testing.T) {
	event := &v1.Event{
		Reason: "Failed",
		Message: "Failed to pull image \"nginx:invalid\": rpc error: code = Unknown desc = failed to pull and unpack image",
	}

	rootCause := analyzer.ParseImagePullEvent(event)
	assert.Equal(t, types.RootCauseImageNotFound, rootCause)
}
```

**Run unit tests**:
```bash
go test ./tests/unit/... -v
```

### Integration Tests

**Tests against real K8s API (kind cluster)**

```go
// Example: tests/integration/analyzer_test.go
package integration

import (
	"context"
	"testing"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/k8t/pkg/analyzer"
	"github.com/yourorg/k8t/pkg/k8s"
)

func TestAnalyzePodWithImageNotFound(t *testing.T) {
	// Setup kind cluster with test pod
	ctx := context.Background()
	client := k8s.NewClient(t)

	// Create test pod with invalid image
	pod := createTestPod(t, client, "test-pod", "invalid-image:v999")

	// Run analyzer
	analyzer := analyzer.New(client)
	report, err := analyzer.AnalyzePod(ctx, "test-pod", "default")

	// Assertions
	require.NoError(t, err)
	require.Len(t, report.Findings, 1)
	require.Equal(t, types.RootCauseImageNotFound, report.Findings[0].RootCause)
}
```

**Run integration tests**:
```bash
# Start kind cluster
kind create cluster --name k8t-test

# Run tests
go test ./tests/integration/... -v

# Cleanup
kind delete cluster --name k8t-test
```

### Contract Tests

**Validate K8s API interactions and RBAC**

```go
// Example: tests/contract/rbac_test.go
package contract

import (
	"context"
	"testing"
	"github.com/stretchr/testify/require"
	authv1 "k8s.io/api/authorization/v1"
)

func TestMinimumRBACPermissions(t *testing.T) {
	ctx := context.Background()
	client := k8s.NewClient(t)

	// Test required permissions from SR-001
	permissions := []struct {
		resource string
		verb     string
	}{
		{"pods", "get"},
		{"pods", "list"},
		{"events", "list"},
	}

	for _, perm := range permissions {
		sar := &authv1.SelfSubjectAccessReview{
			Spec: authv1.SelfSubjectAccessReviewSpec{
				ResourceAttributes: &authv1.ResourceAttributes{
					Namespace: "default",
					Verb:      perm.verb,
					Resource:  perm.resource,
				},
			},
		}

		result, err := client.AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, sar, metav1.CreateOptions{})
		require.NoError(t, err)
		require.True(t, result.Status.Allowed, "Missing permission: %s %s", perm.verb, perm.resource)
	}
}
```

### Test Data

Create test manifests in `tests/integration/testdata/`:

```yaml
# tests/integration/testdata/pod-image-not-found.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-image-not-found
  namespace: default
spec:
  containers:
  - name: app
    image: nonexistent-registry.io/app:v999
```

```yaml
# tests/integration/testdata/pod-auth-failure.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-auth-failure
  namespace: default
spec:
  containers:
  - name: app
    image: gcr.io/private-project/private-image:v1
  imagePullSecrets:
  - name: nonexistent-secret
```

## Makefile Targets

Create `Makefile` in repository root:

```makefile
.PHONY: build test test-unit test-integration lint security fmt clean

# Build binary
build:
	go build -o k8t ./cmd/k8t

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	go test ./pkg/... ./tests/unit/... -v -cover

# Run integration tests (requires kind cluster)
test-integration:
	@echo "Checking for kind cluster..."
	@kind get clusters | grep -q k8t-test || kind create cluster --name k8t-test
	go test ./tests/integration/... -v
	@echo "Keeping kind cluster running. To delete: kind delete cluster --name k8t-test"

# Run contract tests
test-contract:
	go test ./tests/contract/... -v

# Lint code
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

# Security scan
security:
	@which gosec > /dev/null || go install github.com/securego/gosec/v2/cmd/gosec@latest
	gosec ./...
	@which govulncheck > /dev/null || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# Format code
fmt:
	gofmt -s -w .
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .

# Clean build artifacts
clean:
	rm -f k8t
	go clean

# Install locally
install:
	go install ./cmd/k8t

# Generate mocks (if using mockgen)
mocks:
	@which mockgen > /dev/null || go install github.com/golang/mock/mockgen@latest
	mockgen -source=pkg/k8s/client.go -destination=pkg/k8s/mock_client.go -package=k8s
```

## Debugging Tips

### Enable Verbose Logging

```bash
# Set log level
export K8T_LOG_LEVEL=debug

# Run with verbose flag
./k8t analyze imagepullbackoff pod my-pod --verbose
```

### Debug K8s API Calls

```bash
# Enable kubectl proxy
kubectl proxy --port=8080 &

# Point k8t to proxy
export KUBERNETES_SERVICE_HOST=localhost
export KUBERNETES_SERVICE_PORT=8080

# Run analyzer (API calls visible in proxy logs)
./k8t analyze imagepullbackoff pod my-pod
```

### Test with Sample Pods

```bash
# Create pod with image not found
kubectl run test-not-found --image=nonexistent:v999

# Wait for ImagePullBackOff
kubectl wait --for=condition=Ready=false pod/test-not-found --timeout=60s

# Run analyzer
./k8t analyze imagepullbackoff pod test-not-found

# Cleanup
kubectl delete pod test-not-found
```

## Code Style Guidelines

### Go Conventions

- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` and `goimports` for formatting
- Run `golangci-lint` before committing
- Write godoc comments for exported functions

### Error Handling

```go
// Good: Wrap errors with context
if err != nil {
	return fmt.Errorf("failed to fetch pod %s/%s: %w", namespace, name, err)
}

// Good: Check specific error types
if errors.Is(err, k8serrors.NotFound) {
	return ErrPodNotFound
}

// Bad: Silent error
if err != nil {
	return nil // Don't ignore errors
}
```

### Testing Conventions

```go
// Test function names: Test<Function><Scenario>
func TestAnalyzePod_ImageNotFound(t *testing.T) { ... }
func TestAnalyzePod_AuthFailure(t *testing.T) { ... }

// Use table-driven tests for multiple scenarios
func TestParseImageReference(t *testing.T) {
	tests := []struct {
		name      string
		imageRef  string
		want      *types.ImageReference
		wantErr   bool
	}{
		{
			name:     "simple image",
			imageRef: "nginx",
			want:     &types.ImageReference{Registry: "docker.io", Repository: "library/nginx", Tag: "latest"},
		},
		// ... more cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseImageReference(tt.imageRef)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

## CI/CD Integration

### GitHub Actions Example

`.github/workflows/ci.yml`:

```yaml
name: CI

on:
  push:
    branches: [ main, 001-imagepullbackoff-analyzer ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run unit tests
        run: make test-unit

      - name: Run lint
        run: make lint

      - name: Run security scan
        run: make security

      - name: Setup kind
        uses: helm/kind-action@v1
        with:
          version: v0.20.0

      - name: Run integration tests
        run: make test-integration

      - name: Run contract tests
        run: make test-contract

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
```

## Performance Profiling

### CPU Profiling

```go
// Add to main.go
import "runtime/pprof"

if cpuprofile := os.Getenv("CPUPROFILE"); cpuprofile != "" {
	f, _ := os.Create(cpuprofile)
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
}
```

```bash
# Run with profiling
CPUPROFILE=cpu.prof ./k8t analyze imagepullbackoff pod my-pod

# Analyze profile
go tool pprof cpu.prof
(pprof) top10
(pprof) web
```

### Memory Profiling

```bash
# Run with memory profiling
go test -memprofile=mem.prof ./pkg/analyzer/...

# Analyze profile
go tool pprof mem.prof
(pprof) top10
```

## Common Issues & Solutions

### Issue: "connection refused" when accessing cluster

**Solution**: Verify kubeconfig and cluster connectivity
```bash
kubectl cluster-info
kubectl get nodes
```

### Issue: "permission denied" errors

**Solution**: Check RBAC permissions
```bash
kubectl auth can-i get pods
kubectl auth can-i list events
```

### Issue: Integration tests fail with "cluster not found"

**Solution**: Ensure kind cluster is running
```bash
kind get clusters
kind create cluster --name k8t-test
```

### Issue: Module dependencies conflict

**Solution**: Update and tidy modules
```bash
go get -u ./...
go mod tidy
```

## Next Steps

### Phase 2: User Story 2 (P2) - Detailed Reports

1. Implement `--detailed` flag
2. Add network diagnostics (DNS/TCP/HTTP)
3. Include event timeline in output
4. Test with various network scenarios

### Phase 3: User Story 3 (P3) - Multi-Pod Analysis

1. Implement workload analysis (deployment, statefulset, daemonset)
2. Implement namespace-wide analysis
3. Group findings by root cause
4. Optimize for performance with many pods

### Beyond MVP

- kubectl plugin integration
- Prometheus metrics export
- Historical trend analysis
- Auto-remediation capabilities

## Resources

- **Kubernetes client-go**: https://github.com/kubernetes/client-go
- **kind documentation**: https://kind.sigs.k8s.io/
- **Cobra guide**: https://github.com/spf13/cobra
- **Go project layout**: https://github.com/golang-standards/project-layout
- **Effective Go**: https://go.dev/doc/effective_go
- **k8t Constitution**: `.specify/memory/constitution.md`
- **Feature Spec**: `specs/001-imagepullbackoff-analyzer/spec.md`

## Getting Help

- Review the specification: `specs/001-imagepullbackoff-analyzer/spec.md`
- Check the data model: `specs/001-imagepullbackoff-analyzer/data-model.md`
- Read CLI contract: `specs/001-imagepullbackoff-analyzer/contracts/cli-interface.md`
- Review constitution: `.specify/memory/constitution.md`

## Contributing

1. Follow the constitution principles (security, reliability, simplicity)
2. Write tests before implementation (TDD encouraged but not required per constitution)
3. Run `make lint` and `make security` before committing
4. Update documentation for any API changes
5. Reference spec requirements in commit messages

Happy coding! 🚀
