# bsky

bluesky CLI client written in Go

## Usage

```
NAME:
   bsky - bsky

USAGE:
   bsky [global options] command [command options]

VERSION:
   0.0.67

DESCRIPTION:
   A cli application for bluesky

COMMANDS:
   show-profile         Show profile
   update-profile       Update profile
   show-session         Show session
   timeline, tl         Show timeline
   stream               Show timeline as stream
   thread               Show thread
   post                 Post new text
   vote                 Vote the post
   votes                Show votes of the post
   repost               Repost the post
   reposts              Show reposts of the post
   follow               Follow the handle
   unfollow             Unfollow the handle
   follows              Show follows
   followers            Show followers
   block                Block the handle
   unblock              Unblock the handle
   blocks               Show blocks
   delete               Delete the note
   search               Search Bluesky
   login                Login the social
   notification, notif  Show notifications
   invite-codes         Show invite codes
   list-app-passwords   Show App-passwords
   add-app-password     Add App-password
   revoke-app-password  Revoke App-password
   mcp                  Start MCP server
   help, h              Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -a value       profile name
   -V             verbose (default: false)
   --help, -h     show help
   --version, -v  print the version
```

```
$ bsky login [handle] [password]
$ bsky timeline
```

```
$ bsky post -image ~/pizza.jpg 'I love 🍕'
```

```
$ bsky vote at://did:plc:xxxxxxxxxxxxxxxxxxxxxxxx/app.bsky.feed.post/yyyyyyyyyyyyy
$ bsky repost at://did:plc:xxxxxxxxxxxxxxxxxxxxxxxx/app.bsky.feed.post/yyyyyyyyyyyyy
```

### Extended Usage Information

Individual commands have their own help texts. Call via `-h` / `--help` and the name of the command.

### JSON Output

The output for most commands can be formatted as JSON via `--json`. See Extended Usage Information for the individual commands that support JSON output.

## Installation

Download binary from Release page.

Or install with go install command.
```
go install github.com/mattn/bsky@latest
```

### MCP Server

`bsky mcp` starts an MCP (Model Context Protocol) server over stdio. This allows AI assistants like Claude to interact with Bluesky directly.

Available tools:

| Tool | Description |
|------|-------------|
| `bluesky_timeline` | Show timeline posts |
| `bluesky_post` | Create a new post |
| `bluesky_search` | Search posts |
| `bluesky_show_profile` | Show a user profile |
| `bluesky_search_actors` | Search for users |
| `bluesky_thread` | Show a post thread |
| `bluesky_notification` | Show notifications |
| `bluesky_like` | Like a post |
| `bluesky_repost` | Repost a post |
| `bluesky_follow` | Follow a user |

Example configuration for Claude Code (`~/.claude/settings.json`):

```json
{
  "mcpServers": {
    "bsky": {
      "command": "bsky",
      "args": ["mcp"]
    }
  }
}
```

### To enable Autocomplete

Download the correct file from `/scripts` directory and add the following line to your shell configuration file.

ZSH:
```sh
# Add the following line to your .zshrc
source /path/to/autocomplete.zsh
```

Bash:
```bash
# Add the following line to your .bashrc
source /path/to/autocomplete.sh
```

PowerShell:
```powershell
# Add the following line to your $PROFILE
/path/to/autocomplete.ps1
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
