package cmd

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
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
		// The Desktop API currently returns a file:// cache path for some
		// encrypted image attachments. That path contains ciphertext, while the
		// original message's srcURL includes the metadata needed to decrypt it.
		// Recover it from Beeper's local index (read-only) when available.
		if resolved, err := encryptedSourceURL(mxcURL); err == nil && resolved != "" {
			mxcURL = resolved
		}
		path := fmt.Sprintf("/v1/assets/serve?url=%s", url.QueryEscape(mxcURL))
		outputPath, _ := cmd.Flags().GetString("output")
		result, err := client.ServeAsset(path, outputPath)
		if err != nil {
			output.Fatal("API_ERROR", err)
		}
		if err := decryptServedAsset(mxcURL, result.SavedPath); err != nil {
			output.Fatal("DECRYPT_ERROR", err)
		}
		output.JSON(map[string]interface{}{
			"contentType": result.ContentType,
			"savedPath":   result.SavedPath,
			"size":        result.Size,
		})
	},
}

type localAttachment struct {
	ID     string `json:"id"`
	SrcURL string `json:"srcURL"`
}

type localMessage struct {
	Attachments []localAttachment `json:"attachments"`
}

type encryptedFileInfo struct {
	Hashes map[string]string `json:"hashes"`
	IV     string            `json:"iv"`
	Key    struct {
		K string `json:"k"`
	} `json:"key"`
}

// encryptedSourceURL maps Beeper's encrypted local cache path back to the
// attachment srcURL stored in its local message index. The index is opened
// immutable/read-only; failure simply preserves normal asset behavior.
func encryptedSourceURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "file" {
		return "", nil
	}
	base := filepath.Base(parsed.Path)
	if base == "." || base == "/" || base == "" {
		return "", nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dbPath := filepath.Join(home, "Library", "Application Support", "BeeperTexts", "index.db")
	if _, err := os.Stat(dbPath); err != nil {
		return "", err
	}
	dbURL := (&url.URL{Scheme: "file", Path: dbPath}).String()
	dsn := dbURL + "?mode=ro&immutable=1"
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT message FROM mx_room_messages WHERE message LIKE ?`, "%"+base+"%")
	if err != nil {
		return "", err
	}
	defer rows.Close()
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return "", err
		}
		var message localMessage
		if err := json.Unmarshal([]byte(raw), &message); err != nil {
			continue
		}
		for _, attachment := range message.Attachments {
			if strings.Contains(attachment.ID, base) && strings.Contains(attachment.SrcURL, "encryptedFileInfoJSON=") {
				return attachment.SrcURL, nil
			}
		}
	}
	return "", rows.Err()
}

// decryptServedAsset handles encrypted Matrix media when Desktop's serve
// endpoint returns its cached ciphertext instead of decrypted bytes.
func decryptServedAsset(rawURL, path string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	encodedInfo := parsed.Query().Get("encryptedFileInfoJSON")
	if encodedInfo == "" {
		return nil
	}
	infoBytes, err := base64.RawURLEncoding.DecodeString(encodedInfo)
	if err != nil {
		return fmt.Errorf("decode encrypted media metadata: %w", err)
	}
	var info encryptedFileInfo
	if err := json.Unmarshal(infoBytes, &info); err != nil {
		return fmt.Errorf("parse encrypted media metadata: %w", err)
	}
	key, err := base64.RawURLEncoding.DecodeString(info.Key.K)
	if err != nil {
		return fmt.Errorf("decode media key: %w", err)
	}
	iv, err := base64.RawURLEncoding.DecodeString(info.IV)
	if err != nil {
		return fmt.Errorf("decode media IV: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("initialize AES: %w", err)
	}
	if len(iv) != block.BlockSize() {
		return fmt.Errorf("invalid media IV length %d", len(iv))
	}
	ciphertext, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if expected := info.Hashes["sha256"]; expected != "" {
		// Matrix stores the integrity hash of the encrypted payload.
		actual := sha256.Sum256(ciphertext)
		if strings.TrimRight(base64.StdEncoding.EncodeToString(actual[:]), "=") != expected {
			// Beeper Desktop sometimes serves its already-decrypted local cache
			// even when the request URL still carries Matrix encryption metadata.
			// Keep recognizable plaintext intact; opaque mismatches remain errors
			// so corrupted ciphertext is not silently accepted.
			if servedAssetAppearsPlaintext(ciphertext) {
				return nil
			}
			return fmt.Errorf("encrypted media checksum mismatch")
		}
	}
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCTR(block, iv).XORKeyStream(plaintext, ciphertext)
	return os.WriteFile(path, plaintext, 0o644)
}

func servedAssetAppearsPlaintext(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	return http.DetectContentType(data) != "application/octet-stream"
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
