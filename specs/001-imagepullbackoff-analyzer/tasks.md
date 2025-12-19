# Tasks: ImagePullBackOff Analyzer

**Feature**: 001-imagepullbackoff-analyzer
**Input**: Design documents from `/specs/001-imagepullbackoff-analyzer/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), data-model.md, contracts/, quickstart.md

**Tests**: Tests are included based on Constitution requirement for reliability (all features MUST include automated tests).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `pkg/`, `cmd/`, `tests/` at repository root (k8t/)
- Paths shown below follow Go project layout from plan.md

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [ ] T001 Initialize Go module with github.com/aboigues/k8t
- [ ] T002 Create directory structure per plan.md (cmd/k8t, pkg/{analyzer,k8s,output,types}, tests/{unit,integration,contract})
- [ ] T003 [P] Install core dependencies: client-go v0.29+, cobra v1.8+, yaml.v3, color, zap
- [ ] T004 [P] Create Makefile with targets: build, test, test-unit, test-integration, lint, security, fmt
- [ ] T005 [P] Create .gitignore for Go project (binaries, vendor/, *.test, coverage.out)
- [ ] T006 [P] Create README.md with project description and quickstart instructions
- [ ] T007 [P] Setup .goreleaser.yml for multi-platform builds (Linux/macOS/Windows, amd64/arm64)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T008 [P] Define RootCause enum and Severity types in pkg/types/rootcause.go
- [ ] T009 [P] Define ImageReference struct in pkg/types/finding.go
- [ ] T010 [P] Define DiagnosticFinding struct in pkg/types/finding.go
- [ ] T011 [P] Define AnalysisReport struct in pkg/types/report.go
- [ ] T012 [P] Implement K8s client initialization from kubeconfig in pkg/k8s/client.go
- [ ] T013 [P] Implement input validation functions (namespace/pod name sanitization) in pkg/k8s/validation.go
- [ ] T014 [P] Create audit logger using zap for stdout/stderr in pkg/output/audit.go
- [ ] T015 [P] Write unit tests for RootCause enum and Severity mapping in tests/unit/rootcause_test.go
- [ ] T016 [P] Write unit tests for ImageReference parsing in tests/unit/finding_test.go
- [ ] T017 [P] Write unit tests for input validation in tests/unit/validation_test.go

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Basic Root Cause Identification (Priority: P1) 🎯 MVP

**Goal**: Single pod analysis with root cause identification and remediation steps

**Independent Test**: Create pod with invalid image, run `k8t analyze imagepullbackoff pod <name>`, verify root cause identified correctly with remediation steps

### Implementation for User Story 1

- [ ] T018 [P] [US1] Implement pod fetching by name/namespace in pkg/k8s/pods.go
- [ ] T019 [P] [US1] Implement event fetching filtered by pod in pkg/k8s/events.go
- [ ] T020 [US1] Implement event parsing to extract ImagePullBackOff errors in pkg/analyzer/events.go
- [ ] T021 [US1] Implement root cause detection logic with pattern matching (image not found, auth failure, network, rate limit) in pkg/analyzer/imagepull.go
- [ ] T022 [US1] Implement remediation step generation per root cause in pkg/analyzer/imagepull.go
- [ ] T023 [US1] Implement transient vs persistent failure detection (3+ failures over 5+ min) in pkg/analyzer/events.go
- [ ] T024 [P] [US1] Implement text output formatter with colored terminal output in pkg/output/table.go
- [ ] T025 [P] [US1] Implement JSON output formatter in pkg/output/json.go
- [ ] T026 [P] [US1] Implement YAML output formatter in pkg/output/yaml.go
- [ ] T027 [US1] Implement main analyzer orchestrator for single pod in pkg/analyzer/analyzer.go
- [ ] T028 [US1] Implement cobra CLI root command and analyze command in cmd/k8t/main.go
- [ ] T029 [US1] Implement cobra subcommand for single pod analysis with flags (--namespace, --output, --kubeconfig) in cmd/k8t/analyze.go
- [ ] T030 [US1] Wire up analyzer, K8s client, and output formatters in cmd/k8t/analyze.go
- [ ] T031 [US1] Implement error handling for pod not found, permission denied, API timeout in cmd/k8t/analyze.go
- [ ] T032 [US1] Add secret redaction for credentials in all output formats in pkg/output/redact.go

### Tests for User Story 1

- [ ] T033 [P] [US1] Write unit tests for event parsing (all root cause scenarios) in tests/unit/events_test.go
- [ ] T034 [P] [US1] Write unit tests for root cause detection logic in tests/unit/imagepull_test.go
- [ ] T035 [P] [US1] Write unit tests for transient vs persistent detection in tests/unit/events_test.go
- [ ] T036 [P] [US1] Write unit tests for remediation step generation in tests/unit/imagepull_test.go
- [ ] T037 [P] [US1] Write unit tests for text/JSON/YAML formatters in tests/unit/formatter_test.go
- [ ] T038 [P] [US1] Write unit tests for secret redaction in tests/unit/redact_test.go
- [ ] T039 [P] [US1] Create test pod manifests (image not found, auth failure, network issue, rate limit) in tests/integration/testdata/
- [ ] T040 [US1] Write integration test for image-not-found scenario with kind cluster in tests/integration/pod_not_found_test.go
- [ ] T041 [US1] Write integration test for auth-failure scenario with kind cluster in tests/integration/pod_auth_test.go
- [ ] T042 [US1] Write contract test for minimum RBAC permissions (pods/get, pods/list, events/list) in tests/contract/rbac_test.go

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently (MVP complete!)

---

## Phase 4: User Story 2 - Detailed Diagnostic Report (Priority: P2)

**Goal**: Enhanced diagnostics with network testing, event timeline, and detailed reports

**Independent Test**: Run analyzer with `--detailed` flag, verify report includes DNS/TCP/HTTP checks, event timeline, and detailed remediation

### Implementation for User Story 2

- [ ] T043 [P] [US2] Define NetworkDiagnostics, DNSResult, TCPResult, HTTPResult structs in pkg/types/finding.go
- [ ] T044 [P] [US2] Implement DNS resolution testing in pkg/analyzer/network.go
- [ ] T045 [P] [US2] Implement TCP connection testing to registry port in pkg/analyzer/network.go
- [ ] T046 [P] [US2] Implement HTTP HEAD request to registry in pkg/analyzer/network.go
- [ ] T047 [US2] Orchestrate DNS+TCP+HTTP tests for network issues in pkg/analyzer/network.go
- [ ] T048 [US2] Implement event timeline generation (first failure, last failure, frequency) in pkg/analyzer/events.go
- [ ] T049 [US2] Implement imagePullSecrets analysis (check existence, type, registry match) in pkg/analyzer/secrets.go
- [ ] T050 [US2] Add --detailed flag support to CLI in cmd/k8t/analyze.go
- [ ] T051 [US2] Extend analyzer to include network diagnostics when --detailed flag set in pkg/analyzer/analyzer.go
- [ ] T052 [US2] Extend output formatters to include network diagnostics and timeline in pkg/output/table.go, json.go, yaml.go

### Tests for User Story 2

- [ ] T053 [P] [US2] Write unit tests for DNS resolution testing in tests/unit/network_test.go
- [ ] T054 [P] [US2] Write unit tests for TCP connection testing in tests/unit/network_test.go
- [ ] T055 [P] [US2] Write unit tests for HTTP HEAD testing in tests/unit/network_test.go
- [ ] T056 [P] [US2] Write unit tests for event timeline generation in tests/unit/events_test.go
- [ ] T057 [P] [US2] Write unit tests for imagePullSecrets analysis in tests/unit/secrets_test.go
- [ ] T058 [US2] Write integration test for detailed report with network diagnostics in tests/integration/pod_network_detailed_test.go
- [ ] T059 [US2] Write integration test for detailed report with imagePullSecrets analysis in tests/integration/pod_secrets_detailed_test.go

**Checkpoint**: User Story 2 complete - detailed diagnostics now available

---

## Phase 5: User Story 3 - Multi-Pod Analysis (Priority: P3)

**Goal**: Analyze multiple pods (workload/namespace) with findings grouped by root cause

**Independent Test**: Create deployment with 5 replicas using invalid image, run analyzer on deployment, verify common root cause identified across all pods

### Implementation for User Story 3

- [ ] T060 [P] [US3] Implement workload pod listing (deployment, statefulset, daemonset, replicaset) in pkg/k8s/pods.go
- [ ] T061 [P] [US3] Implement namespace-wide pod listing with ImagePullBackOff filter in pkg/k8s/pods.go
- [ ] T062 [US3] Implement finding aggregation logic (group by root cause) in pkg/analyzer/aggregator.go
- [ ] T063 [US3] Implement ReportSummary generation (totals, breakdowns, affected pods count) in pkg/types/report.go
- [ ] T064 [US3] Extend analyzer to handle multiple pods in pkg/analyzer/analyzer.go
- [ ] T065 [US3] Implement cobra subcommand for workload analysis (deployment/statefulset/daemonset) in cmd/k8t/analyze.go
- [ ] T066 [US3] Implement cobra subcommand for namespace analysis in cmd/k8t/analyze.go
- [ ] T067 [US3] Add --issues-only and --max-pods flags to CLI in cmd/k8t/analyze.go
- [ ] T068 [US3] Extend output formatters to display multi-pod results grouped by root cause in pkg/output/table.go
- [ ] T069 [US3] Extend JSON/YAML formatters for multi-pod reports in pkg/output/json.go, yaml.go

### Tests for User Story 3

- [ ] T070 [P] [US3] Write unit tests for workload pod listing in tests/unit/pods_test.go
- [ ] T071 [P] [US3] Write unit tests for namespace pod listing in tests/unit/pods_test.go
- [ ] T072 [P] [US3] Write unit tests for finding aggregation (grouping by root cause) in tests/unit/aggregator_test.go
- [ ] T073 [P] [US3] Write unit tests for ReportSummary generation in tests/unit/report_test.go
- [ ] T074 [P] [US3] Create test deployment manifest (5 replicas, invalid image) in tests/integration/testdata/
- [ ] T075 [US3] Write integration test for deployment analysis in tests/integration/deployment_test.go
- [ ] T076 [US3] Write integration test for namespace analysis in tests/integration/namespace_test.go

**Checkpoint**: All user stories complete - tool now handles single pods, detailed reports, and multi-pod analysis

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories

- [ ] T077 [P] Add comprehensive error messages for common failure scenarios (RBAC, API timeout, pod not found) in pkg/k8s/errors.go
- [ ] T078 [P] Implement --verbose flag for audit trail output to stderr in cmd/k8t/root.go
- [ ] T079 [P] Implement --quiet flag to suppress non-essential output in cmd/k8t/root.go
- [ ] T080 [P] Implement --timeout flag with default 30s in cmd/k8t/root.go
- [ ] T081 [P] Implement --no-color flag to disable ANSI colors in pkg/output/table.go
- [ ] T082 [P] Add shell completion generation (bash, zsh, fish, powershell) in cmd/k8t/completion.go
- [ ] T083 [P] Implement proper exit codes (0=success, 1=usage error, 2=API error, 3=not found, 4=timeout) in cmd/k8t/main.go
- [ ] T084 [P] Add input validation error messages with remediation suggestions in pkg/k8s/validation.go
- [ ] T085 [P] Create GitHub Actions CI workflow (.github/workflows/ci.yml) with unit tests, lint, security scan
- [ ] T086 [P] Add kind-based integration tests to CI workflow in .github/workflows/ci.yml
- [ ] T087 [P] Run gosec security scanner in CI in .github/workflows/ci.yml
- [ ] T088 [P] Run govulncheck vulnerability scanner in CI in .github/workflows/ci.yml
- [ ] T089 [P] Write comprehensive README.md with installation, usage examples, and RBAC requirements
- [ ] T090 [P] Add performance benchmarks for analyzer (ensure <10s P95 for single pod) in tests/unit/benchmark_test.go
- [ ] T091 [P] Add code coverage reporting to CI (target 80%+ for critical paths) in .github/workflows/ci.yml
- [ ] T092 [P] Document all CLI commands and flags in README.md or docs/

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3, 4, 5)**: All depend on Foundational phase completion
  - User stories can proceed in priority order: US1 (P1) → US2 (P2) → US3 (P3)
  - Or in parallel if team capacity allows (after Foundational phase)
- **Polish (Phase 6)**: Depends on desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1 - MVP)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Extends US1 but independently testable
- **User Story 3 (P3)**: Can start after Foundational (Phase 2) - Extends US1 but independently testable

### Within Each User Story

- Types before logic (types → K8s client → analyzer → output → CLI)
- Tests can run in parallel with implementation (TDD encouraged but not required per constitution)
- Unit tests marked [P] can run in parallel
- Integration tests run sequentially (need kind cluster state)

### Parallel Opportunities

- **Setup tasks**: T003-T007 (dependencies, configs, docs) can run in parallel
- **Foundational types**: T008-T011 (different type files) can run in parallel
- **Foundational K8s**: T012-T014 (client, validation, audit) can run in parallel
- **Foundational tests**: T015-T017 can run in parallel
- **User Story 1 implementation**: T018-T019 (pods/events fetching), T024-T026 (formatters) can run in parallel
- **User Story 1 tests**: T033-T038 (unit tests), T039 (test data) can run in parallel
- **User Story 2 implementation**: T043 (types), T044-T046 (network tests) can run in parallel
- **User Story 2 tests**: T053-T057 (unit tests) can run in parallel
- **User Story 3 implementation**: T060-T061 (pod listing) can run in parallel
- **User Story 3 tests**: T070-T073 (unit tests), T074 (test data) can run in parallel
- **Polish tasks**: T077-T084, T089-T092 (most polish tasks) can run in parallel

---

## Parallel Example: User Story 1 (MVP)

```bash
# Launch implementation tasks in parallel:
Task T018 [P] [US1]: Implement pod fetching in pkg/k8s/pods.go
Task T019 [P] [US1]: Implement event fetching in pkg/k8s/events.go
Task T024 [P] [US1]: Implement text formatter in pkg/output/table.go
Task T025 [P] [US1]: Implement JSON formatter in pkg/output/json.go
Task T026 [P] [US1]: Implement YAML formatter in pkg/output/yaml.go

# Then run sequential tasks:
Task T020 [US1]: Event parsing (depends on T019)
Task T021 [US1]: Root cause detection (depends on T020)
...

# Launch test tasks in parallel:
Task T033 [P] [US1]: Unit test event parsing
Task T034 [P] [US1]: Unit test root cause detection
Task T035 [P] [US1]: Unit test transient detection
Task T036 [P] [US1]: Unit test remediation steps
Task T037 [P] [US1]: Unit test formatters
Task T038 [P] [US1]: Unit test redaction
Task T039 [P] [US1]: Create test manifests
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T007)
2. Complete Phase 2: Foundational (T008-T017) - CRITICAL - blocks all stories
3. Complete Phase 3: User Story 1 (T018-T042)
4. **STOP and VALIDATE**: Test User Story 1 independently
   - `k8t analyze imagepullbackoff pod test-pod`
   - Verify root cause identification works
   - Verify all output formats (text/JSON/YAML)
   - Verify RBAC permissions minimal
5. Deploy/demo MVP if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (T018-T042) → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 (T043-T059) → Test independently → Deploy/Demo (detailed reports)
4. Add User Story 3 (T060-T076) → Test independently → Deploy/Demo (multi-pod analysis)
5. Add Polish (T077-T092) → Final release-ready version

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together (T001-T017)
2. Once Foundational is done:
   - Developer A: User Story 1 (T018-T042)
   - Developer B: User Story 2 (T043-T059) - can start in parallel
   - Developer C: User Story 3 (T060-T076) - can start in parallel
3. Stories complete and integrate independently
4. Team tackles Polish together (T077-T092)

---

## Notes

- **[P] tasks** = different files, no dependencies, can run in parallel
- **[Story] label** maps task to specific user story for traceability
- Each user story is independently completable and testable
- Tests included per Constitution requirement (reliability mandate)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- **Constitution compliance**: Security (RBAC, redaction, audit), Reliability (tests, error handling), Diagnostic Excellence (accuracy, actionable output)

---

## Task Summary

**Total Tasks**: 92

**By Phase**:
- Phase 1 (Setup): 7 tasks
- Phase 2 (Foundational): 10 tasks
- Phase 3 (US1 - MVP): 25 tasks (15 implementation + 10 tests)
- Phase 4 (US2): 17 tasks (10 implementation + 7 tests)
- Phase 5 (US3): 17 tasks (10 implementation + 7 tests)
- Phase 6 (Polish): 16 tasks

**By User Story**:
- User Story 1 (P1 - MVP): 25 tasks
- User Story 2 (P2): 17 tasks
- User Story 3 (P3): 17 tasks
- Infrastructure (Setup + Foundational + Polish): 33 tasks

**Parallel Opportunities**: 45 tasks marked [P] can run in parallel within their phase

**Independent Test Criteria**:
- US1: Create pod with invalid image → run analyzer → verify root cause + remediation
- US2: Run with --detailed → verify network checks + timeline in output
- US3: Create deployment (5 replicas, invalid image) → analyze → verify grouped findings

**MVP Scope**: Phase 1 + Phase 2 + Phase 3 = 42 tasks (Setup + Foundational + User Story 1)
