package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Server/app metadata",
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		var result interface{}
		if err := client.GetUnauthenticated("/v1/info", &result); err != nil {
			output.Fatal("CONNECTION_ERROR", err)
		}
		output.JSON(result)
	},
}
