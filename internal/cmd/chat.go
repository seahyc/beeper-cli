package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Chat commands",
}

var chatListCmd = &cobra.Command{
	Use:   "list",
	Short: "List chats",
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
		if v, _ := cmd.Flags().GetString("account"); v != "" {
			params.Set("accountID", v)
		}
		if v, _ := cmd.Flags().GetString("type"); v != "" {
			params.Set("chatType", v)
		}
		if v, _ := cmd.Flags().GetBool("unread"); v {
			params.Set("unread", "true")
		}
		if v, _ := cmd.Flags().GetBool("inbox"); v {
			params.Set("inbox", "true")
		}
		if v, _ := cmd.Flags().GetBool("muted"); v {
			params.Set("muted", "true")
		}
		path := "/v1/chats"
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
		var result interface{}
		if err := client.Get(path, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var chatGetCmd = &cobra.Command{
	Use:   "get <chatID>",
	Short: "Get chat details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		params := url.Values{}
		if v, _ := cmd.Flags().GetInt("max-participants"); v > 0 {
			params.Set("maxParticipants", fmt.Sprintf("%d", v))
		}
		path := fmt.Sprintf("/v1/chats/%s", args[0])
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
		var result interface{}
		if err := client.Get(path, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var chatSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search chats",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		params := url.Values{}
		params.Set("query", args[0])
		if v, _ := cmd.Flags().GetString("scope"); v != "" {
			params.Set("scope", v)
		}
		if v, _ := cmd.Flags().GetString("type"); v != "" {
			params.Set("chatType", v)
		}
		if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
			params.Set("limit", fmt.Sprintf("%d", v))
		}
		if v, _ := cmd.Flags().GetString("cursor"); v != "" {
			params.Set("cursor", v)
		}
		if v, _ := cmd.Flags().GetString("direction"); v != "" {
			params.Set("direction", v)
		}
		if v, _ := cmd.Flags().GetString("account"); v != "" {
			params.Set("accountID", v)
		}
		if v, _ := cmd.Flags().GetBool("inbox"); v {
			params.Set("inbox", "true")
		}
		if v, _ := cmd.Flags().GetBool("unread"); v {
			params.Set("unread", "true")
		}
		if v, _ := cmd.Flags().GetBool("muted"); v {
			params.Set("muted", "true")
		}
		path := "/v1/chats/search?" + params.Encode()
		var result interface{}
		if err := client.Get(path, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var chatCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new chat",
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		accountID, _ := cmd.Flags().GetString("account")
		if accountID == "" {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("--account is required"))
		}

		body := map[string]interface{}{
			"accountID": accountID,
		}
		if v, _ := cmd.Flags().GetString("mode"); v != "" {
			body["mode"] = v
		}
		if v, _ := cmd.Flags().GetString("type"); v != "" {
			body["type"] = v
		}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			body["title"] = v
		}
		if v, _ := cmd.Flags().GetString("participants"); v != "" {
			body["participantIDs"] = strings.Split(v, ",")
		}
		phone, _ := cmd.Flags().GetString("phone")
		email, _ := cmd.Flags().GetString("email")
		username, _ := cmd.Flags().GetString("username")
		name, _ := cmd.Flags().GetString("name")
		if phone != "" || email != "" || username != "" || name != "" {
			user := map[string]string{}
			if phone != "" {
				user["phoneNumber"] = phone
			}
			if email != "" {
				user["email"] = email
			}
			if username != "" {
				user["username"] = username
			}
			if name != "" {
				user["fullName"] = name
			}
			body["user"] = user
		}
		if v, _ := cmd.Flags().GetString("message"); v != "" {
			body["messageText"] = v
		}
		if v, _ := cmd.Flags().GetBool("allow-invite"); v {
			body["allowInvite"] = true
		}

		var result interface{}
		if err := client.Post("/v1/chats", body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var chatArchiveCmd = &cobra.Command{
	Use:   "archive <chatID>",
	Short: "Archive a chat",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/archive", args[0])
		body := map[string]interface{}{"archived": true}
		var result interface{}
		if err := client.Post(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var chatUnarchiveCmd = &cobra.Command{
	Use:   "unarchive <chatID>",
	Short: "Unarchive a chat",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/archive", args[0])
		body := map[string]interface{}{"archived": false}
		var result interface{}
		if err := client.Post(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

func init() {
	chatListCmd.Flags().Int("limit", 0, "Max results")
	chatListCmd.Flags().String("cursor", "", "Pagination cursor")
	chatListCmd.Flags().String("direction", "", "Pagination direction (before|after)")
	chatListCmd.Flags().String("account", "", "Filter by account ID")
	chatListCmd.Flags().String("type", "", "Filter by chat type")
	chatListCmd.Flags().Bool("unread", false, "Only unread chats")
	chatListCmd.Flags().Bool("inbox", false, "Only inbox chats")
	chatListCmd.Flags().Bool("muted", false, "Only muted chats")

	chatGetCmd.Flags().Int("max-participants", 0, "Max participants to return")

	chatSearchCmd.Flags().String("scope", "", "Search scope (titles|all)")
	chatSearchCmd.Flags().String("type", "", "Filter by chat type")
	chatSearchCmd.Flags().Int("limit", 0, "Max results")
	chatSearchCmd.Flags().String("cursor", "", "Pagination cursor")
	chatSearchCmd.Flags().String("direction", "", "Pagination direction")
	chatSearchCmd.Flags().String("account", "", "Filter by account ID")
	chatSearchCmd.Flags().Bool("inbox", false, "Only inbox chats")
	chatSearchCmd.Flags().Bool("unread", false, "Only unread chats")
	chatSearchCmd.Flags().Bool("muted", false, "Only muted chats")

	chatCreateCmd.Flags().String("account", "", "Account ID (required)")
	chatCreateCmd.Flags().String("mode", "", "Chat mode")
	chatCreateCmd.Flags().String("type", "", "Chat type")
	chatCreateCmd.Flags().String("title", "", "Chat title")
	chatCreateCmd.Flags().String("participants", "", "Comma-separated participant IDs")
	chatCreateCmd.Flags().String("phone", "", "User phone number")
	chatCreateCmd.Flags().String("email", "", "User email")
	chatCreateCmd.Flags().String("username", "", "Username")
	chatCreateCmd.Flags().String("name", "", "Full name")
	chatCreateCmd.Flags().String("message", "", "Initial message text")
	chatCreateCmd.Flags().Bool("allow-invite", false, "Allow invite")

	chatCmd.AddCommand(chatListCmd, chatGetCmd, chatSearchCmd, chatCreateCmd, chatArchiveCmd, chatUnarchiveCmd)
}
