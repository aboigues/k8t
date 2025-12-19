package analyzer

import (
	"fmt"

	"github.com/aboigues/k8t/pkg/types"
)

// GenerateRemediationSteps returns actionable steps for a root cause
func GenerateRemediationSteps(rootCause types.RootCause, imageRef *types.ImageReference) []string {
	switch rootCause {
	case types.RootCauseImageNotFound:
		return imageNotFoundRemediation(imageRef)
	case types.RootCauseAuthFailure:
		return authFailureRemediation(imageRef)
	case types.RootCauseNetworkIssue:
		return networkIssueRemediation(imageRef)
	case types.RootCauseRateLimit:
		return rateLimitRemediation(imageRef)
	case types.RootCausePermissionDenied:
		return permissionDeniedRemediation(imageRef)
	case types.RootCauseManifestError:
		return manifestErrorRemediation(imageRef)
	case types.RootCauseTransient:
		return transientFailureRemediation()
	case types.RootCauseUnknown:
		return unknownRemediation(imageRef)
	default:
		return []string{"No remediation steps available for this root cause"}
	}
}

// imageNotFoundRemediation returns steps for IMAGE_NOT_FOUND
func imageNotFoundRemediation(img *types.ImageReference) []string {
	if img == nil {
		return []string{
			"Verify the image name and tag are correct in your pod specification",
			"Check if the image exists in the registry",
			"Ensure the image was pushed to the registry after building",
			"Verify the registry URL is correct and accessible from your cluster",
		}
	}

	return []string{
		fmt.Sprintf("Verify the image name and tag are correct: %s", img.FullReference),
		fmt.Sprintf("Check if the image exists: docker pull %s", img.FullReference),
		"Ensure the image was pushed to the registry after building",
		fmt.Sprintf("Verify registry '%s' is accessible from your cluster", img.Registry),
		"Check if the image tag was deleted or moved",
	}
}

// authFailureRemediation returns steps for AUTHENTICATION_FAILURE
func authFailureRemediation(img *types.ImageReference) []string {
	steps := []string{
		"Create or verify the image pull secret with valid registry credentials",
		"Ensure the secret is in the same namespace as the pod",
		"Reference the secret in pod spec: spec.imagePullSecrets",
	}

	if img != nil {
		steps = append(steps, fmt.Sprintf("Create secret: kubectl create secret docker-registry regcred --docker-server=%s --docker-username=<user> --docker-password=<pwd>", img.Registry))
		steps = append(steps, "Add to pod spec: imagePullSecrets: [{name: regcred}]")
	}

	steps = append(steps, "Verify credentials are still valid (not expired or revoked)")

	return steps
}

// networkIssueRemediation returns steps for NETWORK_ISSUE
func networkIssueRemediation(img *types.ImageReference) []string {
	steps := []string{
		"Check cluster network connectivity to external registries",
		"Verify DNS resolution is working in the cluster",
		"Check for firewall or network policies blocking registry access",
		"Verify proxy settings if cluster uses an HTTP proxy",
	}

	if img != nil {
		steps = append(steps, fmt.Sprintf("Test connectivity: kubectl run test --image=busybox --rm -it -- nslookup %s", img.Registry))
		steps = append(steps, fmt.Sprintf("Test HTTPS access: kubectl run test --image=busybox --rm -it -- wget https://%s", img.Registry))
	}

	steps = append(steps, "Check for service mesh or CNI issues that might block external traffic")

	return steps
}

// rateLimitRemediation returns steps for RATE_LIMIT_EXCEEDED
func rateLimitRemediation(img *types.ImageReference) []string {
	steps := []string{
		"Wait for the rate limit window to reset (typically 5-60 minutes)",
		"Reduce the frequency of image pulls (use imagePullPolicy: IfNotPresent)",
		"Consider using a registry mirror or cache to reduce external pulls",
	}

	if img != nil && img.Registry == "docker.io" {
		steps = append(steps, "Docker Hub rate limits: anonymous users (100 pulls/6h), authenticated (200 pulls/6h)")
		steps = append(steps, "Authenticate with Docker Hub to increase rate limits")
		steps = append(steps, "Consider Docker Hub paid plans for higher limits")
	}

	steps = append(steps, "Use image pull secrets with authenticated registry access")
	steps = append(steps, "Consider deploying a registry mirror in your cluster")

	return steps
}

// permissionDeniedRemediation returns steps for PERMISSION_DENIED
func permissionDeniedRemediation(img *types.ImageReference) []string {
	steps := []string{
		"Verify the service account has permission to pull images",
		"Check if the image repository has access restrictions",
		"Ensure the image pull secret has sufficient permissions",
	}

	if img != nil {
		steps = append(steps, fmt.Sprintf("Verify repository '%s/%s' access permissions", img.Registry, img.Repository))
		steps = append(steps, "Check if the registry requires authentication")
	}

	steps = append(steps, "For private registries, ensure the account in pull secret has read access")
	steps = append(steps, "Review registry access policies and IAM permissions")

	return steps
}

// manifestErrorRemediation returns steps for MANIFEST_ERROR
func manifestErrorRemediation(img *types.ImageReference) []string {
	steps := []string{
		"Verify the image manifest is valid and not corrupted",
		"Check if the image was built for the correct platform (linux/amd64, linux/arm64, etc.)",
		"Ensure multi-platform manifest includes your cluster's architecture",
	}

	if img != nil {
		steps = append(steps, fmt.Sprintf("Inspect image manifest: docker manifest inspect %s", img.FullReference))
		steps = append(steps, "Verify platform compatibility with your Kubernetes nodes")
	}

	steps = append(steps, "Try re-pushing the image to fix potential corruption")
	steps = append(steps, "Check registry logs for manifest-related errors")

	return steps
}

// transientFailureRemediation returns steps for TRANSIENT_FAILURE
func transientFailureRemediation() []string {
	return []string{
		"This appears to be a transient failure (< 3 attempts or < 5 minutes)",
		"Kubernetes will automatically retry pulling the image",
		"Monitor the pod status to see if it resolves automatically",
		"If the issue persists beyond 10 minutes, investigate for underlying causes",
		"Check recent events: kubectl describe pod <pod-name>",
	}
}

// unknownRemediation returns steps for UNKNOWN root cause
func unknownRemediation(img *types.ImageReference) []string {
	steps := []string{
		"Review the full error message in pod events for more details",
		"Check pod events: kubectl describe pod <pod-name>",
		"Verify the image reference is correct and complete",
		"Test image pull manually: docker pull <image>",
		"Check registry status and availability",
		"Review cluster logs for additional error context",
	}

	if img != nil {
		steps = append(steps, fmt.Sprintf("Manually test pull: docker pull %s", img.FullReference))
	}

	steps = append(steps, "Contact registry support if the issue persists")

	return steps
}
