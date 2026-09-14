package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yjwong/beeper-cli/internal/api"
)

type validationError struct {
	err error
}

func (e validationError) Error() string {
	return e.err.Error()
}

func (e validationError) Unwrap() error {
	return e.err
}

func asValidationError(err error) error {
	return validationError{err: err}
}

func isValidationError(err error) bool {
	var target validationError
	return errors.As(err, &target)
}

func parseExpirySeconds(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, fmt.Errorf("duration is required")
	}

	if secs, err := strconv.Atoi(raw); err == nil {
		if secs < 0 {
			return 0, fmt.Errorf("duration must be non-negative")
		}
		return secs, nil
	}

	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("duration must be seconds or a Go duration like 1h, 30m, or 24h")
	}
	if d < 0 {
		return 0, fmt.Errorf("duration must be non-negative")
	}
	if d%time.Second != 0 {
		return 0, fmt.Errorf("duration must resolve to whole seconds")
	}
	return int(d / time.Second), nil
}

func getChat(client *api.Client, chatID string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/v1/chats/%s", url.PathEscape(chatID))
	var chat map[string]interface{}
	if err := client.Get(path, &chat); err != nil {
		return nil, err
	}
	return chat, nil
}

func chatExpirySeconds(chat map[string]interface{}) *int {
	value, ok := numberToInt(chat["messageExpirySeconds"])
	if !ok {
		return nil
	}
	return &value
}

func expiryValue(seconds *int) interface{} {
	if seconds == nil {
		return nil
	}
	return *seconds
}

func setChatExpiry(client *api.Client, chatID string, seconds *int) (map[string]interface{}, error) {
	body := map[string]interface{}{"messageExpirySeconds": expiryValue(seconds)}
	path := fmt.Sprintf("/v1/chats/%s", url.PathEscape(chatID))
	var result map[string]interface{}
	if err := client.Patch(path, body, &result); err != nil {
		return nil, err
	}

	if !expiryMatches(result["messageExpirySeconds"], seconds) {
		requested := "cleared"
		if seconds != nil {
			requested = formatSeconds(*seconds)
		}
		return result, fmt.Errorf("Beeper Desktop did not apply disappearing-message timer %s; returned messageExpirySeconds=%v", requested, result["messageExpirySeconds"])
	}

	return result, nil
}

func validateChatExpiry(chat map[string]interface{}, seconds int) error {
	allowed, constrained := allowedExpirySeconds(chat)
	if !constrained || len(allowed) == 0 {
		return nil
	}
	for _, allowedSeconds := range allowed {
		if seconds == allowedSeconds {
			return nil
		}
	}
	return fmt.Errorf("unsupported disappearing-message timer %s for this chat; allowed values: %s", formatSeconds(seconds), formatAllowedSeconds(allowed))
}

func allowedExpirySeconds(chat map[string]interface{}) ([]int, bool) {
	capabilities, ok := chat["capabilities"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	disappearing, ok := capabilities["disappearingTimer"].(map[string]interface{})
	if !ok {
		return nil, false
	}
	rawTimers, ok := disappearing["timers"].([]interface{})
	if !ok {
		return nil, false
	}

	allowed := make([]int, 0, len(rawTimers))
	seen := map[int]bool{}
	for _, raw := range rawTimers {
		timerMs, ok := numberToInt(raw)
		if !ok || timerMs < 0 || timerMs%1000 != 0 {
			continue
		}
		seconds := timerMs / 1000
		if !seen[seconds] {
			allowed = append(allowed, seconds)
			seen[seconds] = true
		}
	}
	sort.Ints(allowed)
	return allowed, true
}

func numberToInt(raw interface{}) (int, bool) {
	switch v := raw.(type) {
	case int:
		return v, true
	case int64:
		i := int(v)
		return i, int64(i) == v
	case float64:
		i := int(v)
		return i, float64(i) == v
	case json.Number:
		i, err := strconv.Atoi(v.String())
		return i, err == nil
	default:
		return 0, false
	}
}

func expiryMatches(raw interface{}, expected *int) bool {
	if expected == nil {
		return raw == nil
	}
	actual, ok := numberToInt(raw)
	return ok && actual == *expected
}

func formatAllowedSeconds(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, formatSeconds(value))
	}
	return strings.Join(parts, ", ")
}

func formatSeconds(seconds int) string {
	if seconds != 0 && seconds%86400 == 0 {
		days := seconds / 86400
		if days == 1 {
			return "24h"
		}
		return fmt.Sprintf("%dd", days)
	}
	if seconds != 0 && seconds%3600 == 0 {
		return fmt.Sprintf("%dh", seconds/3600)
	}
	if seconds != 0 && seconds%60 == 0 {
		return fmt.Sprintf("%dm", seconds/60)
	}
	return fmt.Sprintf("%ds", seconds)
}
