package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const name = "bsky"

const version = "0.0.5"

var revision = "HEAD"

type Config struct {
	Host     string `json:"host"`
	Handle   string `json:"handle"`
	Password string `json:"password"`
	dir      string
	verbose  bool
}

func main() {

	app := &cli.App{
		//CustomAppHelpTemplate: HelpTemplate,
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
				Description: "show profile",
				Usage:       "show profile",
				UsageText:   "bsky show-profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
				},
				Action: doShowProfile,
			},
			{
				Name:        "update-profile",
				Description: "update profile",
				Usage:       "update profile",
				UsageText:   "bsky update-profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "avatar", Value: "", Usage: "avatar image", TakesFile: true},
					&cli.StringFlag{Name: "banner", Value: "", Usage: "banner image", TakesFile: true},
				},
				Action: doUpdateProfile,
			},
			{
				Name:        "timeline",
				Description: "show timeline",
				Usage:       "show timeline",
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
				Name:        "thread",
				Description: "show thread",
				Usage:       "show thread",
				UsageText:   "bsky thread [uri]",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "n", Value: 30, Usage: "number of items"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				Action: doThread,
			},
			{
				Name:        "post",
				Description: "post new text",
				Usage:       "post new text",
				UsageText:   "bsky post [text]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "r"},
					&cli.BoolFlag{Name: "stdin"},
					&cli.StringSliceFlag{Name: "image", Aliases: []string{"i"}},
				},
				HelpName:  "post",
				ArgsUsage: "[text]",
				Action:    doPost,
			},
			{
				Name:        "vote",
				Description: "vote the post",
				Usage:       "vote the post",
				UsageText:   "bsky vote [uri]",
				HelpName:    "vote",
				Action:      doVote,
			},
			{
				Name:        "votes",
				Description: "show votes of the post",
				Usage:       "show votes of the post",
				UsageText:   "bsky votes [uri]",
				HelpName:    "votes",
				Action:      doVotes,
				ArgsUsage:   "[uri]",
			},
			{
				Name:        "repost",
				Description: "repost the post",
				Usage:       "repost the post",
				UsageText:   "bsky repost [uri]",
				HelpName:    "repost",
				Action:      doRepost,
			},
			{
				Name:        "reposts",
				Description: "show reposts of the post",
				Usage:       "show reposts of the post",
				UsageText:   "bsky reposts [uri]",
				HelpName:    "reposts",
				Action:      doReposts,
			},
			{
				Name:        "follow",
				Description: "follow the handle",
				Usage:       "follow the handle",
				UsageText:   "bsky follow [handle]",
				HelpName:    "follow",
				Action:      doFollow,
			},
			{
				Name:        "follows",
				Description: "show follows",
				Usage:       "show follows",
				UsageText:   "bsky follows",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "follows",
				Action:   doFollows,
			},
			{
				Name:        "followers",
				Description: "show followers",
				Usage:       "show followers",
				UsageText:   "bsky followres",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "followers",
				Action:   doFollowers,
			},
			{
				Name:        "delete",
				Description: "delete the note",
				Usage:       "delete the note",
				UsageText:   "bsky delete [cid]",
				HelpName:    "delete",
				Action:      doDelete,
			},
			{
				Name:      "login",
				Usage:     "login the social",
				UsageText: "bsky login [handle] [password]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "host", Value: "https://bsky.social"},
				},
				HelpName: "login",
				Action:   doLogin,
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
