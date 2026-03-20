package cmd

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var reminderCmd = &cobra.Command{
	Use:   "reminder",
	Short: "Reminder commands",
}

var reminderSetCmd = &cobra.Command{
	Use:   "set <chatID>",
	Short: "Set a reminder on a chat",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		atStr, _ := cmd.Flags().GetString("at")
		if atStr == "" {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("--at is required"))
		}

		t, err := time.Parse(time.RFC3339, atStr)
		if err != nil {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("--at must be an ISO timestamp (e.g. 2026-03-20T15:00:00Z): %w", err))
		}

		reminder := map[string]interface{}{
			"remindAtMs": t.UnixMilli(),
		}
		if dismiss, _ := cmd.Flags().GetBool("dismiss-on-reply"); dismiss {
			reminder["dismissOnIncomingMessage"] = true
		}

		body := map[string]interface{}{
			"reminder": reminder,
		}

		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/reminders", url.PathEscape(args[0]))
		var result interface{}
		if err := client.Post(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var reminderClearCmd = &cobra.Command{
	Use:   "clear <chatID>",
	Short: "Clear a reminder on a chat",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/reminders", url.PathEscape(args[0]))
		var result interface{}
		if err := client.Delete(path, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

func init() {
	reminderSetCmd.Flags().String("at", "", "ISO timestamp for reminder (required)")
	reminderSetCmd.Flags().Bool("dismiss-on-reply", false, "Auto-dismiss on incoming message")

	reminderCmd.AddCommand(reminderSetCmd, reminderClearCmd)
}
