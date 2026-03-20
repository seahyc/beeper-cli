package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/output"
)

var baseURL string

var rootCmd = &cobra.Command{
	Use:           "beeper",
	Short:         "Beeper Desktop API CLI",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		output.Fatal("COMMAND_ERROR", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&baseURL, "url", "http://localhost:23373", "Beeper Desktop API base URL")
	rootCmd.AddCommand(versionCmd, infoCmd, authCmd, accountCmd, contactCmd, chatCmd, msgCmd, searchCmd, assetCmd, reminderCmd, focusCmd)
	if envURL := os.Getenv("BEEPER_URL"); envURL != "" && !rootCmd.PersistentFlags().Changed("url") {
		baseURL = envURL
	}
}

func getBaseURL() string {
	return baseURL
}
