package cmd

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yjwong/beeper-cli/internal/api"
	"github.com/yjwong/beeper-cli/internal/output"
)

var msgCmd = &cobra.Command{
	Use:   "msg",
	Short: "Message commands",
}

// encodeChatID URL-encodes a chat ID for use in URL paths.
// Beeper chat IDs contain '!' and ':' which must be percent-encoded.
func encodeChatID(id string) string {
	return url.PathEscape(id)
}

type localMediaMetadata struct {
	FileName string
	MimeType string
	Width    int
	Height   int
}

func sniffLocalMediaMetadata(filePath string) (*localMediaMetadata, error) {
	meta := &localMediaMetadata{
		FileName: filepath.Base(filePath),
		MimeType: mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath))),
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	header := make([]byte, 512)
	n, err := f.Read(header)
	if err != nil && err.Error() != "EOF" {
		return nil, err
	}
	header = header[:n]

	if detected := http.DetectContentType(header); detected != "" && detected != "application/octet-stream" {
		meta.MimeType = detected
	}
	if meta.MimeType == "" {
		meta.MimeType = "application/octet-stream"
	}

	if _, err := f.Seek(0, 0); err == nil {
		if cfg, _, err := image.DecodeConfig(f); err == nil {
			meta.Width = cfg.Width
			meta.Height = cfg.Height
		}
	}

	return meta, nil
}

var msgListCmd = &cobra.Command{
	Use:   "list <chatID>",
	Short: "List messages in a chat",
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
		path := fmt.Sprintf("/v1/chats/%s/messages", encodeChatID(args[0]))
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

var msgSendCmd = &cobra.Command{
	Use:   "send <chatID>",
	Short: "Send a message",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		text, _ := cmd.Flags().GetString("text")
		filePath, _ := cmd.Flags().GetString("file")
		replyTo, _ := cmd.Flags().GetString("reply-to")

		body := map[string]interface{}{}
		if text != "" {
			body["text"] = text
		}
		if replyTo != "" {
			body["replyToMessageID"] = replyTo
		}

		if filePath != "" {
			localMeta, err := sniffLocalMediaMetadata(filePath)
			if err != nil {
				output.Fatal("VALIDATION_ERROR", fmt.Errorf("failed to inspect media file: %w", err))
			}

			var uploadResult struct {
				UploadID string `json:"uploadID"`
				FileName string `json:"fileName"`
				MimeType string `json:"mimeType"`
				Width    int    `json:"width"`
				Height   int    `json:"height"`
				Duration int    `json:"duration"`
			}
			if err := client.UploadFile("/v1/assets/upload", filePath, &uploadResult); err != nil {
				output.Fatal("UPLOAD_ERROR", err)
			}

			attachment := map[string]interface{}{
				"uploadID": uploadResult.UploadID,
			}

			fileName := uploadResult.FileName
			if fileName == "" {
				fileName = localMeta.FileName
			}
			if fileName != "" {
				attachment["fileName"] = fileName
			}

			mimeType := uploadResult.MimeType
			if mimeType == "" || mimeType == "application/octet-stream" {
				mimeType = localMeta.MimeType
			}
			if mimeType != "" {
				attachment["mimeType"] = mimeType
			}

			width := uploadResult.Width
			height := uploadResult.Height
			if width <= 0 || height <= 0 {
				width = localMeta.Width
				height = localMeta.Height
			}
			if width > 0 || height > 0 {
				attachment["size"] = map[string]int{"width": width, "height": height}
			}
			if v, _ := cmd.Flags().GetString("attach-type"); v != "" {
				attachment["type"] = v
			}
			if v, _ := cmd.Flags().GetString("filename"); v != "" {
				attachment["fileName"] = v
			}
			if v, _ := cmd.Flags().GetString("mime"); v != "" {
				attachment["mimeType"] = v
			}
			body["attachment"] = attachment
		}

		if len(body) == 0 {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("--text or --file is required"))
		}

		path := fmt.Sprintf("/v1/chats/%s/messages", encodeChatID(args[0]))
		var result interface{}
		if err := client.Post(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var msgEditCmd = &cobra.Command{
	Use:   "edit <chatID> <msgID>",
	Short: "Edit a message",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		text, _ := cmd.Flags().GetString("text")
		if text == "" {
			output.Fatal("VALIDATION_ERROR", fmt.Errorf("--text is required"))
		}
		path := fmt.Sprintf("/v1/chats/%s/messages/%s", encodeChatID(args[0]), url.PathEscape(args[1]))
		body := map[string]interface{}{"text": text}
		var result interface{}
		if err := client.Put(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var msgSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search messages",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		params := url.Values{}
		params.Set("query", args[0])
		if v, _ := cmd.Flags().GetString("chat"); v != "" {
			params.Set("chatID", v)
		}
		if v, _ := cmd.Flags().GetString("account"); v != "" {
			params.Set("accountID", v)
		}
		if v, _ := cmd.Flags().GetString("chat-type"); v != "" {
			params.Set("chatType", v)
		}
		if v, _ := cmd.Flags().GetString("sender"); v != "" {
			params.Set("senderID", v)
		}
		if v, _ := cmd.Flags().GetBool("media"); v {
			params.Set("hasMedia", "true")
		}
		if v, _ := cmd.Flags().GetString("after"); v != "" {
			params.Set("after", v)
		}
		if v, _ := cmd.Flags().GetString("before"); v != "" {
			params.Set("before", v)
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
		if v, _ := cmd.Flags().GetBool("exclude-low-priority"); v {
			params.Set("excludeLowPriority", "true")
		}
		if v, _ := cmd.Flags().GetBool("muted"); v {
			params.Set("muted", "true")
		}
		path := "/v1/messages/search?" + params.Encode()
		var result interface{}
		if err := client.Get(path, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var msgReactCmd = &cobra.Command{
	Use:   "react <chatID> <msgID> <emoji>",
	Short: "React to a message",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/messages/%s/reactions", encodeChatID(args[0]), url.PathEscape(args[1]))
		body := map[string]interface{}{"reactionKey": args[2]}
		var result interface{}
		if err := client.Post(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

var msgUnreactCmd = &cobra.Command{
	Use:   "unreact <chatID> <msgID> <emoji>",
	Short: "Remove reaction from a message",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		client := api.NewClient(getBaseURL())
		path := fmt.Sprintf("/v1/chats/%s/messages/%s/reactions", encodeChatID(args[0]), url.PathEscape(args[1]))
		body := map[string]interface{}{"reactionKey": args[2]}
		var result interface{}
		if err := client.DeleteWithBody(path, body, &result); err != nil {
			output.Fatal("API_ERROR", err)
		}
		output.JSON(result)
	},
}

func init() {
	msgListCmd.Flags().Int("limit", 0, "Max results")
	msgListCmd.Flags().String("cursor", "", "Pagination cursor")
	msgListCmd.Flags().String("direction", "", "Pagination direction (before|after)")

	msgSendCmd.Flags().String("text", "", "Message text")
	msgSendCmd.Flags().String("file", "", "File path to upload and attach")
	msgSendCmd.Flags().String("reply-to", "", "Message ID to reply to")
	msgSendCmd.Flags().String("attach-type", "", "Attachment type override")
	msgSendCmd.Flags().String("filename", "", "Override attachment filename")
	msgSendCmd.Flags().String("mime", "", "Override attachment MIME type")

	msgEditCmd.Flags().String("text", "", "New message text")

	msgSearchCmd.Flags().String("chat", "", "Filter by chat ID")
	msgSearchCmd.Flags().String("account", "", "Filter by account ID")
	msgSearchCmd.Flags().String("chat-type", "", "Filter by chat type")
	msgSearchCmd.Flags().String("sender", "", "Filter by sender ID")
	msgSearchCmd.Flags().Bool("media", false, "Only messages with media")
	msgSearchCmd.Flags().String("after", "", "Messages after timestamp")
	msgSearchCmd.Flags().String("before", "", "Messages before timestamp")
	msgSearchCmd.Flags().Int("limit", 0, "Max results")
	msgSearchCmd.Flags().String("cursor", "", "Pagination cursor")
	msgSearchCmd.Flags().String("direction", "", "Pagination direction")
	msgSearchCmd.Flags().Bool("exclude-low-priority", false, "Exclude low priority")
	msgSearchCmd.Flags().Bool("muted", false, "Only muted chats")

	msgCmd.AddCommand(msgListCmd, msgSendCmd, msgEditCmd, msgSearchCmd, msgReactCmd, msgUnreactCmd)
}
