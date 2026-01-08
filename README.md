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

### As a kubectl Plugin (Recommended)

Install k8t as a kubectl plugin using krew:

```bash
kubectl krew install k8t
```

Or manually install the plugin:

```bash
git clone https://github.com/aboigues/k8t.git
cd k8t
make install-plugin
```

### Standalone Binary

#### From Source

```bash
git clone https://github.com/aboigues/k8t.git
cd k8t
make build
sudo cp bin/k8t /usr/local/bin/
```

#### Using Go Install

```bash
go install github.com/aboigues/k8t/cmd/k8t@latest
```

## Quick Start

k8t can be used as a kubectl plugin or as a standalone binary. Both provide identical functionality.

### Analyze a Single Pod

```bash
# As kubectl plugin
kubectl k8t analyze imagepullbackoff my-pod -n my-namespace

# Or as standalone binary
k8t analyze imagepullbackoff my-pod -n my-namespace

# JSON output for automation
kubectl k8t analyze imagepullbackoff my-pod -o json
```

### Analyze Multiple Pods

```bash
# Analyze all pods in a namespace
kubectl k8t analyze imagepullbackoff namespace my-namespace

# Analyze a deployment
kubectl k8t analyze imagepullbackoff deployment my-deployment -n my-namespace

# Show only pods with issues
kubectl k8t analyze imagepullbackoff namespace my-namespace --issues-only

# Shorthand: analyze all pods with issues
kubectl k8t aa -A
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
kubectl k8t analyze imagepullbackoff my-pod -o json
```

### YAML

YAML format for Kubernetes-native workflows:

```bash
kubectl k8t analyze imagepullbackoff my-pod -o yaml
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

## kubectl Plugin Integration

k8t is available as a kubectl plugin, providing seamless integration with your existing kubectl workflows.

### How it Works

When installed as a kubectl plugin, the binary is named `kubectl-k8t`. kubectl automatically discovers and executes plugins from your PATH when you run `kubectl k8t`.

### Installation via Krew

[Krew](https://krew.sigs.k8s.io/) is the recommended plugin manager for kubectl:

```bash
# Install krew (if not already installed)
# See: https://krew.sigs.k8s.io/docs/user-guide/setup/install/

# Install k8t plugin
kubectl krew install k8t

# Use the plugin
kubectl k8t analyze imagepullbackoff my-pod -n my-namespace
```

### Manual Plugin Installation

```bash
# Build the plugin binary
make build-plugin

# Install to PATH
make install-plugin

# Verify installation
kubectl plugin list | grep k8t
```

### Compatibility

The kubectl plugin is fully compatible with the standalone binary. All commands, flags, and features work identically in both modes.

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
