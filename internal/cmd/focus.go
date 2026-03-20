package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var focusCmd = &cobra.Command{
	Use:   "focus",
	Short: "Focus Beeper Desktop app",
	Run: func(cmd *cobra.Command, args []string) {
		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("chat"); v != "" {
			body["chatID"] = v
		}
		if v, _ := cmd.Flags().GetString("message"); v != "" {
			body["messageID"] = v
		}
		if v, _ := cmd.Flags().GetString("draft"); v != "" {
			body["draftText"] = v
		}
		if v, _ := cmd.Flags().GetString("draft-file"); v != "" {
			data, err := os.ReadFile(v)
			if err != nil {
				output.Fatal("FILE_ERROR", err)
			}
			body["draftText"] = string(data)
		}

		client := api.NewClient(getBaseURL())
		var result interface{}
		if err := client.Post("/v1/focus", body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

func init() {
	focusCmd.Flags().String("chat", "", "Navigate to specific chat")
	focusCmd.Flags().String("message", "", "Jump to specific message")
	focusCmd.Flags().String("draft", "", "Pre-fill draft text")
	focusCmd.Flags().String("draft-file", "", "Pre-fill draft from file")
}
