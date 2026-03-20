package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TokenData struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
}

var (
	tokenStore *TokenData
	tokenOnce  sync.Once
	tokenMu    sync.RWMutex
)

func tokenFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".beeper", "token.json")
}

func GetToken() *TokenData {
	tokenOnce.Do(func() {
		tokenStore = &TokenData{}
		_ = tokenStore.Load()
	})
	return tokenStore
}

func (t *TokenData) Load() error {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	data, err := os.ReadFile(tokenFilePath())
	if err != nil {
		return err
	}
	return json.Unmarshal(data, t)
}

func (t *TokenData) Save() error {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	dir := filepath.Dir(tokenFilePath())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tokenFilePath(), data, 0600)
}

func (t *TokenData) isValidUnlocked() bool {
	return t.AccessToken != "" && time.Now().Add(5*time.Minute).Before(t.ExpiresAt)
}

func (t *TokenData) IsValid() bool {
	tokenMu.RLock()
	defer tokenMu.RUnlock()
	return t.isValidUnlocked()
}

func (t *TokenData) Clear() error {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	*t = TokenData{}
	return os.Remove(tokenFilePath())
}
