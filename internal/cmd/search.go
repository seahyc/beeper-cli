package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Unified search: chats + participants + messages",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/search?query=%s", url.QueryEscape(args[0]))
		var result interface{}
		if err := client.Get(path, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}
