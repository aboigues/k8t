# Implementation Plan: ImagePullBackOff Analyzer

**Branch**: `001-imagepullbackoff-analyzer` | **Date**: 2025-12-18 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-imagepullbackoff-analyzer/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

The ImagePullBackOff Analyzer is a Kubernetes diagnostic CLI tool that identifies root causes of ImagePullBackOff errors in pods by analyzing pod events, imagePullSecrets, registry connectivity (DNS + TCP + HTTP HEAD), and container specifications. The tool provides actionable remediation steps and supports multiple output formats (text, JSON, YAML). MVP (P1) focuses on single-pod root cause identification for common failures (image not found, authentication, network, rate limiting).

## Technical Context

**Language/Version**: Go 1.21+ (K8s ecosystem standard, excellent performance, mature tooling)
**Primary Dependencies**:
- k8s.io/client-go v0.29+ (official K8s Go client)
- github.com/spf13/cobra v1.8+ (CLI framework, used by kubectl/helm)
- gopkg.in/yaml.v3 (YAML output)
- github.com/fatih/color (terminal colors)
- go.uber.org/zap (structured logging for audit trail)

**Storage**: N/A - Stateless diagnostic tool, no persistent storage required
**Testing**:
- Go testing + testify (unit tests with assertions/mocking)
- kind v0.20+ (Kubernetes IN Docker for integration tests)
- gosec + govulncheck (security scanning in CI)

**Target Platform**: Linux/macOS/Windows - Cross-platform CLI tool, single static binary via goreleaser
**Project Type**: Single project - CLI application with library core
**Performance Goals**: <10s single pod analysis (P95), <30s for multi-pod (P95), >90% root cause accuracy
**Constraints**: <50MB memory footprint, read-only cluster access, minimum RBAC permissions (pods/get, pods/list, events/list)
**Scale/Scope**: Single cluster analysis, handles 1-1000 pods per invocation, supports K8s versions n/n-1/n-2

**Rationale**: See [research.md](./research.md) for detailed technology selection analysis. Go chosen for K8s ecosystem fit, client-go maturity, performance (<100ms startup), and single binary distribution.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### I. Security-First Development ✅ PASS

- ✅ Principle of least privilege: RBAC permissions explicitly defined (SR-001: pods/get, pods/list, events/list)
- ✅ Secure credential handling: Uses standard kubeconfig, no plaintext storage (Assumption #1)
- ✅ RBAC documented: Minimum permissions defined in SR-001
- ✅ Audit trail: SR-004 requires logging all cluster access to stdout/stderr
- ✅ Input validation: SR-005 requires validation to prevent injection attacks
- ✅ Secret redaction: SR-003, SR-007 require redaction of secrets from all output
- ✅ CI security scans: SC-010, SC-011 require SAST and dependency scanning
- ✅ CIS K8s Benchmark: Applicable standards will be followed

### II. Reliability & CI Testing ✅ PASS

- ✅ Automated tests required: Constitution mandates unit, integration, contract tests
- ✅ CI validation: Security and reliability tests must run on every commit
- ✅ Error handling: RR-001, RR-002 define comprehensive error handling requirements
- ✅ Edge case handling: 6 edge cases identified in spec, RR-004 handles pod deletion
- ✅ Predictable performance: RR-003 defines <30s P95 for single pod, SC-004 defines <10s for 95% cases
- ✅ Independent user stories: 3 prioritized stories (P1=MVP, P2=detailed reports, P3=multi-pod)
- ✅ Failure scenario tests: SC-013 requires tests for API errors, permissions, timeouts
- ✅ Regression suite: Constitution requires regression tests before merge

### III. Diagnostic Excellence ✅ PASS

- ✅ Accurate root cause: RR-005 requires >90% accuracy, no false positives
- ✅ Actionable output: FR-002 requires remediation steps, SC-003 requires 100% actionable output
- ✅ Performance: RR-003 and SC-004 define specific performance targets
- ✅ Simplicity (YAGNI): Scope clearly bounded, Out of Scope section prevents feature creep
- ✅ Avoid abstractions: Single project structure, no premature abstractions planned
- ✅ Clear output: FR-007 defines multiple output formats (text/JSON/YAML)
- ✅ Documentation: Tool purpose, resources examined, RBAC, and examples in spec

### Kubernetes-Specific Requirements ✅ PASS

- ✅ Kubeconfig authentication: Assumption #1 confirms kubectl access
- ✅ RBAC documented: SR-001 defines minimum permissions
- ✅ Read-only operations: SR-006 requires read-only mode
- ✅ Namespace isolation: SR-001 scoped to target namespace
- ✅ No state modification: SR-006 explicitly forbids cluster modifications
- ✅ Rate limiting: Constitution requires graceful handling (to be implemented in RR-002)
- ✅ Secret redaction: SR-003, SR-007 mandate redaction
- ✅ Safe to share: SR-007 requires diagnostic output safe to share externally
- ✅ Version support: RR-006 requires support for K8s versions n, n-1, n-2
- ✅ Deprecation handling: Constitution requires graceful API version deprecation handling

### Quality Gates Status

**Specification Gate**: ✅ PASSED
- No implementation details in spec
- All requirements testable (15 FR + 7 SR + 6 RR)
- Security requirements explicit (RBAC, credentials, audit)
- Reliability requirements measurable (error handling, performance, edge cases)
- Success criteria include security and reliability metrics (SC-009 through SC-013)
- Zero [NEEDS CLARIFICATION] markers remain after clarification session

**Planning Gate Requirements** (checked post-Phase 1):
- Technical context: Language/dependencies need resolution in Phase 0
- Project structure: Will be defined below
- Security analysis: RBAC, audit trail, credential handling documented
- Reliability considerations: Error scenarios, performance limits defined
- Complexity violations: None anticipated (single project, simple architecture)

**Overall Status**: ✅ READY TO PROCEED - No constitution violations detected

## Project Structure

### Documentation (this feature)

```text
specs/001-imagepullbackoff-analyzer/
├── spec.md              # Feature specification
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (technology decisions)
├── data-model.md        # Phase 1 output (data structures)
├── quickstart.md        # Phase 1 output (developer guide)
├── contracts/           # Phase 1 output (CLI contract spec)
│   └── cli-interface.md # CLI commands and output schemas
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
k8t/                                # Repository root
├── cmd/
│   └── k8t/
│       └── main.go                 # CLI entry point, cobra root command
├── pkg/
│   ├── analyzer/
│   │   ├── analyzer.go            # Main analysis orchestrator
│   │   ├── events.go              # Pod event parsing and analysis
│   │   ├── imagepull.go           # ImagePullBackOff root cause logic
│   │   ├── network.go             # DNS/TCP/HTTP connectivity checks
│   │   └── secrets.go             # ImagePullSecrets validation
│   ├── k8s/
│   │   ├── client.go              # K8s client wrapper with error handling
│   │   ├── pods.go                # Pod fetching and filtering
│   │   └── events.go              # Event fetching and filtering
│   ├── output/
│   │   ├── formatter.go           # Output format dispatcher
│   │   ├── json.go                # JSON marshaling
│   │   ├── yaml.go                # YAML marshaling
│   │   └── table.go               # Human-readable table formatting
│   └── types/
│       ├── finding.go             # DiagnosticFinding type
│       ├── rootcause.go           # RootCause enum and categories
│       └── report.go              # AnalysisReport structure
├── tests/
│   ├── unit/                      # Unit tests with mocked K8s client
│   │   ├── analyzer_test.go
│   │   ├── events_test.go
│   │   └── network_test.go
│   ├── integration/               # kind-based integration tests
│   │   ├── e2e_test.go           # End-to-end scenarios
│   │   └── testdata/             # Test pod manifests
│   └── contract/                  # K8s API contract tests
│       └── rbac_test.go          # RBAC permission validation
├── go.mod                         # Go module definition
├── go.sum                         # Dependency checksums
├── Makefile                       # Build, test, lint targets
├── .goreleaser.yml                # Multi-platform release config
├── .github/
│   └── workflows/
│       └── ci.yml                 # GitHub Actions CI pipeline
└── README.md                      # Project documentation
```

**Structure Decision**: Single Go project with clear separation of concerns:
- `cmd/k8t`: CLI entry point and command definitions (cobra)
- `pkg/analyzer`: Core diagnostic logic (stateless, testable)
- `pkg/k8s`: Kubernetes API interactions (client-go wrappers)
- `pkg/output`: Multi-format output generation (text/JSON/YAML)
- `pkg/types`: Shared data types and domain models
- `tests/`: Three-tier testing (unit, integration, contract)

This structure follows Go project layout conventions and enables:
- Clear dependency boundaries (analyzer depends on k8s, not vice versa)
- Easy unit testing (mock K8s client in pkg/k8s)
- Reusable core logic (pkg/analyzer can be imported as library)
- Simple CLI composition (cmd/k8t orchestrates pkg components)

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

**No violations detected** - Project follows constitution principles:
- Single project structure (no unnecessary complexity)
- No premature abstractions
- Simple, focused architecture
- YAGNI principles applied throughout
