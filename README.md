# bsky

bluesky CLI client written in Go

## Usage

```
NAME:
   bsky - bsky

USAGE:
   bsky [global options] command [command options] [arguments...]

VERSION:
   0.0.6

DESCRIPTION:
   A cli application for bluesky

COMMANDS:
   show-profile    show profile
   update-profile  update profile
   timeline, tl    show timeline
   search          search posts
   thread          show thread
   post            post new text
   vote            vote the post
   votes           show votes of the post
   repost          repost the post
   reposts         show reposts of the post
   follow          follow the handle
   follows         show follows
   followers       show followers
   delete          delete the note
   login           login the social
   help, h         Shows a list of commands or help for one command

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

## Installation

Download binary from Release page.

Or install with go install command.
```
go install github.com/mattn/bsky@latest
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


## Development

```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
