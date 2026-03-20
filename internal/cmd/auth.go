package cmd

import (
	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/auth"
	"github.com/yjwong/beeper-cli/internal/output"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check auth status",
	Run: func(cmd *cobra.Command, args []string) {
		token := auth.GetToken()
		status := map[string]interface{}{
			"authenticated": token.IsValid(),
			"expires_at":    token.ExpiresAt,
		}
		if token.IsValid() {
			client := api.NewClient(getBaseURL())
			var info interface{}
			if err := client.Get("/v1/info", &info); err != nil {
				status["api_reachable"] = false
				status["api_error"] = err.Error()
			} else {
				status["api_reachable"] = true
			}
		}
		output.JSON(status)
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Revoke token and log out",
	Run: func(cmd *cobra.Command, args []string) {
		if err := auth.RevokeToken(getBaseURL()); err != nil {
			output.Fatal("LOGOUT_ERROR", err)
		}
		output.JSON(map[string]interface{}{"success": true, "message": "Logged out"})
	},
}

func init() {
	authCmd.AddCommand(authStatusCmd, authLogoutCmd)
}
