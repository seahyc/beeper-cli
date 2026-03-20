package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Account commands",
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "List connected accounts",
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		var result interface{}
		if err := client.Get("/v1/accounts", &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

func init() {
	accountCmd.AddCommand(accountListCmd)
}
