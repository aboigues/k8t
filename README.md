# k8t - Kubernetes Administration Toolkit

A collection of diagnostic tools to help Kubernetes administrators troubleshoot common issues.

## Features

### ImagePullBackOff Analyzer

Identifies root causes of ImagePullBackOff errors in pods by analyzing:
- Pod events and error patterns
- Image pull secrets and authentication
- Registry connectivity (DNS, TCP, HTTP)
- Container image specifications

**Capabilities:**
- Single pod analysis with actionable remediation steps
- Detailed diagnostics with network testing and event timeline
- Multi-pod analysis for workloads and namespaces
- Multiple output formats: text (colored), JSON, YAML

## Installation

### From Source

```bash
git clone https://github.com/yourorg/k8t.git
cd k8t
make build
sudo cp bin/k8t /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourorg/k8t/cmd/k8t@latest
```

## Quick Start

### Analyze a Single Pod

```bash
# Basic analysis
k8t analyze imagepullbackoff pod my-pod -n my-namespace

# Detailed analysis with network diagnostics
k8t analyze imagepullbackoff pod my-pod -n my-namespace --detailed

# JSON output for automation
k8t analyze imagepullbackoff pod my-pod -o json
```

### Analyze Multiple Pods

```bash
# Analyze all pods in a namespace
k8t analyze imagepullbackoff namespace my-namespace

# Analyze a deployment
k8t analyze imagepullbackoff deployment my-deployment -n my-namespace

# Show only pods with issues
k8t analyze imagepullbackoff namespace my-namespace --issues-only
```

## RBAC Requirements

The tool requires the following Kubernetes permissions:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: k8t-analyzer
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["list"]
```

## Output Formats

### Text (Default)

Human-readable colored output with remediation steps.

### JSON

Machine-readable format for automation and integration:

```bash
k8t analyze imagepullbackoff pod my-pod -o json
```

### YAML

YAML format for Kubernetes-native workflows:

```bash
k8t analyze imagepullbackoff pod my-pod -o yaml
```

## Root Causes Detected

- `IMAGE_NOT_FOUND` - Image does not exist in registry
- `AUTHENTICATION_FAILURE` - Invalid or missing image pull secrets
- `NETWORK_ISSUE` - DNS resolution, TCP connection, or HTTP errors
- `RATE_LIMIT_EXCEEDED` - Registry rate limiting (e.g., Docker Hub)
- `PERMISSION_DENIED` - Insufficient permissions to pull image
- `MANIFEST_ERROR` - Invalid image manifest or platform mismatch
- `TRANSIENT_FAILURE` - Temporary errors (less than 3 failures over 5 minutes)
- `UNKNOWN` - Unable to determine root cause

## Development

### Prerequisites

- Go 1.21 or later
- kubectl configured with cluster access
- kind (for integration tests)

### Build

```bash
make build
```

### Test

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run integration tests (requires kind)
make test-integration
```

### Lint and Security

```bash
# Run linters
make lint

# Run security scanners
make security
```

### CI Checks

```bash
# Run all CI checks (format, vet, lint, security, test)
make ci
```

## Architecture

```
k8t/
├── cmd/k8t/              # CLI entry point
├── pkg/
│   ├── analyzer/         # Core diagnostic logic
│   ├── k8s/              # Kubernetes API interactions
│   ├── output/           # Output formatters (text/JSON/YAML)
│   └── types/            # Shared data types
└── tests/
    ├── unit/             # Unit tests
    ├── integration/      # Integration tests (kind)
    └── contract/         # API contract tests
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development guidelines.

## License

Apache 2.0 - See [LICENSE](LICENSE) for details.

## Security

- All cluster access is read-only
- Credentials are handled securely via kubeconfig
- Secrets and sensitive data are redacted from output
- Audit trail logged to stdout/stderr

Report security vulnerabilities to security@yourorg.com
