package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yjwong/beeper-cli/internal/auth"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetUnauthenticated makes a request without auth (for /v1/info discovery)
func (c *Client) GetUnauthenticated(path string, result interface{}) error {
	return c.doRequest("GET", path, nil, result, false)
}

func (c *Client) Get(path string, result interface{}) error {
	return c.doRequest("GET", path, nil, result, true)
}

func (c *Client) Post(path string, body, result interface{}) error {
	return c.doRequest("POST", path, body, result, true)
}

func (c *Client) Put(path string, body, result interface{}) error {
	return c.doRequest("PUT", path, body, result, true)
}

func (c *Client) Patch(path string, body, result interface{}) error {
	return c.doRequest("PATCH", path, body, result, true)
}

func (c *Client) Delete(path string, result interface{}) error {
	return c.doRequest("DELETE", path, nil, result, true)
}

func (c *Client) DeleteWithBody(path string, body, result interface{}) error {
	return c.doRequest("DELETE", path, body, result, true)
}

func (c *Client) doRequest(method, path string, body, result interface{}, needsAuth bool) error {
	return c.doRequestInner(method, path, body, result, needsAuth, false)
}

func (c *Client) doRequestInner(method, path string, body, result interface{}, needsAuth, isRetry bool) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if needsAuth {
		token := auth.GetToken()
		if !token.IsValid() {
			auth.EnsureValidToken(c.baseURL)
		}
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode == 401 && needsAuth && !isRetry {
		// Token expired mid-request, re-auth and retry once
		auth.EnsureValidToken(c.baseURL)
		return c.doRequestInner(method, path, body, result, true, true)
	}

	if resp.StatusCode >= 400 {
		var apiErr struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("%s: %s", apiErr.Code, apiErr.Message)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if len(respBody) == 0 {
			return nil
		}
		return json.Unmarshal(respBody, result)
	}
	return nil
}

// UploadFile uploads a file via multipart/form-data
func (c *Client) UploadFile(path, filePath string, result interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath)))
	if contentType == "" {
		header := make([]byte, 512)
		n, err := file.Read(header)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to sniff file content type: %w", err)
		}
		contentType = http.DetectContentType(header[:n])
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to rewind file after sniffing: %w", err)
		}
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	partHeader := textproto.MIMEHeader{}
	partHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(filePath)))
	partHeader.Set("Content-Type", contentType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest("POST", c.baseURL+path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	token := auth.GetToken()
	if !token.IsValid() {
		auth.EnsureValidToken(c.baseURL)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("upload failed %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		return json.Unmarshal(respBody, result)
	}
	return nil
}

type DownloadResult struct {
	SourceURL  string
	SavedPath  string
	SourcePath string
}

type ServeResult struct {
	ContentType string
	SavedPath   string
	Size        int64
}

// DownloadFile resolves a Beeper asset to a local file and optionally copies it to outputPath.
func (c *Client) DownloadFile(path string, body interface{}, outputPath string) (*DownloadResult, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	token := auth.GetToken()
	if !token.IsValid() {
		auth.EnsureValidToken(c.baseURL)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("download failed %d: %s", resp.StatusCode, string(respBody))
	}

	// The download endpoint returns JSON with a local file URL.
	var downloadResp struct {
		URL    string `json:"url"`
		SrcURL string `json:"srcURL"`
	}
	if err := json.Unmarshal(respBody, &downloadResp); err != nil {
		return nil, err
	}

	sourceURL := downloadResp.URL
	if sourceURL == "" {
		sourceURL = downloadResp.SrcURL
	}
	if sourceURL == "" {
		return nil, fmt.Errorf("download response did not include a file URL")
	}

	sourcePath, err := filePathFromURL(sourceURL)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(sourcePath); err != nil {
		return nil, fmt.Errorf("download source does not exist: %w", err)
	}

	savedPath := sourcePath
	if outputPath != "" {
		savedPath, err = resolveDownloadOutputPath(outputPath, sourcePath)
		if err != nil {
			return nil, err
		}
		if err := copyFile(sourcePath, savedPath); err != nil {
			return nil, err
		}
	}

	return &DownloadResult{
		SourceURL:  sourceURL,
		SourcePath: sourcePath,
		SavedPath:  savedPath,
	}, nil
}

// ServeAsset fetches the rendered bytes from Beeper's asset serve endpoint and saves them locally.
func (c *Client) ServeAsset(path string, outputPath string) (*ServeResult, error) {
	req, err := http.NewRequest("GET", c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	token := auth.GetToken()
	if !token.IsValid() {
		auth.EnsureValidToken(c.baseURL)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("serve failed %d: %s", resp.StatusCode, string(respBody))
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(respBody)
	}

	savedPath, err := resolveServeOutputPath(outputPath, contentType)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(savedPath, respBody, 0o644); err != nil {
		return nil, err
	}

	return &ServeResult{
		ContentType: contentType,
		SavedPath:   savedPath,
		Size:        int64(len(respBody)),
	}, nil
}

func filePathFromURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid download URL: %w", err)
	}
	if parsed.Scheme != "file" {
		return "", fmt.Errorf("unsupported download URL scheme %q", parsed.Scheme)
	}

	path, err := url.PathUnescape(parsed.Path)
	if err != nil {
		return "", fmt.Errorf("invalid escaped download path: %w", err)
	}
	if path == "" {
		return "", fmt.Errorf("download URL did not include a file path")
	}
	return path, nil
}

func resolveDownloadOutputPath(outputPath, sourcePath string) (string, error) {
	info, err := os.Stat(outputPath)
	if err == nil {
		if info.IsDir() {
			return filepath.Join(outputPath, filepath.Base(sourcePath)), nil
		}
		return outputPath, nil
	}
	if !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to inspect output path: %w", err)
	}

	// Treat paths with a trailing separator as directories even before they exist.
	if strings.HasSuffix(outputPath, string(os.PathSeparator)) {
		if err := os.MkdirAll(outputPath, 0o755); err != nil {
			return "", fmt.Errorf("failed to create output directory: %w", err)
		}
		return filepath.Join(outputPath, filepath.Base(sourcePath)), nil
	}

	parentDir := filepath.Dir(outputPath)
	if parentDir != "." {
		if err := os.MkdirAll(parentDir, 0o755); err != nil {
			return "", fmt.Errorf("failed to create parent directory: %w", err)
		}
	}
	return outputPath, nil
}

func resolveServeOutputPath(outputPath, contentType string) (string, error) {
	if outputPath != "" {
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return "", fmt.Errorf("failed to create output directory: %w", err)
		}
		return outputPath, nil
	}

	exts := preferredExtensionsForContentType(contentType)
	pattern := "beeper-asset-*"
	if len(exts) > 0 {
		pattern += exts[0]
	}

	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	path := f.Name()
	if err := f.Close(); err != nil {
		return "", err
	}
	return path, nil
}

func preferredExtensionsForContentType(contentType string) []string {
	baseType := strings.TrimSpace(strings.Split(contentType, ";")[0])
	switch baseType {
	case "video/mp4":
		return []string{".mp4"}
	case "image/jpeg":
		return []string{".jpg"}
	case "image/png":
		return []string{".png"}
	case "application/pdf":
		return []string{".pdf"}
	}

	exts, _ := mime.ExtensionsByType(baseType)
	return exts
}

func copyFile(sourcePath, destPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer func() {
		_ = destFile.Close()
	}()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy downloaded file: %w", err)
	}
	if err := destFile.Close(); err != nil {
		return fmt.Errorf("failed to finalize output file: %w", err)
	}
	return nil
}
