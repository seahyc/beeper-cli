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

func patchChat(chatID string, body map[string]interface{}) {
	client := api.NewClient(getBaseURL())
	path := fmt.Sprintf("/v1/chats/%s", chatID)
	var result interface{}
	if err := client.Patch(path, body, &result); err != nil {
		output.Fatal("API_ERROR", err)
	}
	output.JSON(result)
}

var chatLowPriorityCmd = &cobra.Command{
	Use:   "low-priority <chatID>",
	Short: "Mark a chat as low priority",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"isLowPriority": true})
	},
}

var chatUnlowPriorityCmd = &cobra.Command{
	Use:   "unlow-priority <chatID>",
	Short: "Unmark a chat as low priority",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"isLowPriority": false})
	},
}

var chatMuteCmd = &cobra.Command{
	Use:   "mute <chatID>",
	Short: "Mute a chat",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"isMuted": true})
	},
}

var chatUnmuteCmd = &cobra.Command{
	Use:   "unmute <chatID>",
	Short: "Unmute a chat",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"isMuted": false})
	},
}

var chatPinCmd = &cobra.Command{
	Use:   "pin <chatID>",
	Short: "Pin a chat",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"isPinned": true})
	},
}

var chatUnpinCmd = &cobra.Command{
	Use:   "unpin <chatID>",
	Short: "Unpin a chat",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"isPinned": false})
	},
}

var chatReadCmd = &cobra.Command{
	Use:   "read <chatID>",
	Short: "Mark a chat as read",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/read", args[0])
		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("message"); v != "" {
			body["messageID"] = v
		}
		var result interface{}
		if err := client.Post(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var chatUnreadCmd = &cobra.Command{
	Use:   "unread <chatID>",
	Short: "Mark a chat as unread",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/unread", args[0])
		body := map[string]interface{}{}
		if v, _ := cmd.Flags().GetString("message"); v != "" {
			body["messageID"] = v
		}
		var result interface{}
		if err := client.Post(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var chatNotifyAnywayCmd = &cobra.Command{
	Use:   "notify-anyway <chatID>",
	Short: "Force a delivery notification (iMessage/macOS only)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/notify-anyway", args[0])
		var result interface{}
		if err := client.Post(path, map[string]interface{}{}, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var chatSetTitleCmd = &cobra.Command{
	Use:   "set-title <chatID> <title>",
	Short: "Set a custom chat title",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"title": args[1]})
	},
}

var chatClearTitleCmd = &cobra.Command{
	Use:   "clear-title <chatID>",
	Short: "Clear the custom chat title",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"title": nil})
	},
}

var chatSetDescriptionCmd = &cobra.Command{
	Use:   "set-description <chatID> <description>",
	Short: "Set a group chat description/topic",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"description": args[1]})
	},
}

var chatClearDescriptionCmd = &cobra.Command{
	Use:   "clear-description <chatID>",
	Short: "Clear a group chat description/topic",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"description": nil})
	},
}

var chatSetImageCmd = &cobra.Command{
	Use:   "set-image <chatID> <imagePath>",
	Short: "Set a group chat avatar (local filesystem path)",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"imgURL": args[1]})
	},
}

var chatClearImageCmd = &cobra.Command{
	Use:   "clear-image <chatID>",
	Short: "Clear the group chat avatar",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"imgURL": nil})
	},
}

var chatSetExpiryCmd = &cobra.Command{
	Use:   "set-expiry <chatID> <seconds>",
	Short: "Set the disappearing-message timer in seconds",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		var secs int
		if _, err := fmt.Sscanf(args[1], "%d", &secs); err != nil || secs < 0 {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("seconds must be a non-negative integer"))
		}
		patchChat(args[0], map[string]interface{}{"messageExpirySeconds": secs})
	},
}

var chatClearExpiryCmd = &cobra.Command{
	Use:   "clear-expiry <chatID>",
	Short: "Clear the disappearing-message timer",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		patchChat(args[0], map[string]interface{}{"messageExpirySeconds": nil})
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

	chatReadCmd.Flags().String("message", "", "Optional message ID to mark read through")
	chatUnreadCmd.Flags().String("message", "", "Optional message ID to mark unread from")

	chatCmd.AddCommand(
		chatListCmd, chatGetCmd, chatSearchCmd, chatCreateCmd,
		chatArchiveCmd, chatUnarchiveCmd,
		chatLowPriorityCmd, chatUnlowPriorityCmd,
		chatMuteCmd, chatUnmuteCmd,
		chatPinCmd, chatUnpinCmd,
		chatReadCmd, chatUnreadCmd,
		chatNotifyAnywayCmd,
		chatSetTitleCmd, chatClearTitleCmd,
		chatSetDescriptionCmd, chatClearDescriptionCmd,
		chatSetImageCmd, chatClearImageCmd,
		chatSetExpiryCmd, chatClearExpiryCmd,
	)
}
