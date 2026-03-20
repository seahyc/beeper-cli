package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"github.com/yjwong/beeper-cli/internal/output"
)

type serverInfo struct {
	Endpoints struct {
		OAuth struct {
			AuthorizationEndpoint string `json:"authorization_endpoint"`
			TokenEndpoint         string `json:"token_endpoint"`
			RegistrationEndpoint  string `json:"registration_endpoint"`
			RevocationEndpoint    string `json:"revocation_endpoint"`
		} `json:"oauth"`
	} `json:"endpoints"`
}

type registrationResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func discoverEndpoints(baseURL string) (*serverInfo, error) {
	resp, err := http.Get(baseURL + "/v1/info")
	if err != nil {
		return nil, fmt.Errorf("cannot connect to Beeper Desktop at %s. Is it running? %w", baseURL, err)
	}
	defer resp.Body.Close()
	var info serverInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("failed to parse server info: %w", err)
	}
	return &info, nil
}

func registerClient(registrationEndpoint, redirectURI string) (*registrationResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"client_name":   "beeper-cli",
		"redirect_uris": []string{redirectURI},
		"grant_types":   []string{"authorization_code", "refresh_token"},
	})
	resp, err := http.Post(registrationEndpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var reg registrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&reg); err != nil {
		return nil, err
	}
	return &reg, nil
}

func startCallbackServer() (string, <-chan string, func(), error) {
	// Try preferred port, fallback to OS-assigned
	var listener net.Listener
	var err error
	for _, port := range []string{":9876", ":0"} {
		listener, err = net.Listen("tcp", "localhost"+port)
		if err == nil {
			break
		}
	}
	if err != nil {
		return "", nil, nil, fmt.Errorf("failed to start callback server: %w", err)
	}

	addr := listener.Addr().(*net.TCPAddr)
	redirectURI := fmt.Sprintf("http://localhost:%d/callback", addr.Port)
	codeChan := make(chan string, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code", 400)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Authorization successful!</h2><p>You can close this tab.</p></body></html>")
		codeChan <- code
	})

	server := &http.Server{Handler: mux}
	go server.Serve(listener)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}

	return redirectURI, codeChan, cleanup, nil
}

func exchangeCode(tokenEndpoint, code, redirectURI, clientID, clientSecret string) (*tokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}
	var tok tokenResponse
	return &tok, json.Unmarshal(body, &tok)
}

func refreshAccessToken(tokenEndpoint, refreshToken, clientID, clientSecret string) (*tokenResponse, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
	}
	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("refresh failed: %s", string(body))
	}
	var tok tokenResponse
	return &tok, json.Unmarshal(body, &tok)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}
	return cmd.Start()
}

func doFullAuthFlow(baseURL string) error {
	info, err := discoverEndpoints(baseURL)
	if err != nil {
		return err
	}

	redirectURI, codeChan, cleanup, err := startCallbackServer()
	if err != nil {
		return err
	}
	defer cleanup()

	token := GetToken()

	// Register client if we don't have credentials
	if token.ClientID == "" {
		reg, err := registerClient(info.Endpoints.OAuth.RegistrationEndpoint, redirectURI)
		if err != nil {
			return fmt.Errorf("client registration failed: %w", err)
		}
		token.ClientID = reg.ClientID
		token.ClientSecret = reg.ClientSecret
	}

	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s",
		info.Endpoints.OAuth.AuthorizationEndpoint,
		url.QueryEscape(token.ClientID),
		url.QueryEscape(redirectURI),
	)

	fmt.Println("First run — opening browser for Beeper authorization...")
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Open this URL in your browser:\n%s\n", authURL)
	}

	// Wait for callback (timeout after 2 minutes)
	select {
	case code := <-codeChan:
		tok, err := exchangeCode(info.Endpoints.OAuth.TokenEndpoint, code, redirectURI, token.ClientID, token.ClientSecret)
		if err != nil {
			return err
		}
		token.AccessToken = tok.AccessToken
		token.RefreshToken = tok.RefreshToken
		token.ExpiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
		return token.Save()
	case <-time.After(2 * time.Minute):
		return fmt.Errorf("authorization timed out after 2 minutes")
	}
}

func doRefresh(baseURL string) error {
	info, err := discoverEndpoints(baseURL)
	if err != nil {
		return err
	}
	token := GetToken()
	tok, err := refreshAccessToken(info.Endpoints.OAuth.TokenEndpoint, token.RefreshToken, token.ClientID, token.ClientSecret)
	if err != nil {
		return err
	}
	token.AccessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		token.RefreshToken = tok.RefreshToken
	}
	token.ExpiresAt = time.Now().Add(time.Duration(tok.ExpiresIn) * time.Second)
	return token.Save()
}

// EnsureValidToken checks token validity and refreshes or re-auths as needed.
func EnsureValidToken(baseURL string) {
	token := GetToken()
	if token.IsValid() {
		return
	}
	// Try refresh first
	if token.RefreshToken != "" {
		if err := doRefresh(baseURL); err == nil {
			return
		}
	}
	// Full auth flow
	if err := doFullAuthFlow(baseURL); err != nil {
		output.Fatal("AUTH_ERROR", err)
	}
}

func RevokeToken(baseURL string) error {
	info, err := discoverEndpoints(baseURL)
	if err != nil {
		return err
	}
	token := GetToken()
	if token.AccessToken != "" {
		data := url.Values{
			"token":         {token.AccessToken},
			"client_id":     {token.ClientID},
			"client_secret": {token.ClientSecret},
		}
		http.PostForm(info.Endpoints.OAuth.RevocationEndpoint, data)
	}
	return token.Clear()
}
