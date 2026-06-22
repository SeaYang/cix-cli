// Package http implements the `cix http` command group.
package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// NewCmd creates the `cix http` command group and its subcommands.
func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "http",
		Short: "HTTP client utilities",
		Long:  "Send HTTP requests and inspect responses.",
	}

	cmd.PersistentFlags().Int("timeout", 30, "Request timeout in seconds")
	cmd.PersistentFlags().StringSlice("header", []string{}, "Request headers (key:value)")

	cmd.AddCommand(newGetCmd())
	cmd.AddCommand(newPostCmd())

	return cmd
}

// setHeaders parses "key:value" entries and applies them to the request headers.
func setHeaders(h http.Header, headers []string) error {
	for _, hdr := range headers {
		parts := strings.SplitN(hdr, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid header %q (expected key:value)", hdr)
		}
		h.Set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
	}
	return nil
}
