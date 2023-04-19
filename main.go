package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const name = "bsky"

const version = "0.0.27"

var revision = "HEAD"

type config struct {
	Host     string `json:"host"`
	Handle   string `json:"handle"`
	Password string `json:"password"`
	dir      string
	verbose  bool
}

func main() {
	app := &cli.App{
		Name:        name,
		Usage:       name,
		Version:     version,
		Description: "A cli application for bluesky",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "a", Usage: "profile name"},
			&cli.BoolFlag{Name: "V", Usage: "verbose"},
		},
		Commands: []*cli.Command{
			{
				Name:        "show-profile",
				Description: "Show profile",
				Usage:       "Show profile",
				UsageText:   "bsky show-profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				Action: doShowProfile,
			},
			{
				Name:        "update-profile",
				Description: "Update profile",
				Usage:       "Update profile",
				UsageText:   "bsky update-profile [OPTIONS]... [{display name} [description]]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "avatar", Value: "", Usage: "avatar image", TakesFile: true},
					&cli.StringFlag{Name: "banner", Value: "", Usage: "banner image", TakesFile: true},
				},
				Action: doUpdateProfile,
			},
			{
				Name:        "show-session",
				Description: "Show session",
				Usage:       "Show session",
				UsageText:   "bsky show-session",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				Action: doShowSession,
			},
			{
				Name:        "timeline",
				Description: "Show timeline",
				Usage:       "Show timeline",
				UsageText:   "bsky timeline",
				Aliases:     []string{"tl"},
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
					&cli.IntFlag{Name: "n", Value: 30, Usage: "number of items"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				Action: doTimeline,
			},
			{
				Name:        "stream",
				Description: "Show timeline as stream",
				Usage:       "Show timeline as stream",
				UsageText:   "bsky stream",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				Action: doStream,
			},
			{
				Name:        "thread",
				Description: "Show thread",
				Usage:       "Show thread",
				UsageText:   "bsky thread [uri]",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "n", Value: 30, Usage: "number of items"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				Action: doThread,
			},
			{
				Name:        "post",
				Description: "Post new text",
				Usage:       "Post new text",
				UsageText:   "bsky post [text]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "r"},
					&cli.StringFlag{Name: "q"},
					&cli.BoolFlag{Name: "stdin"},
					&cli.StringSliceFlag{Name: "image", Aliases: []string{"i"}},
				},
				HelpName:  "post",
				ArgsUsage: "[text]",
				Action:    doPost,
			},
			{
				Name:        "vote",
				Description: "Vote the post",
				Usage:       "Vote the post",
				UsageText:   "bsky vote [uri]",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "down"},
				},
				HelpName: "vote",
				Action:   doVote,
			},
			{
				Name:        "votes",
				Description: "Show votes of the post",
				Usage:       "Show votes of the post",
				UsageText:   "bsky votes [uri]",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName:  "votes",
				Action:    doVotes,
				ArgsUsage: "[uri]",
			},
			{
				Name:        "repost",
				Description: "Repost the post",
				Usage:       "Repost the post",
				UsageText:   "bsky repost [uri]",
				HelpName:    "repost",
				Action:      doRepost,
			},
			{
				Name:        "reposts",
				Description: "Show reposts of the post",
				Usage:       "Show reposts of the post",
				UsageText:   "bsky reposts [uri]",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "reposts",
				Action:   doReposts,
			},
			{
				Name:        "follow",
				Description: "Follow the handle",
				Usage:       "Follow the handle",
				UsageText:   "bsky follow [handle]",
				HelpName:    "follow",
				Action:      doFollow,
			},
			{
				Name:        "follows",
				Description: "Show follows",
				Usage:       "Show follows",
				UsageText:   "bsky follows",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "follows",
				Action:   doFollows,
			},
			{
				Name:        "followers",
				Description: "Show followers",
				Usage:       "Show followers",
				UsageText:   "bsky followres",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "followers",
				Action:   doFollowers,
			},
			{
				Name:        "delete",
				Description: "Delete the note",
				Usage:       "Delete the note",
				UsageText:   "bsky delete [cid]",
				HelpName:    "delete",
				Action:      doDelete,
			},
			{
				Name:        "login",
				Description: "Login the social",
				Usage:       "Login the social",
				UsageText:   "bsky login [handle] [password]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "host", Value: "https://bsky.social"},
				},
				HelpName: "login",
				Action:   doLogin,
			},
			{
				Name:        "notification",
				Description: "Show notifications",
				Usage:       "Show notifications",
				UsageText:   "bsky notification",
				Aliases:     []string{"notif"},
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "a", Usage: "show all"},
				},
				HelpName: "notification",
				Action:   doNotification,
			},
		},
		Metadata: map[string]any{},
		Before: func(cCtx *cli.Context) error {
			profile := cCtx.String("a")
			cfg, fp, err := loadConfig(profile)
			cCtx.App.Metadata["path"] = fp
			if cCtx.Args().Get(0) == "login" {
				return nil
			}
			if err != nil {
				return fmt.Errorf("cannot load config file: %w", err)
			}
			cCtx.App.Metadata["config"] = cfg
			cfg.verbose = cCtx.Bool("V")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
