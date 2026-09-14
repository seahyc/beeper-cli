package cmd

import (
	"strings"
	"testing"
)

func TestParseExpirySeconds(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want int
	}{
		{name: "plain seconds", raw: "3600", want: 3600},
		{name: "hours", raw: "1h", want: 3600},
		{name: "minutes", raw: "30m", want: 1800},
		{name: "compound", raw: "1h30m", want: 5400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExpirySeconds(tt.raw)
			if err != nil {
				t.Fatalf("parseExpirySeconds(%q) returned error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("parseExpirySeconds(%q) = %d, want %d", tt.raw, got, tt.want)
			}
		})
	}
}

func TestParseExpirySecondsRejectsInvalidValues(t *testing.T) {
	for _, raw := range []string{"", "-1", "-1h", "1.5s", "abc"} {
		t.Run(raw, func(t *testing.T) {
			if _, err := parseExpirySeconds(raw); err == nil {
				t.Fatalf("parseExpirySeconds(%q) returned nil error", raw)
			}
		})
	}
}

func TestValidateChatExpiryUsesAdvertisedMillisecondTimers(t *testing.T) {
	chat := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"disappearingTimer": map[string]interface{}{
				"timers": []interface{}{
					float64(86400000),
					float64(172800000),
				},
			},
		},
	}

	if err := validateChatExpiry(chat, 86400); err != nil {
		t.Fatalf("validateChatExpiry rejected allowed value: %v", err)
	}

	err := validateChatExpiry(chat, 3600)
	if err == nil {
		t.Fatal("validateChatExpiry accepted unsupported value")
	}
	if !strings.Contains(err.Error(), "allowed values: 24h, 2d") {
		t.Fatalf("error did not list allowed values: %v", err)
	}
}

func TestValidateChatExpiryAllowsUnconstrainedChats(t *testing.T) {
	chat := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"disappearingTimer": map[string]interface{}{},
		},
	}
	if err := validateChatExpiry(chat, 3600); err != nil {
		t.Fatalf("validateChatExpiry rejected unconstrained chat: %v", err)
	}
}

func TestExpiryMatches(t *testing.T) {
	seconds := 3600
	if !expiryMatches(float64(3600), &seconds) {
		t.Fatal("expiryMatches did not match numeric value")
	}
	if expiryMatches(nil, &seconds) {
		t.Fatal("expiryMatches matched nil for concrete value")
	}
	if !expiryMatches(nil, nil) {
		t.Fatal("expiryMatches did not match nil clear value")
	}
}
