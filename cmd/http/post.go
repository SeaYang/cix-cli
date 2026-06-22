package http

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newPostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "post <url>",
		Short:   "Send an HTTP POST request",
		Args:    cobra.ExactArgs(1),
		Example: `  cix http post https://httpbin.org/post --data '{"key":"value"}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			rawURL := args[0]
			data, _ := cmd.Flags().GetString("data")
			timeout, _ := cmd.Flags().GetInt("timeout")
			headers, _ := cmd.Flags().GetStringSlice("header")
			verbose, _ := cmd.Flags().GetBool("verbose")

			client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
			req, err := http.NewRequest("POST", rawURL, strings.NewReader(data))
			if err != nil {
				return fmt.Errorf("invalid URL: %w", err)
			}
			req.Header.Set("Content-Type", "application/json")
			if err := setHeaders(req.Header, headers); err != nil {
				return err
			}

			if verbose {
				fmt.Printf("> POST %s\n", rawURL)
				fmt.Printf("> %s\n", data)
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

	cmd.Flags().StringP("data", "d", "", "Request body (JSON)")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}
