# bsky

bluesky CLI client written in Go

## Usage

```

NAME:
   bsky - A new cli application

USAGE:
   bsky [global options] command [command options] [arguments...]

DESCRIPTION:
   A cli application for bluesky

COMMANDS:
   show-profile    show profile
   update-profile  update profile
   timeline, tl    show timeline
   show-post       show the post
   thread          show thread
   post            post new text
   vote            vote the post
   votes           show votes of the post
   repost          repost the post
   reposts         show reposts of the post
   delete          delete the note
   login           login the social
   help, h         Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -a value    profile name
   -V          verbose (default: false)
   --help, -h  show help
```

## Installation

Download binary from Release page.

Or install with go install command.
```
go install github.com/mattn/bsky@latest
```

## License

MIT

## Author

Yasuhiro Matsumoto (a.k.a. mattn)
