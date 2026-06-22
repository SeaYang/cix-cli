package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <url>",
		Short: "Send an HTTP GET request",
		Args:  cobra.ExactArgs(1),
		Example: `  cix http get https://httpbin.org/get
  cix http get https://api.example.com --header "Authorization:Bearer token"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rawURL := args[0]
			timeout, _ := cmd.Flags().GetInt("timeout")
			headers, _ := cmd.Flags().GetStringSlice("header")
			verbose, _ := cmd.Flags().GetBool("verbose")

			client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
			req, err := http.NewRequest("GET", rawURL, nil)
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}
			if err := setHeaders(req.Header, headers); err != nil {
				return err
			}

			if verbose {
				fmt.Printf("> GET %s\n", rawURL)
				for k, v := range req.Header {
					fmt.Printf("> %s: %s\n", k, strings.Join(v, ", "))
				}
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			fmt.Printf("HTTP %d\n", resp.StatusCode)
			fmt.Println(string(body))
			return nil
		},
	}
}
