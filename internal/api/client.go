package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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

func (c *Client) Delete(path string, result interface{}) error {
	return c.doRequest("DELETE", path, nil, result, true)
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

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
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

// DownloadFile downloads a file and saves to disk
func (c *Client) DownloadFile(path string, body interface{}, outputPath string) (string, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	token := auth.GetToken()
	if !token.IsValid() {
		auth.EnsureValidToken(c.baseURL)
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("download failed %d: %s", resp.StatusCode, string(respBody))
	}

	// The download endpoint returns JSON with a local file URL
	var downloadResp struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(respBody, &downloadResp); err != nil {
		return "", err
	}

	return downloadResp.URL, nil
}
