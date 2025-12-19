package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/aboigues/k8t/pkg/analyzer"
	"github.com/aboigues/k8t/pkg/k8s"
	"github.com/aboigues/k8t/pkg/output"
)

// Global flags
var (
	kubeconfig string
	verbose    bool
	quiet      bool
	noColor    bool
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

// newRootCmd creates the root command
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "k8t",
		Short: "Kubernetes Administration Toolkit",
		Long: `k8t is a diagnostic CLI tool for identifying root causes of
ImagePullBackOff errors in Kubernetes pods.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "Suppress all output except errors")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Add subcommands
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newAnalyzeCmd())
	rootCmd.AddCommand(newCheckCmd())

	return rootCmd
}

// newAnalyzeCmd creates the analyze command
func newAnalyzeCmd() *cobra.Command {
	analyzeCmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze Kubernetes resources for issues",
		Long:  "Analyze various Kubernetes resources to identify and diagnose problems",
	}

	// Add subcommands
	analyzeCmd.AddCommand(newImagePullBackOffCmd())

	return analyzeCmd
}

// Flags for imagepullbackoff command
var (
	namespace     string
	outputFormat  string
	timeoutStr    string
)

// newImagePullBackOffCmd creates the imagepullbackoff subcommand
func newImagePullBackOffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "imagepullbackoff <pod-name>",
		Short: "Analyze ImagePullBackOff errors for a pod",
		Long: `Analyze ImagePullBackOff errors for a specific pod and provide
root cause analysis with remediation steps.`,
		Args: cobra.ExactArgs(1),
		RunE: runImagePullBackOffAnalysis,
	}

	// Command-specific flags
	cmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Kubernetes namespace")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json, yaml)")
	cmd.Flags().StringVar(&timeoutStr, "timeout", "30s", "Analysis timeout duration")

	return cmd
}

// runImagePullBackOffAnalysis executes the ImagePullBackOff analysis
func runImagePullBackOffAnalysis(cmd *cobra.Command, args []string) error {
	podName := args[0]

	// Parse timeout
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return fmt.Errorf("invalid timeout duration '%s': %w", timeoutStr, err)
	}

	// Parse output format
	format, err := output.ParseFormat(outputFormat)
	if err != nil {
		return err
	}

	// Create Kubernetes client
	client, err := k8s.NewClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Validate cluster connectivity
	if err := client.Validate(); err != nil {
		return fmt.Errorf("failed to connect to Kubernetes cluster: %w", err)
	}

	// Create audit logger
	auditLogger, err := output.NewAuditLogger(verbose)
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}
	defer auditLogger.Close()

	// Create analyzer
	az := analyzer.NewAnalyzer(client, auditLogger, timeout)

	// Run analysis
	ctx := context.Background()
	report, err := az.AnalyzePod(ctx, namespace, podName)
	if err != nil {
		return handleAnalysisError(err, auditLogger)
	}

	// Format and output results
	if !quiet {
		if err := output.Format(report, format, noColor, os.Stdout); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	}

	// Exit with appropriate code
	if report.Summary.PodsWithIssues > 0 {
		// Issues found, but analysis succeeded
		return nil
	}

	return nil
}

// handleAnalysisError maps errors to appropriate exit codes and messages
func handleAnalysisError(err error, logger *output.AuditLogger) error {
	// Log the error
	if logger != nil {
		logger.LogError("Analysis failed", err)
	}

	// Map to specific error types
	switch e := err.(type) {
	case *analyzer.PodNotFoundError:
		fmt.Fprintf(os.Stderr, "ERROR: Pod not found\n\n")
		fmt.Fprintf(os.Stderr, "Pod '%s' does not exist in namespace '%s'.\n\n", e.PodName, e.Namespace)
		fmt.Fprintf(os.Stderr, "Suggestions:\n")
		fmt.Fprintf(os.Stderr, "  • Check pod name spelling\n")
		fmt.Fprintf(os.Stderr, "  • Verify namespace is correct\n")
		fmt.Fprintf(os.Stderr, "  • List pods: kubectl get pods -n %s\n", e.Namespace)
		os.Exit(3)
		return err

	case *analyzer.PermissionError:
		fmt.Fprintf(os.Stderr, "ERROR: Insufficient RBAC permissions\n\n")
		fmt.Fprintf(os.Stderr, "Required: %s/%s in namespace '%s'\n\n", e.Resource, e.Verb, e.Namespace)
		fmt.Fprintf(os.Stderr, "To grant permissions, create a Role and RoleBinding:\n\n")
		fmt.Fprintf(os.Stderr, "kubectl create role k8t-reader --verb=get,list --resource=pods,events -n %s\n", e.Namespace)
		fmt.Fprintf(os.Stderr, "kubectl create rolebinding k8t-binding --role=k8t-reader --user=<your-user> -n %s\n", e.Namespace)
		os.Exit(2)
		return err

	case *analyzer.TimeoutError:
		fmt.Fprintf(os.Stderr, "ERROR: Analysis timeout\n\n")
		fmt.Fprintf(os.Stderr, "Failed to complete analysis within %v\n\n", e.Timeout)
		fmt.Fprintf(os.Stderr, "Suggestions:\n")
		fmt.Fprintf(os.Stderr, "  • Retry the analysis\n")
		fmt.Fprintf(os.Stderr, "  • Increase timeout: --timeout 60s\n")
		fmt.Fprintf(os.Stderr, "  • Check cluster connectivity\n")
		os.Exit(4)
		return err

	case *analyzer.NoImagePullBackOffError:
		fmt.Fprintf(os.Stderr, "INFO: No ImagePullBackOff detected\n\n")
		fmt.Fprintf(os.Stderr, "Pod '%s' in namespace '%s' does not have ImagePullBackOff status.\n\n", e.PodName, e.Namespace)
		fmt.Fprintf(os.Stderr, "Check pod status: kubectl describe pod %s -n %s\n", e.PodName, e.Namespace)
		return nil

	default:
		// Generic error
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(2)
		return err
	}
}

// Flags for check command
var (
	allNamespaces bool
	checkNamespace string
)

// newCheckCmd creates the check command
func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check cluster for potential issues",
		Long: `Check the Kubernetes cluster for potential issues across all namespaces
or a specific namespace. This command scans for common problems like
ImagePullBackOff, CrashLoopBackOff, and other pod errors.`,
		RunE: runCheckAnalysis,
	}

	// Command-specific flags
	cmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Check all namespaces")
	cmd.Flags().StringVarP(&checkNamespace, "namespace", "n", "default", "Namespace to check")

	return cmd
}

// runCheckAnalysis executes the cluster check
func runCheckAnalysis(cmd *cobra.Command, args []string) error {
	// Create Kubernetes client
	client, err := k8s.NewClient(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Validate cluster connectivity
	if err := client.Validate(); err != nil {
		return fmt.Errorf("failed to connect to Kubernetes cluster: %w", err)
	}

	// Create audit logger
	auditLogger, err := output.NewAuditLogger(verbose)
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}
	defer auditLogger.Close()

	ctx := context.Background()
	var namespacesToCheck []string

	// Determine which namespaces to check
	if allNamespaces {
		namespaces, err := client.ListNamespaces(ctx)
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %w", err)
		}
		namespacesToCheck = namespaces
	} else {
		namespacesToCheck = []string{checkNamespace}
	}

	// Track overall results
	totalIssues := 0
	issuesByNamespace := make(map[string]int)

	// Check each namespace
	for _, ns := range namespacesToCheck {
		if verbose {
			fmt.Fprintf(os.Stderr, "Checking namespace: %s\n", ns)
		}

		pods, err := client.ListPodsInNamespace(ctx, ns)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to list pods in namespace %s: %v\n", ns, err)
			continue
		}

		if verbose {
			fmt.Fprintf(os.Stderr, "Found %d pods in namespace %s\n", len(pods), ns)
		}

		nsIssues := 0
		for _, pod := range pods {
			// Check pod status for common issues
			hasIssue, issueType := checkPodIssues(pod)
			if hasIssue {
				if !quiet {
					fmt.Printf("[%s] Pod: %s/%s - Status: %s\n",
						issueType, ns, pod.Name, pod.Status.Phase)
				}
				nsIssues++
				totalIssues++
			}
		}

		if nsIssues > 0 {
			issuesByNamespace[ns] = nsIssues
		}
	}

	// Display summary
	if !quiet {
		fmt.Println("\n--- Summary ---")
		if totalIssues == 0 {
			fmt.Println("No issues found!")
		} else {
			fmt.Printf("Total issues found: %d\n", totalIssues)
			fmt.Println("\nIssues by namespace:")
			for ns, count := range issuesByNamespace {
				fmt.Printf("  %s: %d issue(s)\n", ns, count)
			}
		}
	}

	// Return error if issues found (cobra will handle exit code)
	if totalIssues > 0 {
		return fmt.Errorf("found %d issue(s) in cluster", totalIssues)
	}

	return nil
}

// checkPodIssues checks if a pod has common issues
func checkPodIssues(pod k8s.PodInfo) (bool, string) {
	// Check for ImagePullBackOff or ErrImagePull
	for _, containerStatus := range pod.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			reason := containerStatus.State.Waiting.Reason
			switch reason {
			case "ImagePullBackOff", "ErrImagePull":
				return true, "ImagePullBackOff"
			case "CrashLoopBackOff":
				return true, "CrashLoopBackOff"
			case "CreateContainerConfigError":
				return true, "ConfigError"
			case "InvalidImageName":
				return true, "InvalidImage"
			}
		}

		// Check if container is restarting frequently
		if containerStatus.RestartCount > 5 {
			return true, "HighRestarts"
		}
	}

	// Check pod phase
	if pod.Status.Phase == "Failed" || pod.Status.Phase == "Unknown" {
		return true, "PodFailed"
	}

	return false, ""
}
