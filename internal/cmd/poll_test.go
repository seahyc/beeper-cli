package cmd

import "testing"

func TestParseBeeperPollText(t *testing.T) {
	raw := `<p>Msia wedding RSVP (1/4): 谁已确认出席？</p><ol><li>妈咪</li><li>爸爸</li><li>Ying Xin &amp; Marcus</li></ol><p>(This message is a poll. Please open WhatsApp to vote.)</p>`

	question, options, ok := parseBeeperPollText(raw)
	if !ok {
		t.Fatal("expected poll text to parse")
	}
	if question != "Msia wedding RSVP (1/4): 谁已确认出席？" {
		t.Fatalf("question = %q", question)
	}

	want := []string{"妈咪", "爸爸", "Ying Xin & Marcus"}
	if len(options) != len(want) {
		t.Fatalf("options length = %d, want %d: %#v", len(options), len(want), options)
	}
	for i := range want {
		if options[i] != want[i] {
			t.Fatalf("options[%d] = %q, want %q", i, options[i], want[i])
		}
	}
}

func TestParseBeeperPollTextIgnoresRegularOrderedList(t *testing.T) {
	raw := `<p>Dinner order</p><ol><li>Noodles</li><li>Rice</li></ol>`

	if _, _, ok := parseBeeperPollText(raw); ok {
		t.Fatal("regular ordered list should not parse as a poll")
	}
}

func TestParsePollMessageCopiesMetadata(t *testing.T) {
	message := map[string]interface{}{
		"id":         "94457",
		"sortKey":    "66776",
		"chatID":     "!chat:beeper.local",
		"accountID":  "whatsapp",
		"timestamp":  "2026-06-05T01:18:46.000Z",
		"senderID":   "@sender:beeper.local",
		"senderName": "Sender",
		"text":       `<p>Question?</p><ol><li>A</li><li>B</li></ol><p>(This message is a poll. Please open WhatsApp to vote.)</p>`,
	}

	poll, ok := parsePollMessage(message)
	if !ok {
		t.Fatal("expected message to parse as poll")
	}
	if poll.MessageID != "94457" || poll.SortKey != "66776" || poll.ChatID != "!chat:beeper.local" {
		t.Fatalf("metadata was not copied: %#v", poll)
	}
	if poll.VoteCountsAvailable {
		t.Fatal("Beeper API poll parsing should not report vote counts as available")
	}
}
