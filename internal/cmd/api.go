package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var apiCmd = &cobra.Command{
	Use:     "api <method> <path>",
	Aliases: []string{"raw"},
	Short:   "Call a Beeper Desktop API endpoint directly",
	Example: `  beeper api get /v1/info --no-auth
  beeper api get '/v1/chats?limit=5'
  beeper api post /v1/chats/<encoded-chat-id>/messages --body '{"text":"Hello"}'`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		method := strings.ToUpper(args[0])
		if !isSupportedRawMethod(method) {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("unsupported method %q", method))
		}
		if !isValidRawPath(args[1]) {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("path must be a relative Beeper API path beginning with a single slash"))
		}

		body, err := rawRequestBody(cmd)
		if err != nil {
			output.Fatal("VALIDATION_ERROR", err)
		}

		contentType, _ := cmd.Flags().GetString("content-type")
		noAuth, _ := cmd.Flags().GetBool("no-auth")
		includeHeaders, _ := cmd.Flags().GetBool("headers")

		client := api.NewClient(getBaseURL())
		resp, err := client.Raw(method, args[1], body, contentType, !noAuth)
		if err != nil {
			output.Fatal("API_ERROR", err)
		}

		result := map[string]interface{}{
			"ok":         resp.StatusCode >= 200 && resp.StatusCode < 300,
			"statusCode": resp.StatusCode,
			"status":     resp.Status,
			"method":     method,
			"path":       args[1],
		}
		if includeHeaders {
			result["headers"] = resp.Headers
		}
		if len(resp.Body) > 0 {
			result["body"] = decodeRawResponseBody(resp.Body, resp.Headers.Get("Content-Type"))
		} else {
			result["body"] = nil
		}
		output.JSON(result)
	},
}

func isValidRawPath(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//")
}

func isSupportedRawMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func rawRequestBody(cmd *cobra.Command) ([]byte, error) {
	body, _ := cmd.Flags().GetString("body")
	bodyFile, _ := cmd.Flags().GetString("body-file")
	if body != "" && bodyFile != "" {
		return nil, fmt.Errorf("use only one of --body or --body-file")
	}
	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read --body-file: %w", err)
		}
		return data, nil
	}
	if body == "" {
		return nil, nil
	}
	return []byte(body), nil
}

func decodeRawResponseBody(body []byte, contentType string) interface{} {
	var parsed interface{}
	if json.Unmarshal(body, &parsed) == nil {
		return parsed
	}
	if strings.HasPrefix(strings.ToLower(contentType), "text/") || utf8.Valid(body) {
		return string(body)
	}
	return map[string]interface{}{
		"base64":     true,
		"bodyBase64": base64.StdEncoding.EncodeToString(body),
	}
}

func init() {
	apiCmd.Flags().String("body", "", "Raw request body")
	apiCmd.Flags().String("body-file", "", "Read raw request body from file")
	apiCmd.Flags().String("content-type", "application/json", "Request Content-Type when a body is present")
	apiCmd.Flags().Bool("no-auth", false, "Do not attach the Beeper OAuth bearer token")
	apiCmd.Flags().Bool("headers", true, "Include response headers")
}
