// Package version implements the `cix version` command.
package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build-time variables, overridden via -ldflags "-X ...=...".
var (
	Version   = "0.1.0"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// NewCmd creates the `cix version` command.
func NewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("cix %s\n", Version)
			fmt.Printf("  Build Time : %s\n", BuildTime)
			fmt.Printf("  Git Commit : %s\n", GitCommit)
			fmt.Printf("  Go Version : %s\n", runtime.Version())
			fmt.Printf("  OS/Arch    : %s/%s\n", runtime.GOOS, runtime.GOARCH)
		},
	}
}
