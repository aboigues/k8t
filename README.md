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
- Multiple output formats: text (colored), JSON, YAML, XML, TOML

### CrashLoopBackOff Analyzer

Diagnoses root causes of container crashes and restart loops by analyzing:
- Container exit codes and termination status
- Application logs from crashed containers
- Kubernetes events and error patterns
- Resource limits and OOMKilled status
- Liveness/readiness probe configurations

**Capabilities:**
- Automatic log retrieval from previous container instances
- Pattern matching on logs and events for common crash scenarios
- Exit code analysis and interpretation
- Comprehensive remediation guidance
- Multiple output formats: text (colored), JSON, YAML, XML, TOML

**Root Causes Detected:**
- `OOM_KILLED` - Container exceeded memory limits
- `APPLICATION_ERROR` - Application crashed due to internal errors
- `CONFIG_ERROR` - Missing or invalid configuration
- `MISSING_DEPENDENCY` - Required service unavailable
- `PROBE_FAILURE` - Liveness/readiness probes failing
- `PERMISSION_ERROR` - Filesystem/security permissions issues
- `PORT_CONFLICT` - Port binding failures
- `EXIT_CODE_ERROR` - Non-zero exit codes

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

### Analyze ImagePullBackOff Issues

```bash
# As kubectl plugin
kubectl k8t analyze imagepullbackoff my-pod -n my-namespace

# Or as standalone binary
k8t analyze imagepullbackoff my-pod -n my-namespace

# JSON output for automation
kubectl k8t analyze imagepullbackoff my-pod -o json

# XML or TOML output
kubectl k8t analyze imagepullbackoff my-pod -o xml
kubectl k8t analyze imagepullbackoff my-pod -o toml
```

### Analyze CrashLoopBackOff Issues

```bash
# Analyze a crashing pod
kubectl k8t analyze crashloopbackoff my-pod -n my-namespace

# JSON output with full diagnostics
kubectl k8t analyze crashloopbackoff my-pod -o json

# Increase timeout for slow log retrieval
kubectl k8t analyze crashloopbackoff my-pod --timeout 60s
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
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get"]  # Required for CrashLoopBackOff analysis
```

## Output Formats

k8t supports multiple output formats for different use cases:

### Text (Default)

Human-readable colored output with remediation steps:

```bash
kubectl k8t analyze imagepullbackoff my-pod
kubectl k8t analyze crashloopbackoff my-pod
```

### JSON

Machine-readable format for automation and integration:

```bash
kubectl k8t analyze imagepullbackoff my-pod -o json
kubectl k8t analyze crashloopbackoff my-pod -o json
```

### YAML

YAML format for Kubernetes-native workflows:

```bash
kubectl k8t analyze imagepullbackoff my-pod -o yaml
kubectl k8t analyze crashloopbackoff my-pod -o yaml
```

### XML

XML format for enterprise systems integration:

```bash
kubectl k8t analyze imagepullbackoff my-pod -o xml
kubectl k8t analyze crashloopbackoff my-pod -o xml
```

### TOML

TOML format for configuration management:

```bash
kubectl k8t analyze imagepullbackoff my-pod -o toml
kubectl k8t analyze crashloopbackoff my-pod -o toml
```

## Root Causes Detected

### ImagePullBackOff Root Causes

- `IMAGE_NOT_FOUND` - Image does not exist in registry
- `AUTHENTICATION_FAILURE` - Invalid or missing image pull secrets
- `NETWORK_ISSUE` - DNS resolution, TCP connection, or HTTP errors
- `RATE_LIMIT_EXCEEDED` - Registry rate limiting (e.g., Docker Hub)
- `PERMISSION_DENIED` - Insufficient permissions to pull image
- `MANIFEST_ERROR` - Invalid image manifest or platform mismatch
- `TRANSIENT_FAILURE` - Temporary errors (less than 3 failures over 5 minutes)
- `UNKNOWN` - Unable to determine root cause

### CrashLoopBackOff Root Causes

- `OOM_KILLED` - Container exceeded memory limits
- `APPLICATION_ERROR` - Application crashed due to internal errors
- `CONFIG_ERROR` - Missing or invalid configuration
- `MISSING_DEPENDENCY` - Required service or resource unavailable
- `PROBE_FAILURE` - Liveness or readiness probe failing
- `PERMISSION_ERROR` - Filesystem or security context permissions issue
- `PORT_CONFLICT` - Port already in use or binding failed
- `EXIT_CODE_ERROR` - Container exited with non-zero exit code
- `TRANSIENT_FAILURE` - Temporary failures that may self-resolve
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
│   │   ├── analyzer.go   # ImagePullBackOff & CrashLoopBackOff analyzers
│   │   ├── detector.go   # Root cause detection with pattern matching
│   │   ├── remediation.go # Remediation steps for each root cause
│   │   └── events.go     # Event parsing and analysis
│   ├── k8s/              # Kubernetes API interactions
│   │   ├── client.go     # Kubernetes client setup
│   │   ├── pods.go       # Pod and container status retrieval
│   │   └── events.go     # Event filtering and conversion
│   ├── output/           # Output formatters
│   │   ├── formatter.go  # Format dispatcher
│   │   ├── text.go       # Colored text output
│   │   ├── json.go       # JSON output
│   │   ├── yaml.go       # YAML output
│   │   ├── xml.go        # XML output
│   │   └── toml.go       # TOML output
│   └── types/            # Shared data types
│       ├── finding.go    # Diagnostic finding structures
│       ├── report.go     # Analysis report structure
│       └── rootcause.go  # Root cause definitions
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

# Use the plugin for ImagePullBackOff
kubectl k8t analyze imagepullbackoff my-pod -n my-namespace

# Use the plugin for CrashLoopBackOff
kubectl k8t analyze crashloopbackoff my-pod -n my-namespace
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
