# beeper-cli

CLI for the local [Beeper Desktop](https://www.beeper.com/) HTTP API.

`beeper-cli` lets you search chats, read messages, send messages with attachments, react/edit/delete, manage reminders, upload/download media, and work across connected messaging accounts from the terminal.

It talks to the Beeper Desktop app running on your machine. It is not a hosted service and does not work unless Beeper Desktop is running locally.

## What It Can Do

- Authenticate against Beeper Desktop via OAuth
- List connected messaging accounts
- Search contacts on connected accounts
- List chats, inspect chat metadata, create chats, archive/unarchive chats
- Read messages in a chat
- Search messages across chats
- List poll messages and parse their question/options from chat history
- Send text messages and messages with file attachments
- Reply to a specific message
- Edit sent messages
- Delete sent messages
- React and unreact to messages
- Upload files and base64 content as Beeper assets
- Download cached media assets
- Render encrypted media assets into usable local files
- Set and clear chat reminders
- Focus Beeper Desktop on a chat or message

Supported account types depend on what you have connected in Beeper Desktop, for example WhatsApp, Telegram, Signal, Discord, Slack, Instagram, iMessage, and others exposed by Beeper.

## How It Works

The CLI uses Beeper Desktop's local API, which defaults to:

```text
http://localhost:23373
```

You can override that with either:

- `--url http://host:port`
- `BEEPER_URL=http://host:port`

Authentication is local and automatic:

- the first authenticated command triggers browser OAuth
- tokens are cached in `~/.beeper/token.json`

## Installation

### Build

```bash
git clone git@github.com:seahyc/beeper-cli.git
cd beeper-cli
make build
```

This builds `./beeper`.

### Install to `~/.local/bin`

```bash
make install
```

## Requirements

- Beeper Desktop must be running
- You must already be logged into Beeper Desktop
- At least one messaging account should be connected in Beeper

## Quick Start

Check that the local API is reachable:

```bash
beeper info
```

Check auth status:

```bash
beeper auth status
```

List accounts:

```bash
beeper account list
```

List recent chats:

```bash
beeper chat list --limit 10
```

Read recent messages in a chat:

```bash
beeper msg list "!chatID:beeper.local" --limit 10
```

Send a message:

```bash
beeper msg send "!chatID:beeper.local" --text "Hello!"
```

Send a message with an image:

```bash
beeper msg send "!chatID:beeper.local" --text "See attached" --file ./photo.jpg
```

List recent polls in a chat:

```bash
beeper poll list "!chatID:beeper.local" --limit 10
```

Reply to a specific message:

```bash
beeper msg send "!chatID:beeper.local" --text "Following up on this" --reply-to "73188"
```

## Command Areas

### Server / Auth

- `beeper info`
- `beeper auth status`
- `beeper auth logout`

### Accounts / Contacts

- `beeper account list`
- `beeper contact search <accountID> <query>`
- `beeper contact list <accountID>`

### Chats

- `beeper chat list`
- `beeper chat get "!chatID:beeper.local"`
- `beeper chat search "query"`
- `beeper chat create --account <id> --type single ...`
- `beeper chat create --account <id> --type group ...`
- `beeper chat archive "!chatID:beeper.local"`
- `beeper chat unarchive "!chatID:beeper.local"`
- `beeper chat pinned "!chatID:beeper.local"`

### Messages

- `beeper msg list "!chatID:beeper.local" --limit 20`
- `beeper msg search "keyword" --chat "!chatID:beeper.local" --limit 20`
- `beeper msg search "keyword" --chat "!chatID:beeper.local" --local --pages 40`
- `beeper msg send "!chatID:beeper.local" --text "Hello"`
- `beeper msg edit "!chatID:beeper.local" "msgID" --text "Edited"`
- `beeper msg delete "!chatID:beeper.local" "msgID"`
- `beeper msg react "!chatID:beeper.local" "msgID" '👍'`
- `beeper msg unreact "!chatID:beeper.local" "msgID" '👍'`

### Polls

- `beeper poll list "!chatID:beeper.local" --limit 10`
- `beeper poll list "!chatID:beeper.local" --pages 20 --page-size 100`

### Assets / Media

- `beeper asset upload ./file.jpg`
- `beeper asset upload-base64 --content <b64> --filename img.png --mime image/png`
- `beeper asset download <mxc://url>`
- `beeper asset download <mxc://url> --output ./downloads/`
- `beeper asset serve --url <mxc://url>`
- `beeper asset serve --url <mxc://url> --output ./preview.jpg`

### Reminders / Focus

- `beeper reminder set "!chatID:beeper.local" --at "2026-05-04T10:00:00Z"`
- `beeper reminder clear "!chatID:beeper.local"`
- `beeper focus --chat "!chatID:beeper.local"`
- `beeper focus --chat "!chatID:beeper.local" --message "msgID"`

## Media Behavior

There are two useful asset flows:

### `asset download`

`asset download` resolves a Beeper asset into the local cached file that Beeper Desktop already has on disk.

Example:

```bash
beeper asset download "mxc://beeper.local/abc123"
```

Typical output:

```json
{
  "url": "file:///Users/you/Library/Application%20Support/BeeperTexts/media/...",
  "sourcePath": "/Users/you/Library/Application Support/BeeperTexts/media/...",
  "savedPath": "/Users/you/Library/Application Support/BeeperTexts/media/..."
}
```

If you pass `--output`, it copies the cached file to your chosen path.

```bash
beeper asset download "mxc://beeper.local/abc123" --output ./downloads/
```

### `asset serve`

`asset serve` fetches the rendered media bytes from Beeper's serve endpoint and writes them to a local file.

This is especially useful when the message attachment exposes an encrypted `mxc://...?...encryptedFileInfoJSON=...` URL and you want a directly usable file.

```bash
beeper asset serve --url "mxc://beeper.local/abc123?encryptedFileInfoJSON=..." --output ./preview.jpg
```

Typical output:

```json
{
  "contentType": "image/jpeg",
  "savedPath": "/tmp/beeper-asset-12345.jpg",
  "size": 75600
}
```

This flow has been verified for:

- images
- encrypted images
- PDFs/files
- encrypted WhatsApp video

## Poll Behavior

`beeper poll list` scans chat history and extracts poll question/options from messages where Beeper Desktop renders a WhatsApp poll as HTML text.

Example:

```bash
beeper poll list "!chatID:beeper.local" --limit 5
```

Typical output includes:

```json
{
  "messageID": "94457",
  "question": "RSVP: who has confirmed?",
  "options": ["Alice", "Bob"],
  "source": "beeper_api_text",
  "voteCountsAvailable": false
}
```

Beeper Desktop's local API currently exposes WhatsApp polls as text only. It does not expose native poll creation, live vote counts, or voter names. Use WhatsApp Web or the native WhatsApp app for those operations.

## Notes and Caveats

- Chat IDs look like `!abc123:beeper.local`
- The CLI URL-encodes Beeper path identifiers internally; quoting chat IDs in shell commands is still recommended
- `msg search --limit` is effectively capped by the Beeper Desktop API at `20`; chat-scoped searches fall back to local history scanning when the API rejects the query or ignores `--chat`
- Use `msg search --local --chat ... --pages N` when you need deterministic scoped search over older chat history
- `chat pinned` probes known Beeper Desktop pinned-message endpoints and reports `supported: false` when the local API does not expose pinned messages
- `poll list` can parse poll question/options from Beeper message text, but native poll creation and live vote counts are not exposed by Beeper Desktop's local API
- `msg edit`, `msg delete`, `msg react`, and `msg unreact` should use the numeric message ID from `msg list`
- `asset download` understands both Beeper response shapes: `url` and `srcURL`
- `asset serve` returns a local file path, not a remote URL
- If you omit `--output` for `asset serve`, it writes to a temp file

## Development

Build locally:

```bash
make build
```

Install locally:

```bash
make install
```

Clean local binary:

```bash
make clean
```
