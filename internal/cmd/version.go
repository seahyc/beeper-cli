package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/output"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		output.JSON(map[string]string{
			"version": version,
			"commit":  commit,
			"date":    date,
		})
	},
}
