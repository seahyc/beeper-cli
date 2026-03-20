package cmd

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var contactCmd = &cobra.Command{
	Use:   "contact",
	Short: "Contact commands",
}

var contactSearchCmd = &cobra.Command{
	Use:   "search <accountID> <query>",
	Short: "Search contacts on an account",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/accounts/%s/contacts?query=%s", url.PathEscape(args[0]), url.QueryEscape(args[1]))
		var result interface{}
		if err := client.Get(path, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var contactListCmd = &cobra.Command{
	Use:   "list <accountID>",
	Short: "List contacts on an account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		params := url.Values{}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			params.Set("limit", fmt.Sprintf("%d", v))
		}
		if v, _ := cmd.Flags().GetString("cursor"); v != "" {
			params.Set("cursor", v)
		}
		if v, _ := cmd.Flags().GetString("direction"); v != "" {
			params.Set("direction", v)
		}
		if v, _ := cmd.Flags().GetString("query"); v != "" {
			params.Set("query", v)
		}
		path := fmt.Sprintf("/v1/accounts/%s/contacts/list?%s", url.PathEscape(args[0]), params.Encode())
		var result interface{}
		if err := client.Get(path, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

func init() {
	contactListCmd.Flags().Int("limit", 0, "Max results")
	contactListCmd.Flags().String("cursor", "", "Pagination cursor")
	contactListCmd.Flags().String("direction", "", "Pagination direction (before|after)")
	contactListCmd.Flags().String("query", "", "Filter text")
	contactCmd.AddCommand(contactSearchCmd, contactListCmd)
}
