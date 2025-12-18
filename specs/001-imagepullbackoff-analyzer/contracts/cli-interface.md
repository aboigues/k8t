# CLI Interface Contract: ImagePullBackOff Analyzer

**Feature**: 001-imagepullbackoff-analyzer
**Date**: 2025-12-18
**Phase**: Phase 1 - Contract Design

## Overview

This document defines the command-line interface contract for the k8t ImagePullBackOff analyzer. It specifies commands, flags, output formats, and exit codes.

## Command Structure

### Root Command

```bash
k8t analyze imagepullbackoff [TARGET] [flags]
```

**Aliases**:
```bash
k8t analyze ipbo [TARGET] [flags]      # Short form
k8t analyze image-pull [TARGET] [flags] # Alternative
```

---

## User Story 1 (P1): Basic Root Cause Identification

### Command: Analyze Single Pod

**Syntax**:
```bash
k8t analyze imagepullbackoff pod <POD_NAME> [flags]
```

**Example Usage**:
```bash
# Analyze single pod in default namespace
k8t analyze imagepullbackoff pod my-app-pod

# Analyze pod in specific namespace
k8t analyze imagepullbackoff pod my-app-pod -n production

# JSON output for automation
k8t analyze imagepullbackoff pod my-app-pod -o json

# YAML output
k8t analyze imagepullbackoff pod my-app-pod -o yaml
```

**Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--namespace` | `-n` | string | `default` | Target namespace |
| `--output` | `-o` | string | `text` | Output format: `text`, `json`, `yaml` (clarification Q1) |
| `--kubeconfig` | | string | `$KUBECONFIG` or `~/.kube/config` | Path to kubeconfig file |
| `--context` | | string | current-context | Kubernetes context to use |

**Output (Text Format)**:
```
ImagePullBackOff Analysis: my-app-pod (namespace: default)
================================================================================

ROOT CAUSE: Image does not exist in registry
SEVERITY: HIGH
AFFECTED CONTAINERS: app-container

SUMMARY:
The image 'docker.io/myorg/myapp:v2.0.0' does not exist in the registry.

DETAILS:
  Image Reference: docker.io/myorg/myapp:v2.0.0
  Registry: docker.io
  Repository: myorg/myapp
  Tag: v2.0.0

  Failure Timeline:
    First Failure: 2025-12-18 10:15:23 UTC
    Last Failure: 2025-12-18 10:20:45 UTC
    Failure Count: 5
    Duration: 5m22s
    Status: PERSISTENT (≥3 failures over ≥5 minutes)

REMEDIATION STEPS:
  1. Verify the image name and tag are correct
  2. Check if the image exists: docker pull docker.io/myorg/myapp:v2.0.0
  3. Ensure the image was pushed to the registry after building
  4. Verify registry URL is accessible from your cluster

RBAC PERMISSIONS REQUIRED:
  - pods/get (namespace: default)
  - events/list (namespace: default)

AUDIT TRAIL:
  [2025-12-18 10:21:00] GET pods/my-app-pod (namespace: default)
  [2025-12-18 10:21:01] LIST events (namespace: default, fieldSelector: involvedObject.name=my-app-pod)

Analysis completed in 1.2s
```

**Output (JSON Format - clarification Q1)**:
```json
{
  "generated_at": "2025-12-18T10:21:00Z",
  "tool_version": "0.1.0",
  "target_type": "pod",
  "target_name": "my-app-pod",
  "namespace": "default",
  "summary": {
    "total_pods_analyzed": 1,
    "pods_with_issues": 1,
    "total_containers": 1,
    "containers_with_issues": 1,
    "root_cause_breakdown": {
      "IMAGE_NOT_FOUND": 1
    },
    "high_severity_count": 1,
    "medium_severity_count": 0,
    "low_severity_count": 0
  },
  "findings": [
    {
      "root_cause": "IMAGE_NOT_FOUND",
      "severity": "HIGH",
      "pod_name": "my-app-pod",
      "pod_namespace": "default",
      "affected_containers": ["app-container"],
      "summary": "The image 'docker.io/myorg/myapp:v2.0.0' does not exist in the registry.",
      "details": "Image pull failed with ErrImagePull: manifest for docker.io/myorg/myapp:v2.0.0 not found",
      "remediation_steps": [
        "Verify the image name and tag are correct",
        "Check if the image exists: docker pull docker.io/myorg/myapp:v2.0.0",
        "Ensure the image was pushed to the registry after building",
        "Verify registry URL is accessible from your cluster"
      ],
      "image_references": [
        {
          "container_name": "app-container",
          "full_reference": "docker.io/myorg/myapp:v2.0.0",
          "registry": "docker.io",
          "repository": "myorg/myapp",
          "tag": "v2.0.0",
          "digest": "",
          "is_digest": false
        }
      ],
      "is_transient": false,
      "failure_count": 5,
      "first_failure_time": "2025-12-18T10:15:23Z",
      "last_failure_time": "2025-12-18T10:20:45Z",
      "failure_duration": "5m22s"
    }
  ],
  "audit_log": [
    {
      "timestamp": "2025-12-18T10:21:00Z",
      "resource_type": "pods",
      "resource_name": "my-app-pod",
      "namespace": "default",
      "operation": "get"
    },
    {
      "timestamp": "2025-12-18T10:21:01Z",
      "resource_type": "events",
      "resource_name": "",
      "namespace": "default",
      "operation": "list"
    }
  ]
}
```

**Exit Codes**:
- `0`: Analysis completed successfully (issues found or no issues)
- `1`: CLI usage error (invalid flags, missing arguments)
- `2`: Kubernetes API error (connection failed, permission denied)
- `3`: Pod not found
- `4`: Analysis timeout (>30s for single pod)

---

## User Story 2 (P2): Detailed Diagnostic Report

### Additional Flags for Detailed Reports

**Syntax**:
```bash
k8t analyze imagepullbackoff pod <POD_NAME> --detailed [flags]
```

**New Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--detailed` | `-d` | bool | `false` | Include detailed diagnostics (timeline, network checks) |
| `--include-events` | | bool | `false` | Include raw Kubernetes events in output |

**Enhanced Output (Detailed Mode)**:
```
ImagePullBackOff Analysis: my-app-pod (namespace: default) [DETAILED]
================================================================================

ROOT CAUSE: Cannot reach registry
SEVERITY: MEDIUM
AFFECTED CONTAINERS: app-container

... (standard output) ...

NETWORK DIAGNOSTICS:
  Registry Host: private-registry.example.com

  DNS Resolution:
    Status: ✓ SUCCESS
    Resolved IPs: 192.168.1.100, 192.168.1.101
    Duration: 45ms

  TCP Connection (port 443):
    Status: ✗ FAILED
    Error: connection timeout after 5s
    Duration: 5001ms

  HTTP HEAD Request:
    Status: ✗ SKIPPED (TCP failed)

EVENT TIMELINE:
  [2025-12-18 10:15:23] BackOff: Back-off pulling image "private-registry.example.com/app:v1.0"
  [2025-12-18 10:16:45] Failed: Failed to pull image: dial tcp 192.168.1.100:443: i/o timeout
  [2025-12-18 10:18:12] BackOff: Back-off pulling image (retrying in 1m20s)
  [2025-12-18 10:19:34] Failed: rpc error: code = Unknown desc = failed to pull and unpack image
  [2025-12-18 10:20:45] BackOff: Back-off pulling image (retrying in 2m40s)

... (rest of output) ...
```

---

## User Story 3 (P3): Multi-Pod Analysis

### Command: Analyze Workload

**Syntax**:
```bash
k8t analyze imagepullbackoff workload <TYPE>/<NAME> [flags]
k8t analyze imagepullbackoff deployment <NAME> [flags]    # Shorthand
k8t analyze imagepullbackoff statefulset <NAME> [flags]  # Shorthand
k8t analyze imagepullbackoff daemonset <NAME> [flags]    # Shorthand
```

**Example Usage**:
```bash
# Analyze all pods in a deployment
k8t analyze imagepullbackoff deployment my-app -n production

# Analyze statefulset
k8t analyze imagepullbackoff statefulset database -n production

# JSON output for automation
k8t analyze imagepullbackoff deployment my-app -o json
```

**Output (Text Format - Grouped by Root Cause per clarification Q4)**:
```
ImagePullBackOff Analysis: deployment/my-app (namespace: production)
================================================================================

SUMMARY:
  Total Pods Analyzed: 3
  Pods with Issues: 2
  Total Containers: 3
  Containers with Issues: 2

ROOT CAUSE BREAKDOWN:
  • AUTHENTICATION_FAILURE: 2 pods (HIGH severity)

================================================================================
ROOT CAUSE: Registry authentication failed
SEVERITY: HIGH
AFFECTED PODS: my-app-6c8b7d9f-abc12, my-app-6c8b7d9f-def34
AFFECTED CONTAINERS: app-container (2 pods)

SUMMARY:
Authentication to registry 'gcr.io' failed. ImagePullSecrets are either missing or invalid.

COMMON REMEDIATION STEPS:
  1. Verify imagePullSecrets exist in namespace: kubectl get secrets -n production
  2. Check secret type is kubernetes.io/dockerconfigjson
  3. Verify secret contains valid credentials for gcr.io
  4. Ensure pod references the secret in spec.imagePullSecrets

POD DETAILS:
  Pod: my-app-6c8b7d9f-abc12
    Containers: app-container
    First Failure: 2025-12-18 10:10:00 UTC
    Failure Count: 4

  Pod: my-app-6c8b7d9f-def34
    Containers: app-container
    First Failure: 2025-12-18 10:10:05 UTC
    Failure Count: 4

IMAGEPULLSECRETS CHECKED:
  • gcr-secret (referenced in pod spec)
    Status: EXISTS
    Type: kubernetes.io/dockerconfigjson
    Issue: Credentials may be expired or invalid

... (rest of output) ...
```

### Command: Analyze Namespace

**Syntax**:
```bash
k8t analyze imagepullbackoff namespace <NAME> [flags]
k8t analyze imagepullbackoff ns <NAME> [flags]  # Shorthand
```

**Example Usage**:
```bash
# Analyze all pods in namespace
k8t analyze imagepullbackoff namespace production

# Limit to pods with issues only
k8t analyze imagepullbackoff namespace production --issues-only
```

**New Flags**:
| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--issues-only` | | bool | `false` | Show only pods with ImagePullBackOff issues |
| `--max-pods` | | int | `1000` | Maximum number of pods to analyze |

---

## Global Flags

Applicable to all commands:

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--verbose` | `-v` | bool | `false` | Enable verbose logging (audit trail to stderr) |
| `--quiet` | `-q` | bool | `false` | Suppress non-essential output |
| `--timeout` | | duration | `30s` | Analysis timeout (per-pod) |
| `--no-color` | | bool | `false` | Disable colored output |
| `--version` | | bool | - | Show tool version and exit |
| `--help` | `-h` | bool | - | Show help and exit |

---

## Output Formats (FR-007, Clarification Q1)

### Text (Human-Readable)

- Default format
- Colored output (ANSI codes) unless `--no-color`
- Tables using `text/tabwriter`
- Status indicators: ✓ (success), ✗ (failure), ! (warning)

### JSON

- Structured output matching `AnalysisReport` type from data-model.md
- Pretty-printed with 2-space indentation
- All timestamps in RFC3339 format
- Safe for automation/scripting

### YAML

- Structured output matching `AnalysisReport` type
- Uses `gopkg.in/yaml.v3` formatting
- Compatible with K8s resource YAML style

---

## Error Handling

### Permission Errors (RR-001)

```bash
$ k8t analyze imagepullbackoff pod my-pod -n production

ERROR: Insufficient RBAC permissions

Required permissions in namespace 'production':
  ✗ pods/get
  ✗ events/list

Current permissions:
  ✓ pods/list

To grant required permissions, apply this RBAC policy:
---
apiVersion: rbac.authorization.k8s.io/v1
kind:Role
metadata:
  name: k8t-analyzer
  namespace: production
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["list"]

Exit code: 2
```

### Pod Not Found

```bash
$ k8t analyze imagepullbackoff pod nonexistent -n production

ERROR: Pod not found

Pod 'nonexistent' does not exist in namespace 'production'.

Suggestions:
  • Check pod name spelling
  • Verify namespace is correct
  • List pods: kubectl get pods -n production

Exit code: 3
```

### API Timeout (RR-002)

```bash
$ k8t analyze imagepullbackoff pod my-pod

ERROR: Kubernetes API timeout

Failed to complete analysis within 30s timeout.

Possible causes:
  • Kubernetes API server overloaded
  • Network connectivity issues
  • Large number of events to process

Suggestions:
  • Retry the analysis
  • Increase timeout: --timeout 60s
  • Check cluster health: kubectl cluster-info

Exit code: 4
```

---

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `KUBECONFIG` | Path to kubeconfig file | `~/.kube/config` |
| `K8T_OUTPUT_FORMAT` | Default output format | `text` |
| `K8T_NAMESPACE` | Default namespace | `default` |
| `K8T_TIMEOUT` | Default timeout | `30s` |
| `NO_COLOR` | Disable colored output (any value) | - |

---

## Examples by Scenario

### Scenario 1: Image Not Found

```bash
$ k8t analyze imagepullbackoff pod broken-app

ImagePullBackOff Analysis: broken-app (namespace: default)
================================================================================

ROOT CAUSE: Image does not exist in registry
SEVERITY: HIGH
AFFECTED CONTAINERS: main

SUMMARY:
The image 'myregistry.io/app:v999' does not exist in the registry.

REMEDIATION STEPS:
  1. Verify the image name and tag are correct
  2. Check if the image exists: docker pull myregistry.io/app:v999
  3. List available tags: curl https://myregistry.io/v2/app/tags/list
  4. Ensure the image was pushed after building

Exit code: 0
```

### Scenario 2: Authentication Failure

```bash
$ k8t analyze imagepullbackoff pod private-app --detailed

ImagePullBackOff Analysis: private-app (namespace: default) [DETAILED]
================================================================================

ROOT CAUSE: Registry authentication failed
SEVERITY: HIGH
AFFECTED CONTAINERS: main

SUMMARY:
Authentication to registry 'gcr.io' failed. ImagePullSecrets are missing or invalid.

IMAGEPULLSECRETS ANALYSIS:
  Referenced Secrets: gcr-creds
  Secret Status: EXISTS
  Secret Type: kubernetes.io/dockerconfigjson
  Registry in Secret: gcr.io
  Issue: Credentials expired (last rotated 90+ days ago)

REMEDIATION STEPS:
  1. Verify secret exists: kubectl get secret gcr-creds -n default
  2. Check secret content (redacted): kubectl get secret gcr-creds -o jsonpath='{.data}'
  3. Rotate credentials and update secret
  4. Service account credentials may need refreshing: gcloud auth configure-docker

Exit code: 0
```

### Scenario 3: Network Issue with Diagnostics

```bash
$ k8t analyze imagepullbackoff pod network-test -d

ImagePullBackOff Analysis: network-test (namespace: default) [DETAILED]
================================================================================

ROOT CAUSE: Cannot reach registry
SEVERITY: MEDIUM
AFFECTED CONTAINERS: main

NETWORK DIAGNOSTICS:
  Registry Host: internal-registry.corp:5000

  DNS Resolution:
    Status: ✗ FAILED
    Error: lookup internal-registry.corp: no such host
    Duration: 102ms

  TCP Connection:
    Status: ✗ SKIPPED (DNS failed)

  HTTP HEAD Request:
    Status: ✗ SKIPPED (DNS failed)

REMEDIATION STEPS:
  1. Check DNS configuration in cluster
  2. Verify registry hostname is correct: 'internal-registry.corp'
  3. Test DNS resolution: kubectl run -it dns-test --image=busybox --restart=Never -- nslookup internal-registry.corp
  4. Check CoreDNS logs: kubectl logs -n kube-system -l k8s-app=kube-dns
  5. Verify network policies allow egress to registry

Exit code: 0
```

### Scenario 4: Multi-Pod Analysis (Deployment)

```bash
$ k8t analyze imagepullbackoff deployment web-app -n production -o json | jq '.summary'

{
  "total_pods_analyzed": 5,
  "pods_with_issues": 3,
  "total_containers": 5,
  "containers_with_issues": 3,
  "root_cause_breakdown": {
    "RATE_LIMIT_EXCEEDED": 3
  },
  "high_severity_count": 0,
  "medium_severity_count": 3,
  "low_severity_count": 0
}
```

---

## Security & Audit (SR-004, SR-007)

### Audit Trail Output (Clarification Q5: stdout/stderr simple parsing)

When `--verbose` flag is used, audit logs are written to stderr:

```bash
$ k8t analyze imagepullbackoff pod my-pod --verbose

[stderr output:]
AUDIT: 2025-12-18T10:21:00Z GET pods/my-pod namespace=default
AUDIT: 2025-12-18T10:21:01Z LIST events namespace=default filter=involvedObject.name=my-pod
AUDIT: 2025-12-18T10:21:02Z Analysis completed duration=1.2s

[stdout output:]
ImagePullBackOff Analysis: my-pod (namespace: default)
...
```

### Credential Redaction (SR-003, SR-007)

All output formats automatically redact sensitive data:
- ImagePullSecrets contents (show existence/type only)
- Registry passwords/tokens
- Authorization headers
- API keys

Example (detailed mode with secrets):
```
IMAGEPULLSECRETS ANALYSIS:
  Referenced Secrets: docker-creds
  Secret Status: EXISTS
  Secret Type: kubernetes.io/dockerconfigjson
  Registry in Secret: docker.io
  Credentials: [REDACTED]  ← Always redacted
  Last Updated: 2025-11-15
```

---

## Shell Completion

Generate completion scripts for various shells:

```bash
# Bash
k8t completion bash > /etc/bash_completion.d/k8t

# Zsh
k8t completion zsh > "${fpath[1]}/_k8t"

# Fish
k8t completion fish > ~/.config/fish/completions/k8t.fish

# PowerShell
k8t completion powershell > k8t.ps1
```

---

## Future Enhancements (Post-MVP)

Not in current scope but documented for reference:

1. **kubectl Plugin Integration**:
   ```bash
   kubectl analyze imagepullbackoff pod my-pod
   ```

2. **Watch Mode** (real-time monitoring):
   ```bash
   k8t analyze imagepullbackoff pod my-pod --watch
   ```

3. **Historical Analysis** (requires persistence):
   ```bash
   k8t analyze imagepullbackoff pod my-pod --since 1h
   ```

4. **Auto-Remediation** (requires write permissions):
   ```bash
   k8t analyze imagepullbackoff pod my-pod --fix
   ```

5. **Export to File**:
   ```bash
   k8t analyze imagepullbackoff namespace prod --output json > report.json
   ```

---

## Contract Validation

This contract satisfies all functional requirements:

- **FR-007**: Multiple output formats (text, JSON, YAML) ✓
- **FR-008**: Single pod analysis via `pod <NAME>` ✓
- **FR-009**: Workload analysis via `deployment|statefulset|daemonset <NAME>` ✓
- **FR-010**: Namespace analysis via `namespace <NAME>` ✓
- **FR-011**: Credentials redacted from all output ✓
- **FR-012**: Clear error messages when events unavailable ✓
- **RR-001**: Permission errors provide actionable RBAC guidance ✓
- **RR-002**: Timeout handling with clear error messages ✓
- **SR-004**: Audit trail via `--verbose` to stderr ✓
- **SR-007**: All output safe to share (redacted) ✓

Constitution compliance:
- **Security-First**: RBAC errors clear, credentials never exposed ✓
- **Diagnostic Excellence**: Actionable output, remediation steps ✓
- **Reliability**: Error handling, timeouts, exit codes ✓
