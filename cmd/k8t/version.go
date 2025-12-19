package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version information - set via ldflags during build
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// newVersionCmd creates the version command
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Display the version, git commit, and build date of k8t",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("k8t version %s\n", Version)
			fmt.Printf("Git commit: %s\n", GitCommit)
			fmt.Printf("Build date: %s\n", BuildDate)
		},
	}
}
