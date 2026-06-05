package cmd

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

const pollAPINote = "Beeper Desktop exposes WhatsApp polls as text only; open WhatsApp for live vote counts or voter names."

var (
	pollQuestionBeforeListRE = regexp.MustCompile(`(?is)<p[^>]*>(.*?)</p>\s*<ol[^>]*>`)
	pollOrderedListRE        = regexp.MustCompile(`(?is)<ol[^>]*>(.*?)</ol>`)
	pollListItemRE           = regexp.MustCompile(`(?is)<li[^>]*>(.*?)</li>`)
	pollBreakRE              = regexp.MustCompile(`(?i)<br\s*/?>`)
	pollTagRE                = regexp.MustCompile(`(?is)<[^>]*>`)
	pollWhitespaceRE         = regexp.MustCompile(`\s+`)
)

var pollCmd = &cobra.Command{
	Use:   "poll",
	Short: "Poll commands",
}

var pollListCmd = &cobra.Command{
	Use:     "list <chatID>",
	Aliases: []string{"ls", "read"},
	Short:   "List poll messages in a chat",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		result, err := runPollList(cmd, client, args[0])
		if err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

type pollMessage struct {
	MessageID           string   `json:"messageID,omitempty"`
	SortKey             string   `json:"sortKey,omitempty"`
	ChatID              string   `json:"chatID,omitempty"`
	AccountID           string   `json:"accountID,omitempty"`
	Timestamp           string   `json:"timestamp,omitempty"`
	SenderID            string   `json:"senderID,omitempty"`
	SenderName          string   `json:"senderName,omitempty"`
	Question            string   `json:"question"`
	Options             []string `json:"options"`
	Source              string   `json:"source"`
	VoteCountsAvailable bool     `json:"voteCountsAvailable"`
	Note                string   `json:"note,omitempty"`
}

func runPollList(cmd *cobra.Command, client *api.Client, chatID string) (map[string]interface{}, error) {
	limit, _ := cmd.Flags().GetInt("limit")
	if limit <= 0 {
		limit = 20
	}
	pageSize, _ := cmd.Flags().GetInt("page-size")
	if pageSize <= 0 {
		pageSize = 100
	}
	pages, _ := cmd.Flags().GetInt("pages")
	if pages <= 0 {
		pages = 5
	}
	cursor, _ := cmd.Flags().GetString("cursor")
	direction, _ := cmd.Flags().GetString("direction")
	if direction == "" {
		direction = "before"
	}

	items := []pollMessage{}
	searchedPages := 0
	hasMore := false
	var newestCursor, oldestCursor string

	for searchedPages < pages && len(items) < limit {
		params := url.Values{}
		params.Set("limit", fmt.Sprintf("%d", pageSize))
		params.Set("direction", direction)
		if cursor != "" {
			params.Set("cursor", cursor)
		}

		path := fmt.Sprintf("/v1/chats/%s/messages?%s", encodeChatID(chatID), params.Encode())
		var page map[string]interface{}
		if err := client.Get(path, &page); err != nil {
			return nil, err
		}
		searchedPages++

		pageItems, _ := page["items"].([]interface{})
		for _, item := range pageItems {
			message, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			poll, ok := parsePollMessage(message)
			if !ok {
				continue
			}
			items = append(items, poll)
			if len(items) >= limit {
				break
			}
		}

		hasMore, _ = page["hasMore"].(bool)
		newestCursor, _ = page["newestCursor"].(string)
		oldestCursor, _ = page["oldestCursor"].(string)
		if !hasMore {
			break
		}
		nextCursor := oldestCursor
		if direction == "after" {
			nextCursor = newestCursor
		}
		if nextCursor == "" || nextCursor == cursor {
			break
		}
		cursor = nextCursor
	}

	return map[string]interface{}{
		"mode":          "beeper_api_text",
		"chatID":        chatID,
		"items":         items,
		"hasMore":       hasMore,
		"newestCursor":  newestCursor,
		"oldestCursor":  oldestCursor,
		"searchedPages": searchedPages,
		"pageSize":      pageSize,
		"limit":         limit,
		"note":          pollAPINote,
	}, nil
}

func parsePollMessage(message map[string]interface{}) (pollMessage, bool) {
	text := messageText(message)
	question, options, ok := parseBeeperPollText(text)
	if !ok {
		return pollMessage{}, false
	}

	return pollMessage{
		MessageID:           stringField(message, "id"),
		SortKey:             stringField(message, "sortKey"),
		ChatID:              stringField(message, "chatID"),
		AccountID:           stringField(message, "accountID"),
		Timestamp:           stringField(message, "timestamp"),
		SenderID:            stringField(message, "senderID"),
		SenderName:          stringField(message, "senderName"),
		Question:            question,
		Options:             options,
		Source:              "beeper_api_text",
		VoteCountsAvailable: false,
		Note:                pollAPINote,
	}, true
}

func parseBeeperPollText(raw string) (string, []string, bool) {
	if !isBeeperPollText(raw) {
		return "", nil, false
	}

	listMatch := pollOrderedListRE.FindStringSubmatch(raw)
	if len(listMatch) < 2 {
		return "", nil, false
	}

	optionMatches := pollListItemRE.FindAllStringSubmatch(listMatch[1], -1)
	options := []string{}
	for _, match := range optionMatches {
		if len(match) < 2 {
			continue
		}
		option := normalizePollHTMLFragment(match[1])
		if option != "" {
			options = append(options, option)
		}
	}
	if len(options) == 0 {
		return "", nil, false
	}

	question := ""
	if questionMatch := pollQuestionBeforeListRE.FindStringSubmatch(raw); len(questionMatch) >= 2 {
		question = normalizePollHTMLFragment(questionMatch[1])
	}
	if question == "" {
		if listIndex := pollOrderedListRE.FindStringIndex(raw); len(listIndex) == 2 {
			question = normalizePollHTMLFragment(raw[:listIndex[0]])
		}
	}
	if question == "" {
		return "", nil, false
	}

	return question, options, true
}

func isBeeperPollText(raw string) bool {
	normalized := strings.ToLower(html.UnescapeString(raw))
	return strings.Contains(normalized, "this message is a poll") && strings.Contains(normalized, "<ol")
}

func normalizePollHTMLFragment(fragment string) string {
	fragment = pollBreakRE.ReplaceAllString(fragment, "\n")
	fragment = strings.ReplaceAll(fragment, "</p>", "\n")
	fragment = pollTagRE.ReplaceAllString(fragment, "")
	fragment = html.UnescapeString(fragment)
	fragment = strings.ReplaceAll(fragment, "\u00a0", " ")
	fragment = pollWhitespaceRE.ReplaceAllString(fragment, " ")
	return strings.TrimSpace(fragment)
}

func messageText(message map[string]interface{}) string {
	for _, key := range []string{"text", "html", "body"} {
		if value, ok := message[key].(string); ok {
			return value
		}
	}
	return ""
}

func stringField(message map[string]interface{}, key string) string {
	value, _ := message[key].(string)
	return value
}

func init() {
	pollListCmd.Flags().Int("limit", 20, "Max poll messages to return")
	pollListCmd.Flags().String("cursor", "", "Pagination cursor")
	pollListCmd.Flags().String("direction", "before", "Pagination direction (before|after)")
	pollListCmd.Flags().Int("pages", 5, "Max message pages to scan")
	pollListCmd.Flags().Int("page-size", 100, "Messages per page to scan")

	pollCmd.AddCommand(pollListCmd)
}
