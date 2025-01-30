package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const name = "bsky"

const version = "0.0.69"

//nolint:unused
var revision = "HEAD"

type config struct {
	Bgs      string `json:"bgs"`
	Host     string `json:"host"`
	Handle   string `json:"handle"`
	Password string `json:"password"`
	dir      string
	verbose  bool
	prefix   string
}

func main() {
	app := &cli.App{
		Name:                 name,
		Usage:                name,
		Version:              version,
		Description:          "A cli application for bluesky",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "a", Usage: "profile name"},
			&cli.BoolFlag{Name: "V", Usage: "verbose"},
		},
		DisableSliceFlagSeparator: true,
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
					&cli.StringFlag{Name: "cursor", Value: "", Usage: "cursor"},
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
					&cli.StringFlag{Name: "pattern", Usage: "pattern"},
					&cli.StringFlag{Name: "reply", Usage: "reply"},
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
					&cli.StringSliceFlag{Name: "image-alt", Aliases: []string{"ia"}},
					&cli.StringFlag{Name: "video", Aliases: []string{"v"}},
					&cli.StringFlag{Name: "video-alt", Aliases: []string{"va"}},
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
				Name:        "unfollow",
				Description: "Unfollow the handle",
				Usage:       "Unfollow the handle",
				UsageText:   "bsky unfollow [handle]",
				HelpName:    "unfollow",
				Action:      doUnfollow,
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
				Name:        "block",
				Description: "Block the handle",
				Usage:       "Block the handle",
				UsageText:   "bsky block [handle/did]",
				HelpName:    "block",
				Action:      doBlock,
			},
			{
				Name:        "unblock",
				Description: "Unblock the handle",
				Usage:       "Unblock the handle",
				UsageText:   "bsky unblock [handle]",
				HelpName:    "unblock",
				Action:      doUnblock,
			},
			{
				Name:        "blocks",
				Description: "Show blocks",
				Usage:       "Show blocks",
				UsageText:   "bsky blocks",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "blocks",
				Action:   doBlocks,
			},
			{
				Name:        "mute",
				Description: "Mute the handle",
				Usage:       "Mute the handle",
				UsageText:   "bsky mute [handle/did]",
				HelpName:    "mute",
				Action:      doMute,
			},
			{
				Name:        "report",
				Description: "Report the handle",
				Usage:       "Report the handle",
				UsageText:   "bsky report [handle/did]",
				HelpName:    "report",
				Action:      doReport,
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "comment", Usage: "report comment"},
				},
			},
			{
				Name:        "moderation-list",
				Description: "Add the handle to a new moderation list",
				Usage:       "Add the handle to a new moderation list",
				UsageText:   "bsky moderation-list [handle/did]",
				HelpName:    "moderation-list",
				Action:      doModList,
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Value: "NewList", Usage: "list name"},
					&cli.StringFlag{Name: "description", Aliases: []string{"desc"}, Value: "", Usage: "description"},
				},
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
				Name:        "search",
				Description: "Search Bluesky",
				Usage:       "Search Bluesky",
				UsageText:   "bsky search [terms]",
				HelpName:    "search",
				Action:      doSearch,
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "n", Value: 100, Usage: "number of items"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
			},
			{
				Name:        "search-actors",
				Description: "Search Actors",
				Usage:       "Search Actors",
				UsageText:   "bsky search-actors [terms]",
				HelpName:    "search-actors",
				Action:      doSearchActors,
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "n", Value: 100, Usage: "number of items"},
				},
			},
			{
				Name:        "login",
				Description: "Login the social",
				Usage:       "Login the social",
				UsageText:   "bsky login [handle] [password]",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "host", Value: "https://bsky.social"},
					&cli.StringFlag{Name: "bgs", Value: "https://bsky.network"},
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
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "notification",
				Action:   doNotification,
			},
			{
				Name:        "invite-codes",
				Description: "Show invite codes",
				Usage:       "Show invite codes",
				UsageText:   "bsky invite-codes",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "used", Usage: "show used codes too"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "invite-codes",
				Action:   doInviteCodes,
			},
			{
				Name:        "list-app-passwords",
				Description: "Show App-passwords",
				Usage:       "Show App-passwords",
				UsageText:   "bsky list-app-passwords",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "list-app-passwords",
				Action:   doListAppPasswords,
			},
			{
				Name:        "add-app-password",
				Description: "Add App-password",
				Usage:       "Add App-password",
				UsageText:   "bsky add-app-password",
				HelpName:    "add-app-password",
				Action:      doAddAppPassword,
			},
			{
				Name:        "revoke-app-password",
				Description: "Revoke App-password",
				Usage:       "Revoke App-password",
				UsageText:   "bsky revoke-app-password",
				HelpName:    "revoke-app-password",
				Action:      doRevokeAppPassword,
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
			if profile != "" {
				cfg.prefix = profile + "-"
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
