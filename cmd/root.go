// Package cmd holds the root command for the cix CLI.
// Each functional subcommand lives in its own sub-package under cmd/ (e.g.
// cmd/version, cmd/http) and is registered here via its NewCmd() constructor.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/seayang/cix-cli/cmd/http"
	"github.com/seayang/cix-cli/cmd/version"
)

// verbose is set by the global --verbose / -v flag.
var verbose bool

var rootCmd = &cobra.Command{
	Use:   "cix",
	Short: "CI pipeline command-line toolkit",
	Long: `cix is a command-line toolkit for CI pipelines.
It bundles common utilities (HTTP client, version info, ...) intended to run
inside container-based pipeline steps.`,
	SilenceUsage: true,
}

// Execute runs the root command and exits with a non-zero status on error.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	// Register every subcommand. New commands go in their own sub-package under
	// cmd/ and are wired up here.
	rootCmd.AddCommand(version.NewCmd())
	rootCmd.AddCommand(http.NewCmd())
}
