# Research: ImagePullBackOff Analyzer Technology Stack

**Date**: 2025-12-18
**Feature**: 001-imagepullbackoff-analyzer
**Phase**: Phase 0 - Technology Selection

## Research Questions

1. **Language/Runtime**: Best language for Kubernetes diagnostic CLI tool
2. **K8s Client Library**: Which Kubernetes client library to use
3. **Testing Framework**: Testing approach for K8s integration testing
4. **Network Libraries**: DNS resolution, TCP connection, HTTP client needs

## Decision Summary

| Component | Decision | Rationale |
|-----------|----------|-----------|
| **Language** | Go 1.21+ | K8s ecosystem standard, excellent performance, mature client-go library |
| **K8s Client** | client-go v0.29+ | Official K8s Go client, full API coverage, well-maintained |
| **Testing** | Go testing + testify + kind | Built-in testing, testify for assertions, kind for integration tests |
| **DNS/Network** | net (stdlib), net/http (stdlib) | Standard library sufficient, no external deps needed |
| **Output Formats** | encoding/json, gopkg.in/yaml.v3, text/tabwriter | JSON/YAML standard libs, tabwriter for tables |
| **CLI Framework** | cobra + viper | Industry standard for K8s tools (kubectl, helm, etc.) |
| **Build** | Go modules + goreleaser | Native Go tooling, cross-platform binaries |

## 1. Language Selection

### Decision: Go 1.21+

**Rationale**:
- **Ecosystem Fit**: Go is the de facto standard for Kubernetes tools (kubectl, k9s, stern, kubectx)
- **client-go Maturity**: Official Kubernetes Go client with complete API coverage and active maintenance
- **Performance**: Compiled binary, low memory footprint (<50MB), fast startup (<100ms)
- **Cross-Platform**: Native cross-compilation to Linux/macOS/Windows from single codebase
- **Security**: Strong typing, memory safety, good security tooling (gosec, govulncheck)
- **Distribution**: Single static binary, no runtime dependencies
- **Community**: Largest K8s tooling ecosystem, extensive examples and libraries

**Alternatives Considered**:

| Alternative | Pros | Cons | Why Rejected |
|-------------|------|------|--------------|
| Python 3.11+ | Rapid development, familiar syntax | Slower startup (~500ms), packaging complexity, larger memory footprint | Performance requirements (<10s P95) and distribution complexity |
| Rust 1.75+ | Maximum performance, memory safety | Steeper learning curve, smaller K8s ecosystem (kube-rs less mature) | Development velocity vs marginal performance gains not justified |

### Supporting Data:
- kubectl, helm, k9s, stern, kubectx all written in Go
- client-go has 11k+ GitHub stars, 3k+ contributors
- Go 1.21+ includes improved performance and security features
- Cross-compilation: `GOOS=linux GOARCH=amd64 go build` (trivial)

## 2. Kubernetes Client Library

### Decision: client-go v0.29+ (matching K8s 1.29+)

**Rationale**:
- **Official Client**: Maintained by Kubernetes project, guaranteed API compatibility
- **Complete Coverage**: Full access to all K8s APIs (core, apps, batch, etc.)
- **Versioning**: client-go version tracks K8s version (v0.29 = K8s 1.29)
- **Type Safety**: Strongly-typed Go structs for all K8s resources
- **Clientset**: Easy-to-use clientset interface for common operations
- **Dynamic Client**: Support for CRDs and dynamic resource types
- **Informers/Listers**: Built-in caching and watching (not needed for this tool but available)

**Key Libraries**:
```go
k8s.io/client-go/kubernetes        // Main clientset
k8s.io/client-go/tools/clientcmd   // Kubeconfig parsing
k8s.io/api/core/v1                 // Core types (Pod, Event, Secret)
k8s.io/apimachinery/pkg/apis/meta/v1 // Meta types (ObjectMeta, ListOptions)
```

**Alternatives Considered**:
- **controller-runtime**: Overkill for simple diagnostic tool, adds complexity
- **kubectl libraries**: Tied to kubectl releases, more dependencies than needed

## 3. Testing Framework

### Decision: Go testing + testify + kind

**Rationale**:

**Unit Testing**: Go's built-in `testing` package + `testify/assert` and `testify/mock`
- Standard Go testing approach
- testify provides better assertions and mocking
- No external test runner needed

**Integration Testing**: kind (Kubernetes IN Docker)
- Lightweight K8s cluster in Docker (starts in ~30s)
- Perfect for CI/CD pipelines
- Test against real K8s API, not mocks
- Supports multiple K8s versions for compatibility testing

**Contract Testing**: Custom test helpers with client-go
- Test K8s API interactions with real API server (kind)
- Validate RBAC permissions
- Test error scenarios (API failures, timeouts)

**Libraries**:
```go
testing                              // Go stdlib
github.com/stretchr/testify/assert  // Better assertions
github.com/stretchr/testify/mock    // Mocking
sigs.k8s.io/kind                    // K8s in Docker (integration tests)
```

**Test Structure**:
```
tests/
├── unit/           # Fast unit tests, mocked dependencies
├── integration/    # kind-based integration tests
└── contract/       # K8s API contract tests
```

**Alternatives Considered**:
- **ginkgo/gomega**: More verbose, BDD style not needed for this project
- **minikube**: Heavier than kind, slower startup
- **k3s**: Good alternative to kind but kind is more popular in CI

## 4. Network Libraries

### Decision: Go standard library (net, net/http)

**Rationale**:

**DNS Resolution**: `net` package
```go
net.LookupHost(hostname)  // DNS resolution
net.LookupIP(hostname)    // IP addresses
```

**TCP Connection Testing**:
```go
net.DialTimeout("tcp", "registry.io:443", 5*time.Second)
```

**HTTP HEAD Requests**:
```go
http.Head(url)  // HTTP HEAD request
// Or custom client with timeout
client := &http.Client{Timeout: 10 * time.Second}
client.Head(url)
```

**Why Standard Library**:
- No external dependencies
- Battle-tested, secure
- Sufficient for requirements (DNS + TCP + HTTP HEAD)
- Smaller binary size
- Security updates via Go releases

**No Need For**:
- HTTP framework (no server needed)
- Advanced HTTP client (simple HEAD requests)
- DNS library (stdlib sufficient)

## 5. Output Formatting

### Decision: encoding/json + gopkg.in/yaml.v3 + text/tabwriter

**Rationale**:

**JSON Output**: `encoding/json` (stdlib)
```go
json.MarshalIndent(result, "", "  ")  // Pretty JSON
```

**YAML Output**: `gopkg.in/yaml.v3`
```go
yaml.Marshal(result)  // YAML output
```

**Human-Readable Tables**: `text/tabwriter` (stdlib)
```go
w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
fmt.Fprintf(w, "CONTAINER\tIMAGE\tSTATUS\tREASON\n")
```

**Colored Terminal Output**: `github.com/fatih/color`
```go
color.Red("ERROR: ")
color.Green("✓ ")
```

**Libraries**:
```go
encoding/json                // JSON (stdlib)
gopkg.in/yaml.v3            // YAML
text/tabwriter              // Tables (stdlib)
github.com/fatih/color      // Terminal colors
```

**Why These Choices**:
- JSON/tabwriter in stdlib (no deps)
- yaml.v3 is K8s ecosystem standard
- color is lightweight, widely used
- Matches kubectl output style

## 6. CLI Framework

### Decision: cobra + viper

**Rationale**:

**cobra**: CLI framework
- Used by kubectl, helm, k9s, etc.
- Excellent flag parsing
- Subcommand support
- Auto-generated help
- Shell completion

**viper**: Configuration management
- Works with cobra
- Reads config files, env vars, flags
- Useful for kubeconfig paths, defaults

**Example CLI Structure**:
```bash
k8t analyze pod my-pod                    # Single pod
k8t analyze pod my-pod --output json      # JSON output
k8t analyze deployment my-app             # All pods in deployment
k8t analyze namespace my-ns               # All pods in namespace
k8t analyze pod my-pod --detailed         # P2: Detailed report
```

**Libraries**:
```go
github.com/spf13/cobra    // CLI framework
github.com/spf13/viper    // Configuration
```

**Alternatives Considered**:
- **urfave/cli**: Less feature-rich than cobra
- **flag (stdlib)**: Too basic for complex CLI
- **cobra is K8s ecosystem standard**: Proven choice

## 7. Build & Distribution

### Decision: Go modules + goreleaser

**Rationale**:

**Dependency Management**: Go modules (built-in since Go 1.11)
```bash
go mod init github.com/aboigues/k8t
go mod tidy
```

**Building**:
```bash
go build -o k8t ./cmd/k8t
```

**Cross-Compilation**:
```bash
GOOS=linux GOARCH=amd64 go build -o k8t-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o k8t-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o k8t-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o k8t-windows-amd64.exe
```

**Distribution**: goreleaser
- Automates multi-platform builds
- Creates GitHub releases
- Generates checksums
- Supports Homebrew taps
- Docker images (optional)

**goreleaser.yml**:
```yaml
builds:
  - binary: k8t
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
```

**Alternatives Considered**:
- **Makefile**: Manual, but works; goreleaser is better for releases
- **Docker-only**: Doesn't meet "single binary" requirement

## 8. Additional Dependencies

### Recommended Libraries:

**Error Handling**:
```go
github.com/pkg/errors  // Error wrapping with stack traces
```

**Logging**:
```go
go.uber.org/zap  // Structured logging for audit trail
```

**RBAC Checking** (optional):
```go
k8s.io/client-go/kubernetes/typed/authorization/v1  // SelfSubjectAccessReview
```

## Implementation Architecture

### Project Structure:
```
k8t/
├── cmd/
│   └── k8t/
│       └── main.go                 # CLI entry point
├── pkg/
│   ├── analyzer/
│   │   ├── analyzer.go            # Main analysis logic
│   │   ├── events.go              # Event parsing
│   │   ├── imagepull.go           # Image pull diagnostics
│   │   ├── network.go             # DNS/TCP/HTTP checks
│   │   └── secrets.go             # ImagePullSecrets analysis
│   ├── k8s/
│   │   ├── client.go              # K8s client wrapper
│   │   └── resources.go           # Resource fetching
│   ├── output/
│   │   ├── formatter.go           # Output formatting
│   │   ├── json.go                # JSON output
│   │   ├── yaml.go                # YAML output
│   │   └── table.go               # Human-readable tables
│   └── types/
│       ├── finding.go             # Diagnostic finding types
│       └── rootcause.go           # Root cause categories
├── tests/
│   ├── unit/                      # Unit tests
│   ├── integration/               # kind-based integration tests
│   └── contract/                  # K8s API contract tests
├── go.mod
├── go.sum
├── Makefile
├── .goreleaser.yml
└── README.md
```

### Core Dependencies (go.mod):
```go
module github.com/aboigues/k8t

go 1.21

require (
    k8s.io/client-go v0.29.0
    k8s.io/api v0.29.0
    k8s.io/apimachinery v0.29.0
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    gopkg.in/yaml.v3 v3.0.1
    github.com/fatih/color v1.16.0
    github.com/stretchr/testify v1.8.4
    go.uber.org/zap v1.26.0
    sigs.k8s.io/kind v0.20.0  // dev/test only
)
```

## Performance Considerations

**Startup Time**: <100ms (Go binary)
**Memory Footprint**: 20-30MB typical, <50MB max
**Binary Size**: 15-25MB (statically linked)
**Analysis Time**:
- DNS lookup: <100ms
- TCP connect: <1s
- HTTP HEAD: <2s
- K8s API calls: <1s each
- Total for single pod: <5s typical, <10s P95

## Security Considerations

**Dependency Security**:
- Use `go mod tidy` to manage dependencies
- Run `govulncheck` in CI for vulnerability scanning
- Run `gosec` for static analysis
- Keep dependencies up to date

**Binary Security**:
- Static binary reduces attack surface
- No dynamic library dependencies
- CGO_ENABLED=0 for fully static builds

**Credential Handling**:
- client-go handles kubeconfig parsing securely
- No credential storage in tool
- Respects KUBECONFIG environment variable

## CI/CD Integration

**GitHub Actions Example**:
```yaml
- uses: actions/setup-go@v4
  with:
    go-version: '1.21'
- run: go test ./...
- run: gosec ./...
- run: govulncheck ./...
- uses: helm/kind-action@v1  # Integration tests
- run: go test ./tests/integration/...
```

## Migration Path

**Phase 0 (MVP - P1)**:
- Core analyzer logic
- Single pod analysis
- Basic output formats
- Essential tests

**Phase 1 (P2)**:
- Detailed reporting
- Enhanced diagnostics

**Phase 2 (P3)**:
- Multi-pod analysis
- Batch operations

**Future Enhancements** (post-MVP):
- kubectl plugin integration (`kubectl analyze imagepullbackoff`)
- JSON schema for structured output
- Prometheus metrics export
- Web UI (optional)

## References

- [client-go documentation](https://github.com/kubernetes/client-go)
- [kind quick start](https://kind.sigs.k8s.io/docs/user/quick-start/)
- [cobra user guide](https://github.com/spf13/cobra)
- [Go project layout](https://github.com/golang-standards/project-layout)
- [Kubernetes API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2025-12-18 | Go 1.21+ selected | K8s ecosystem standard, performance, client-go maturity |
| 2025-12-18 | client-go v0.29+ | Official K8s client, full API coverage |
| 2025-12-18 | kind for integration tests | Lightweight, fast, CI-friendly |
| 2025-12-18 | cobra + viper for CLI | K8s ecosystem standard (kubectl, helm) |
| 2025-12-18 | Standard lib for network | Sufficient, no external deps needed |
| 2025-12-18 | goreleaser for distribution | Automates multi-platform releases |
