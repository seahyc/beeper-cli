package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var assetCmd = &cobra.Command{
	Use:   "asset",
	Short: "Asset commands",
}

var assetUploadCmd = &cobra.Command{
	Use:   "upload <file>",
	Short: "Upload a file (multipart)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		var result interface{}
		if err := client.UploadFile("/v1/assets/upload", args[0], &result); err != nil {
			output.Fatal("UPLOAD_ERROR", err)
		}
		output.JSON(result)
	},
}

var assetUploadBase64Cmd = &cobra.Command{
	Use:   "upload-base64",
	Short: "Upload base64-encoded content",
	Run: func(cmd *cobra.Command, args []string) {
		content, _ := cmd.Flags().GetString("content")
		if content == "" {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("--content is required"))
		}
		body := map[string]interface{}{
			"content": content,
		}
		if v, _ := cmd.Flags().GetString("filename"); v != "" {
			body["fileName"] = v
		}
		if v, _ := cmd.Flags().GetString("mime"); v != "" {
			body["mimeType"] = v
		}

		client := api.NewClient(getBaseURL())
		var result interface{}
		if err := client.Post("/v1/assets/upload/base64", body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var assetDownloadCmd = &cobra.Command{
	Use:   "download <mxc-url>",
	Short: "Download media to local file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		body := map[string]string{"url": args[0]}
		outputPath, _ := cmd.Flags().GetString("output")
		result, err := client.DownloadFile("/v1/assets/download", body, outputPath)
		if err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(map[string]string{
			"url":        result.SourceURL,
			"sourcePath": result.SourcePath,
			"savedPath":  result.SavedPath,
		})
	},
}

var assetServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Fetch served media bytes into a local file",
	Run: func(cmd *cobra.Command, args []string) {
		mxcURL, _ := cmd.Flags().GetString("url")
		if mxcURL == "" {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("--url is required"))
		}
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/assets/serve?url=%s", url.QueryEscape(mxcURL))
		outputPath, _ := cmd.Flags().GetString("output")
		result, err := client.ServeAsset(path, outputPath)
		if err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(map[string]interface{}{
			"contentType": result.ContentType,
			"savedPath":   result.SavedPath,
			"size":        result.Size,
		})
	},
}

func init() {
	assetUploadBase64Cmd.Flags().String("content", "", "Base64-encoded content (required)")
	assetUploadBase64Cmd.Flags().String("filename", "", "Filename")
	assetUploadBase64Cmd.Flags().String("mime", "", "MIME type")

	assetDownloadCmd.Flags().String("output", "", "Output path. If it is an existing directory, or ends with a path separator, the original filename is used inside that directory.")

	assetServeCmd.Flags().String("url", "", "mxc://, localmxc://, or file:// URL (required)")
	assetServeCmd.Flags().String("output", "", "Optional output path for served bytes. Defaults to a temp file.")

	assetCmd.AddCommand(assetUploadCmd, assetUploadBase64Cmd, assetDownloadCmd, assetServeCmd)
}
