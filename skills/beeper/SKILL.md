---
name: beeper
description: >
  Send messages and search chats across WhatsApp, Telegram, Signal, Discord, Slack,
  Instagram, iMessage, LinkedIn, Facebook Messenger, Google Messages via Beeper Desktop API.
  Reactions, reminders, attachments, media download, contact search, edit/delete messages,
  create chats, unified search. Use when user asks about messages, chats, contacts,
  or wants to send/read/search messages across any messaging platform.
---

# Beeper CLI Skill

Send messages and manage chats across all messaging platforms via the `beeper` CLI wrapping Beeper Desktop's HTTP API.

## Capabilities and Use Cases

- Send text messages with optional file attachments across any platform
- Search messages and chats across all connected accounts
- List and manage connected accounts (WhatsApp, Telegram, Signal, Discord, Slack, iMessage, etc.)
- React/unreact to messages with emoji
- Edit/delete sent messages
- Create new chats (DM or group)
- Archive/unarchive, mute, pin, low-priority, mark read/unread, set title/description/avatar/expiry on chats
- Upload/download media assets
- Set/clear chat reminders
- Focus Beeper Desktop on a specific chat
- Search contacts on connected accounts
- All output is JSON

## Quick Reference

**Send message:**
```bash
beeper msg send "!chatID:beeper.local" --text "Hello!"
```

**Send with attachment:**
```bash
beeper msg send "!chatID:beeper.local" --text "See attached" --file ./photo.jpg
```

**List messages:**
```bash
beeper msg list "!chatID:beeper.local" --limit 10
```

**Search messages:**
```bash
beeper msg search "keyword" --chat "!chatID:beeper.local" --limit 20
beeper msg search "keyword" --chat "!chatID:beeper.local" --local --pages 40
```

**Download media to a specific file:**
```bash
beeper asset download "mxc://beeper.local/abc123" --output "./invoice.pdf"
```

**Download media into a directory using the original filename:**
```bash
beeper asset download "mxc://beeper.local/abc123" --output "./downloads/"
```

**React to message:**
```bash
beeper msg react "!chatID:beeper.local" "msgID" '👍'
```

**Remove reaction:**
```bash
beeper msg unreact "!chatID:beeper.local" "msgID"
```

**Edit message:**
```bash
beeper msg edit "!chatID:beeper.local" "msgID" --text "Updated text"
```

**Delete message:**
```bash
beeper msg delete "!chatID:beeper.local" "msgID"
```

## Commands Reference

### Server Info
```bash
beeper info                    # Server/app metadata (no auth required)
```

### Authentication
```bash
beeper auth status             # Check auth status and API reachability
beeper auth logout             # Revoke token and log out
```
Auth is automatic — first API call triggers browser OAuth flow.

### Accounts
```bash
beeper account list            # List all connected messaging accounts
```

### Contacts
```bash
beeper contact search <accountID> <query>    # Search contacts
beeper contact list <accountID>              # List contacts
  --limit 50 --cursor <cur> --direction after --query "name"
```

### Chats
```bash
beeper chat list --limit 10 --account <id> --type dm --unread --inbox
beeper chat get "!chatID:beeper.local" --max-participants 10
beeper chat search "query" --scope titles --type group --limit 10
beeper chat create --account <id> --type single --phone "+1234567890" --message "Hi"
beeper chat create --account <id> --type single --participants "userId" --message "Hi"
beeper chat create --account <id> --type group --title "Team" --participants "id1,id2"
beeper chat archive "!chatID:beeper.local"
beeper chat unarchive "!chatID:beeper.local"
beeper chat low-priority "!chatID:beeper.local"          # never floats to top, never badges
beeper chat unlow-priority "!chatID:beeper.local"
beeper chat mute "!chatID:beeper.local"
beeper chat unmute "!chatID:beeper.local"
beeper chat pin "!chatID:beeper.local"
beeper chat unpin "!chatID:beeper.local"
beeper chat pinned "!chatID:beeper.local"
beeper chat read "!chatID:beeper.local" --message <msgID>     # --message optional
beeper chat unread "!chatID:beeper.local" --message <msgID>   # --message optional
beeper chat notify-anyway "!chatID:beeper.local"         # iMessage on macOS only
beeper chat set-title "!chatID:beeper.local" "Custom title"
beeper chat clear-title "!chatID:beeper.local"
beeper chat set-description "!chatID:beeper.local" "Group topic"
beeper chat clear-description "!chatID:beeper.local"
beeper chat set-image "!chatID:beeper.local" ./avatar.png
beeper chat clear-image "!chatID:beeper.local"
beeper chat set-expiry "!chatID:beeper.local" 86400      # disappearing-message timer (seconds)
beeper chat clear-expiry "!chatID:beeper.local"
```

### Messages
```bash
beeper msg list "!chatID" --limit 20 --cursor <cur> --direction before
beeper msg send "!chatID" --text "Hello" --file ./image.png --attach-type image --filename custom.png --mime image/png
beeper msg edit "!chatID" "msgID" --text "Edited"
beeper msg delete "!chatID" "msgID"
beeper msg search "query" --chat "!chatID" --account <id> --sender <id> --media --after 2024-01-01 --before 2024-12-31 --limit 20
beeper msg search "query" --chat "!chatID" --local --pages 40 --page-size 100 --limit 20
beeper msg react "!chatID" "msgID" '🎉'
beeper msg unreact "!chatID" "msgID"
```

### Unified Search
```bash
beeper search "query"          # Search across chats and messages
```

### Assets
```bash
beeper asset upload ./file.jpg                    # Upload file
beeper asset upload-base64 --content <b64> --filename img.png --mime image/png
beeper asset download <mxc://url>                 # Resolve to Beeper's local cached file
beeper asset download <mxc://url> --output ./invoice.pdf
beeper asset download <mxc://url> --output ./downloads/
beeper asset serve --url <mxc://url>              # Fetch rendered bytes and save to a local temp file
beeper asset serve --url <mxc://url> --output ./preview.jpg
```

### Reminders
```bash
beeper reminder set "!chatID" --at "2024-03-20T10:00:00Z" --dismiss-on-reply
beeper reminder clear "!chatID"
```

### Focus
```bash
beeper focus --chat "!chatID"                     # Focus Beeper on chat
beeper focus --chat "!chatID" --message "msgID"   # Focus on specific message
beeper focus --draft "prefilled text"             # Open with draft text
beeper focus --draft-file ./message.txt           # Open with draft from file
```

## Common Workflows

### Find and message a contact
```bash
# 1. List accounts to find the right one
beeper account list
# 2. Search contacts on that account
beeper contact search "accountID" "John"
# 3. Create a DM (use --participants with the contact ID, or --phone/--username)
beeper chat create --account "accountID" --type single --participants "contactId" --message "Hey!"
```

### Download media from a chat
```bash
# 1. List messages to find one with attachment
beeper msg list "!chatID:beeper.local" --limit 5
# 2. Download the media URL from the message into a directory
beeper asset download "mxc://beeper.local/abc123" --output ./downloads/
```

### Render encrypted media into a viewable local file
```bash
# Useful when srcURL is an encrypted mxc:// URL and you want actual bytes on disk
beeper asset serve --url "mxc://beeper.local/abc123?encryptedFileInfoJSON=..." --output ./preview.jpg
```

### Send a photo with caption
```bash
beeper msg send "!chatID:beeper.local" --text "Check this out!" --file ./photo.jpg
```

## Error Handling

All errors return JSON:
```json
{
  "error": true,
  "code": "ERROR_CODE",
  "message": "Human-readable description"
}
```

Common codes:
- `CONNECTION_ERROR` — Beeper Desktop not running or unreachable
- `AUTH_ERROR` — Authentication failed; re-run any command to trigger OAuth
- `API_ERROR` — Server returned an error
- `VALIDATION_ERROR` — Invalid input (e.g., missing required flags)
- `UPLOAD_ERROR` — File upload failed

## Important Notes

- Chat IDs look like `!abc123:beeper.local` — the CLI handles URL-encoding internally, so both single and double quotes work in shell commands
- `beeper asset download` prints `url`, `sourcePath`, and `savedPath`; if you omit `--output`, `savedPath` is the cached local Beeper file
- `beeper asset download` understands both Beeper response shapes: `url` and `srcURL`
- `beeper asset serve` writes served bytes to disk and returns `savedPath`, `contentType`, and `size`; if `--output` is omitted, it uses a temp file
- `--output` accepts either an explicit file path or a directory path. If you want directory behavior for a new directory, end it with a trailing slash such as `./downloads/`
- `beeper msg search --limit` is effectively capped by the Beeper Desktop API at `20`; chat-scoped searches automatically fall back to local history scanning if the API rejects the query or ignores `--chat`
- Use `beeper msg search --local --chat ... --pages N` when a chat has important older context, pinned-message content, or multi-word terms that the Desktop API search handles poorly
- `beeper chat pinned` probes known Beeper Desktop pinned-message endpoints. If the app does not expose pinned messages through the local API, it returns `supported: false` with the endpoints tried.
- When editing or deleting messages, use the **numeric message ID** from `msg list` (e.g. `29463`), not the `pendingMessageID` returned by `msg send` (e.g. `~beeper-mautrix-go_...`). The pending ID is temporary and won't resolve for edits/deletes.
- The `--url` global flag or `BEEPER_URL` env var overrides the default `http://localhost:23373`
- Beeper Desktop must be running for any command to work
- First authenticated command will open browser for OAuth — subsequent commands use cached token
