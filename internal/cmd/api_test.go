package cmd

import (
	"net/http"
	"testing"
)

func TestIsSupportedRawMethod(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		if !isSupportedRawMethod(method) {
			t.Fatalf("%s should be supported", method)
		}
	}
	if isSupportedRawMethod(http.MethodOptions) {
		t.Fatal("OPTIONS should not be supported")
	}
}

func TestIsValidRawPath(t *testing.T) {
	for _, path := range []string{"/v1/info", "/v1/chats?limit=5"} {
		if !isValidRawPath(path) {
			t.Fatalf("%q should be accepted", path)
		}
	}
	for _, path := range []string{"", "v1/info", "//example.com/path", "https://example.com/path"} {
		if isValidRawPath(path) {
			t.Fatalf("%q should be rejected", path)
		}
	}
}

func TestDecodeRawResponseBodyJSON(t *testing.T) {
	got := decodeRawResponseBody([]byte(`{"ok":true,"count":2}`), "application/json")
	body, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("decoded body type = %T, want map", got)
	}
	if body["ok"] != true || body["count"].(float64) != 2 {
		t.Fatalf("decoded body = %#v", body)
	}
}

func TestDecodeRawResponseBodyText(t *testing.T) {
	got := decodeRawResponseBody([]byte("hello"), "text/plain")
	if got != "hello" {
		t.Fatalf("decoded body = %#v, want hello", got)
	}
}

func TestDecodeRawResponseBodyBinary(t *testing.T) {
	got := decodeRawResponseBody([]byte{0xff, 0x00, 0x01}, "application/octet-stream")
	body, ok := got.(map[string]interface{})
	if !ok {
		t.Fatalf("decoded body type = %T, want map", got)
	}
	if body["base64"] != true || body["bodyBase64"] != "/wAB" {
		t.Fatalf("decoded body = %#v", body)
	}
}
