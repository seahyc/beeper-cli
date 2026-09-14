package cmd

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecryptServedAssetDecryptsCiphertext(t *testing.T) {
	plaintext := []byte("plain asset bytes")
	rawURL, ciphertext := encryptedAssetFixture(t, plaintext)
	path := filepath.Join(t.TempDir(), "asset.bin")
	if err := os.WriteFile(path, ciphertext, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := decryptServedAsset(rawURL, path); err != nil {
		t.Fatalf("decryptServedAsset returned error: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Fatalf("decrypted bytes = %q, want %q", got, plaintext)
	}
}

func TestDecryptServedAssetPreservesAlreadyDecryptedMedia(t *testing.T) {
	jpeg := append([]byte{0xff, 0xd8, 0xff, 0xe0}, bytes.Repeat([]byte{0x42}, 32)...)
	rawURL, _ := encryptedAssetFixture(t, []byte("different encrypted payload"))
	path := filepath.Join(t.TempDir(), "asset.jpg")
	if err := os.WriteFile(path, jpeg, 0o600); err != nil {
		t.Fatal(err)
	}

	if err := decryptServedAsset(rawURL, path); err != nil {
		t.Fatalf("decryptServedAsset returned error for decoded JPEG: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, jpeg) {
		t.Fatal("already-decrypted media was modified")
	}
}

func TestDecryptServedAssetRejectsOpaqueChecksumMismatch(t *testing.T) {
	rawURL, _ := encryptedAssetFixture(t, []byte("expected plaintext"))
	path := filepath.Join(t.TempDir(), "asset.bin")
	opaque := bytes.Repeat([]byte{0x00, 0x01, 0x02, 0x03}, 32)
	if err := os.WriteFile(path, opaque, 0o600); err != nil {
		t.Fatal(err)
	}

	err := decryptServedAsset(rawURL, path)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("decryptServedAsset error = %v, want checksum mismatch", err)
	}
	got, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(got, opaque) {
		t.Fatal("mismatched opaque payload was modified")
	}
}

func encryptedAssetFixture(t *testing.T, plaintext []byte) (string, []byte) {
	t.Helper()
	key := []byte("0123456789abcdef0123456789abcdef")
	iv := []byte("0123456789abcdef")
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	ciphertext := make([]byte, len(plaintext))
	cipher.NewCTR(block, iv).XORKeyStream(ciphertext, plaintext)
	hash := sha256.Sum256(ciphertext)
	info := map[string]interface{}{
		"hashes": map[string]string{"sha256": base64.RawStdEncoding.EncodeToString(hash[:])},
		"iv":     base64.RawURLEncoding.EncodeToString(iv),
		"key":    map[string]interface{}{"k": base64.RawURLEncoding.EncodeToString(key), "ext": true},
	}
	encoded, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}
	return "mxc://local.beeper.com/asset?encryptedFileInfoJSON=" + url.QueryEscape(base64.RawURLEncoding.EncodeToString(encoded)), ciphertext
}
