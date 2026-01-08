# Krew Plugin Distribution Guide

This guide explains how to distribute k8t via the krew plugin manager.

## Prerequisites

- A GitHub release with kubectl-k8t binaries for all platforms
- Access to fork and submit PRs to [krew-index](https://github.com/kubernetes-sigs/krew-index)

## Publishing to Krew

### 1. Create a Release

First, create a new release with goreleaser:

```bash
# Tag the release
git tag -a v0.1.0 -m "Release v0.1.0"
git push origin v0.1.0

# Build and publish release
goreleaser release --clean
```

This will create GitHub releases with binaries for all platforms.

### 2. Update the Krew Manifest

After the release is published, update `k8t.yaml` with:

1. The correct version number
2. SHA256 checksums for each platform

To generate SHA256 checksums:

```bash
# Download each release artifact and compute SHA256
curl -sL https://github.com/aboigues/k8t/releases/download/v0.1.0/kubectl-k8t_v0.1.0_linux_amd64.tar.gz | sha256sum
curl -sL https://github.com/aboigues/k8t/releases/download/v0.1.0/kubectl-k8t_v0.1.0_linux_arm64.tar.gz | sha256sum
curl -sL https://github.com/aboigues/k8t/releases/download/v0.1.0/kubectl-k8t_v0.1.0_darwin_amd64.tar.gz | sha256sum
curl -sL https://github.com/aboigues/k8t/releases/download/v0.1.0/kubectl-k8t_v0.1.0_darwin_arm64.tar.gz | sha256sum
curl -sL https://github.com/aboigues/k8t/releases/download/v0.1.0/kubectl-k8t_v0.1.0_windows_amd64.zip | sha256sum
```

Update each `sha256` field in `k8t.yaml` with the computed values.

### 3. Test the Plugin Locally

Before submitting to krew-index, test the plugin manifest locally:

```bash
# Validate the manifest
kubectl krew install --manifest=k8t.yaml

# Test the plugin
kubectl k8t version
kubectl k8t analyze imagepullbackoff --help

# Uninstall for cleanup
kubectl krew uninstall k8t
```

### 4. Submit to Krew Index

Once tested, submit the plugin to the official krew-index:

```bash
# Fork the krew-index repository
# https://github.com/kubernetes-sigs/krew-index

# Clone your fork
git clone https://github.com/YOUR_USERNAME/krew-index.git
cd krew-index

# Create a new branch
git checkout -b add-k8t-plugin

# Copy the plugin manifest
cp /path/to/k8t/k8t.yaml plugins/k8t.yaml

# Commit and push
git add plugins/k8t.yaml
git commit -m "Add k8t plugin"
git push origin add-k8t-plugin

# Create a pull request
# Go to https://github.com/kubernetes-sigs/krew-index and create a PR
```

### 5. Maintain the Plugin

For subsequent releases:

1. Create a new release with goreleaser
2. Update the plugin manifest with new version and checksums
3. Submit a PR to krew-index with the updated manifest

```bash
cd krew-index
git checkout -b update-k8t-v0.2.0
cp /path/to/k8t/k8t.yaml plugins/k8t.yaml
git add plugins/k8t.yaml
git commit -m "Update k8t plugin to v0.2.0"
git push origin update-k8t-v0.2.0
```

## Automated Updates

Consider using [krew-release-bot](https://github.com/rajatjindal/krew-release-bot) to automate plugin updates:

1. Add a `.krew.yaml` template to your repository
2. Configure GitHub Actions to use krew-release-bot
3. Bot automatically creates PRs to krew-index on new releases

Example GitHub Action:

```yaml
name: Release to Krew
on:
  release:
    types: [published]

jobs:
  krew:
    runs-on: ubuntu-latest
    steps:
      - uses: rajatjindal/krew-release-bot@v0.0.46
        with:
          krew_template_file: k8t.yaml
```

## Troubleshooting

### Plugin not found after installation

```bash
# Ensure krew is properly configured
kubectl krew version

# Update krew index
kubectl krew update

# Try installing again
kubectl krew install k8t
```

### SHA256 mismatch

If users report SHA256 mismatches:

1. Verify the release artifacts haven't been modified
2. Regenerate checksums from the published release artifacts
3. Update the manifest and submit a fix PR to krew-index

## Resources

- [Krew Plugin Developer Guide](https://krew.sigs.k8s.io/docs/developer-guide/)
- [kubectl Plugin Documentation](https://kubernetes.io/docs/tasks/extend-kubectl/kubectl-plugins/)
- [krew-index Repository](https://github.com/kubernetes-sigs/krew-index)
- [krew-release-bot](https://github.com/rajatjindal/krew-release-bot)
