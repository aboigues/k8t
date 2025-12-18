<!--
SYNC IMPACT REPORT
==================
Version: 1.0.0 → 1.1.0
Rationale: MINOR bump - Added domain-specific principles and requirements for Kubernetes diagnostic tooling

Modified Principles:
- UPDATED: I. Security-First Development (was Specification-First) - K8s admin tools require security priority
- UPDATED: II. Reliability & CI Testing (expanded from Independent Testing) - Added CI requirements for security/reliability
- UPDATED: III. Diagnostic Excellence (was Simplicity & YAGNI) - Added diagnostic-specific requirements
- KEPT: Simplicity principles integrated into Diagnostic Excellence

Added Sections:
- Kubernetes-Specific Requirements (RBAC, credentials, least privilege)
- CI/CD Requirements (security scanning, reliability tests)
- Diagnostic Tool Standards (accuracy, actionability, performance)

Removed Sections:
- None

Template Consistency Status:
- ✅ .specify/templates/spec-template.md - Success criteria now include security and reliability metrics
- ✅ .specify/templates/plan-template.md - Constitution Check validates security and K8s requirements
- ✅ .specify/templates/tasks-template.md - Tasks must include security and reliability tests
- ✅ .claude/commands/*.md - No changes needed

Follow-up TODOs:
- Consider documenting specific CI pipeline requirements in separate runbook
- Define CIS Kubernetes Benchmark test suite integration
-->

# k8t Constitution

**Project**: k8t - Kubernetes Diagnostic & Troubleshooting Toolkit
**Purpose**: Provide reliable, secure tools to diagnose and resolve Kubernetes cluster issues

## Core Principles

### I. Security-First Development

Security is NON-NEGOTIABLE for Kubernetes administration tools. Every feature MUST be designed, implemented, and tested with security as the primary concern.

**Requirements**:
- All features MUST follow the principle of least privilege
- Cluster credentials MUST be handled securely (no plaintext storage, use kubeconfig best practices)
- RBAC permissions MUST be explicitly documented for each diagnostic tool
- All cluster access MUST be auditable (log what data was accessed, when, by whom)
- Input validation MUST prevent injection attacks (command injection, YAML injection, etc.)
- Secrets and sensitive data MUST be redacted from logs and diagnostic output
- CI MUST run security scans on every commit (SAST, dependency scanning, container scanning)
- Follow CIS Kubernetes Benchmark security standards where applicable

**Rationale**: K8s diagnostic tools access sensitive cluster data, configurations, and logs. A security vulnerability could expose production secrets, compromise clusters, or enable privilege escalation. Security must be built-in from the start, not added later.

### II. Reliability & CI Testing

K8t MUST be rock-solid reliable. Administrators depend on diagnostic tools during incidents and production issues - failures are unacceptable.

**Requirements**:
- All features MUST include automated tests (unit, integration, contract)
- CI pipeline MUST validate both security AND reliability on every commit
- Error handling MUST be comprehensive and provide clear, actionable error messages
- Diagnostic tools MUST handle edge cases gracefully (empty clusters, permission denied, network timeouts)
- Performance MUST be predictable (no unbounded operations, resource limits documented)
- Each user story MUST be independently testable and deliverable as standalone functionality
- User stories MUST be prioritized (P1 = MVP, P2/P3 = enhancements)
- Tests MUST cover failure scenarios (API errors, malformed data, partial permissions)
- Regression test suite MUST run in CI before any merge

**Rationale**: Unreliable diagnostic tools create more problems than they solve. During incidents, administrators need tools that work correctly every time. Comprehensive CI testing catches issues before production deployment.

### III. Diagnostic Excellence

Diagnostic tools MUST provide accurate, actionable insights with minimal complexity. Focus on solving real Kubernetes problems effectively.

**Requirements**:
- Root cause analysis MUST be accurate (avoid false positives that waste admin time)
- Diagnostic output MUST be actionable (tell users what's wrong AND how to fix it)
- Tools MUST be performant (admins use these during incidents - speed matters)
- Keep implementations simple (YAGNI - build what's needed, not speculative features)
- Avoid premature abstractions (three similar functions better than complex abstraction)
- Output MUST be clear and well-formatted (human-readable by default, JSON for scripting)
- Each diagnostic tool MUST document:
  - What problem it solves
  - What K8s resources it examines
  - What RBAC permissions it requires
  - Example usage and expected output

**Diagnostic Tool Coverage** (current scope):
1. **imagePullBackOff Analyzer**: Root cause analysis for image pull failures
2. **Log Deep Analyzer**: Advanced log aggregation and anomaly detection
3. **Configuration Anomaly Detector**: Find misconfigurations (ports not opened, etc.)
4. **Load Distribution Analyzer**: Identify inappropriate load balancing configurations

**Rationale**: Diagnostic tools must solve specific problems accurately and quickly. Complexity slows development and creates maintenance burden. Simple, focused tools are easier to test, debug, and trust during high-pressure incidents.

## Kubernetes-Specific Requirements

### Cluster Access & Security

- MUST use standard kubeconfig authentication (support multiple contexts)
- MUST document minimum required RBAC permissions for each tool
- MUST support read-only operations by default (write operations require explicit flag)
- MUST respect namespace isolation (no cluster-admin required unless explicitly necessary)
- SHOULD support RBAC impersonation for testing permission scenarios
- MUST never modify cluster state without explicit user confirmation
- MUST handle API rate limiting gracefully

### Data Handling

- Secrets, tokens, and passwords MUST be redacted from all output
- Sensitive ConfigMap/Secret data MUST not be displayed without explicit flag
- Diagnostic reports MUST be safe to share (no accidental credential leaks)
- Log analysis MUST handle PII appropriately (redaction options)

### Compatibility

- MUST support currently maintained Kubernetes versions (n, n-1, n-2)
- MUST document version-specific features or limitations
- MUST handle API version deprecations gracefully

## Development Standards

### Specification-First Approach

While security and reliability are paramount, features still begin with specifications:
- Specifications define WHAT problem the diagnostic tool solves and WHY
- Specifications MUST be technology-agnostic (no implementation details)
- Specifications MUST include security and reliability success criteria
- Specifications MUST define measurable outcomes (accuracy, performance, usability)

### Scope Control

- Only implement requested diagnostic features (no speculative tools)
- No "improvements" or refactoring beyond feature scope
- Document assumptions rather than building excessive configurability
- Delete unused code completely (no commented-out code)

### Code Quality

- Self-evident code preferred over heavily commented code
- Add comments only for non-obvious logic (especially security-critical sections)
- Security-critical code MUST have explanatory comments
- Error messages MUST be helpful (explain what failed and suggest resolution)
- OWASP Top 10 vulnerabilities MUST be prevented (injection, broken auth, sensitive data exposure)

### Version Control

- Commit messages MUST reference feature specifications
- Branch naming: `[###-feature-short-name]` where ### is feature number
- Each feature lives in `specs/[###-feature-name]/` with design artifacts

## Quality Gates

### Specification Gate

Before planning begins, specifications MUST pass:
- No implementation details present
- All functional requirements testable
- Security requirements explicitly defined (RBAC, data handling, audit trail)
- Reliability requirements measurable (error handling, performance, edge cases)
- Success criteria include security and reliability metrics
- Maximum 3 [NEEDS CLARIFICATION] markers (must be resolved before planning)

### Planning Gate

Before implementation begins, plans MUST include:
- Technical context (language, K8s client library, testing framework)
- Project structure (concrete directory layout)
- Constitution compliance check
- Security analysis (RBAC requirements, data access patterns, credential handling)
- Reliability considerations (error scenarios, performance limits, testing strategy)
- Identified complexity violations with justification if any exist

### Implementation Gate

Before feature is considered complete:
- All tasks from tasks.md completed
- All user stories independently tested
- Security tests passed in CI (SAST, dependency scan, RBAC validation)
- Reliability tests passed in CI (unit, integration, error scenarios)
- Specification success criteria validated (functionality, security, reliability)
- Manual security review completed for sensitive operations
- Documentation updated (RBAC requirements, usage examples)
- No unused code remains

## CI/CD Requirements

### Security Pipeline (MANDATORY)

CI MUST run on every commit:
- Static Application Security Testing (SAST)
- Dependency vulnerability scanning
- Container image scanning (if applicable)
- Secrets detection (no credentials in code)
- License compliance check

### Reliability Pipeline (MANDATORY)

CI MUST run on every commit:
- Unit tests (minimum 80% coverage for critical paths)
- Integration tests with live K8s cluster (kind/k3s)
- Contract tests for K8s API interactions
- Error scenario tests (API failures, permission denied, timeouts)
- Performance tests (ensure tools complete within reasonable time)
- Regression test suite

### Quality Gates

- All CI checks MUST pass before merge (no exceptions)
- Security vulnerabilities MUST be fixed before merge (no high/critical vulns)
- Test coverage MUST not decrease
- Performance regressions MUST be justified and documented

## Governance

### Amendment Process

Constitution changes require:
1. Documentation of proposed change with rationale
2. Impact analysis on existing templates, workflows, and CI pipeline
3. Version increment following semantic versioning:
   - MAJOR: Backward-incompatible governance changes or principle removals/redefinitions
   - MINOR: New principle/section added or materially expanded guidance
   - PATCH: Clarifications, wording fixes, non-semantic refinements
4. Update to LAST_AMENDED_DATE
5. Sync Impact Report prepended as HTML comment
6. Propagation of changes to all dependent templates and command files

### Compliance Review

- All pull requests MUST verify compliance with constitution principles
- Planning phase MUST include Constitution Check section
- Security review REQUIRED for features accessing cluster credentials or secrets
- Any principle violations MUST be explicitly justified in Complexity Tracking table
- Unjustified complexity or security violations are grounds for rejection

### Runtime Guidance

For agent-specific runtime development guidance (as opposed to this project-wide governance), teams MAY create separate guidance files referenced in workflows but not embedded in this constitution.

**Version**: 1.1.0 | **Ratified**: 2025-12-18 | **Last Amended**: 2025-12-18
