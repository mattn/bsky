package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/fatih/color"

	"github.com/urfave/cli/v2"
)

func doShowProfile(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.String("handle")
	if arg == "" {
		arg = xrpcc.Auth.Handle
	}

	profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, arg)
	if err != nil {
		return fmt.Errorf("cannot get profile: %w", err)
	}

	if cCtx.Bool("json") {
		json.NewEncoder(os.Stdout).Encode(profile)
		return nil
	}

	fmt.Printf("Did: %s\n", profile.Did)
	fmt.Printf("Handle: %s\n", profile.Handle)
	fmt.Printf("DisplayName: %s\n", stringp(profile.DisplayName))
	fmt.Printf("Description: %s\n", stringp(profile.Description))
	fmt.Printf("Follows: %d\n", profile.FollowsCount)
	fmt.Printf("Followers: %d\n", profile.FollowersCount)
	fmt.Printf("Avatar: %s\n", stringp(profile.Avatar))
	fmt.Printf("Banner: %s\n", stringp(profile.Banner))
	return nil
}

func doUpdateProfile(cCtx *cli.Context) error {
	// read arguments
	var name *string
	if cCtx.Args().Len() >= 1 {
		v := cCtx.Args().Get(0)
		name = &v
	}
	var desc *string
	if cCtx.Args().Len() >= 2 {
		v := cCtx.Args().Get(1)
		desc = &v
	}
	// read optionns
	var avatarFn *string
	if s := cCtx.String("avatar"); s != "" {
		avatarFn = &s
	}
	var bannerFn *string
	if s := cCtx.String("banner"); s != "" {
		bannerFn = &s
	}

	if name == nil && desc == nil && avatarFn == nil && bannerFn == nil {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	var avatar *lexutil.LexBlob
	if avatarFn != nil {
		b, err := os.ReadFile(*avatarFn)
		if err != nil {
			return fmt.Errorf("cannot read image file: %w", err)
		}

		resp, err := comatproto.RepoUploadBlob(context.TODO(), xrpcc, bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("cannot upload image file: %w", err)
		}
		avatar = &lexutil.LexBlob{
			Ref:      resp.Blob.Ref,
			MimeType: http.DetectContentType(b),
			Size:     resp.Blob.Size,
		}
	}
	var banner *lexutil.LexBlob
	if bannerFn != nil {
		b, err := os.ReadFile(*bannerFn)
		if err != nil {
			return fmt.Errorf("cannot read image file: %w", err)
		}
		resp, err := comatproto.RepoUploadBlob(context.TODO(), xrpcc, bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("cannot upload image file: %w", err)
		}
		banner = &lexutil.LexBlob{
			Ref:      resp.Blob.Ref,
			MimeType: http.DetectContentType(b),
			Size:     resp.Blob.Size,
		}
	}

	_, err = comatproto.RepoPutRecord(context.TODO(), xrpcc, &comatproto.RepoPutRecord_Input{
		Repo:       xrpcc.Auth.Did,
		Collection: "app.bsky.actor.profile",
		Rkey:       "self",
		Record: &lexutil.LexiconTypeDecoder{&bsky.ActorProfile{
			Description: desc,
			DisplayName: name,
			Avatar:      avatar,
			Banner:      banner,
		}},
	})

	if err != nil {
		return fmt.Errorf("cannot update profile: %w", err)
	}
	return nil
}

func doFollow(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
		follow := bsky.GraphFollow{
			LexiconTypeID: "app.bsky.graph.follow",
			CreatedAt:     time.Now().Local().Format(time.RFC3339),
			Subject:       arg,
		}

		resp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.graph.follow",
			Repo:       xrpcc.Auth.Did,
			Record: &lexutil.LexiconTypeDecoder{
				Val: &follow,
			},
		})
		if err != nil {
			return err
		}
		fmt.Println(resp.Uri)
	}
	return nil
}

func doFollows(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.String("handle")
	if arg == "" {
		arg = xrpcc.Auth.Handle
	}

	var cursor string
	for {
		follows, err := bsky.GraphGetFollows(context.TODO(), xrpcc, arg, cursor, 100)
		if err != nil {
			return fmt.Errorf("getting record: %w", err)
		}

		if cCtx.Bool("json") {
			for _, f := range follows.Follows {
				json.NewEncoder(os.Stdout).Encode(f)
			}
		} else {
			for _, f := range follows.Follows {
				color.Set(color.FgHiRed)
				fmt.Print(f.Handle)
				color.Set(color.Reset)
				fmt.Printf(" [%s] ", stringp(f.DisplayName))
				color.Set(color.FgBlue)
				fmt.Println(f.Did)
				color.Set(color.Reset)
			}
		}
		if follows.Cursor == nil {
			break
		}
		cursor = *follows.Cursor
	}
	return nil
}

func doFollowers(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.String("handle")
	if arg == "" {
		arg = xrpcc.Auth.Handle
	}

	var cursor string
	for {
		followers, err := bsky.GraphGetFollowers(context.TODO(), xrpcc, arg, cursor, 100)
		if err != nil {
			return fmt.Errorf("getting record: %w", err)
		}

		if cCtx.Bool("json") {
			for _, f := range followers.Followers {
				json.NewEncoder(os.Stdout).Encode(f)
			}
		} else {
			for _, f := range followers.Followers {
				color.Set(color.FgHiRed)
				fmt.Print(f.Handle)
				color.Set(color.Reset)
				fmt.Printf(" [%s] ", stringp(f.DisplayName))
				color.Set(color.FgBlue)
				fmt.Println(f.Did)
				color.Set(color.Reset)
			}
		}
		if followers.Cursor == nil {
			break
		}
		cursor = *followers.Cursor
	}
	return nil
}

func doLogin(cCtx *cli.Context) error {
	fp, _ := cCtx.App.Metadata["path"].(string)
	var cfg config
	cfg.Host = cCtx.String("host")
	cfg.Handle = cCtx.Args().Get(0)
	cfg.Password = cCtx.Args().Get(1)
	if cfg.Handle == "" || cfg.Password == "" {
		cli.ShowSubcommandHelpAndExit(cCtx, 1)
	}
	b, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot make config file: %w", err)
	}
	err = ioutil.WriteFile(fp, b, 0644)
	if err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}
	return nil
}

func doNotification(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	notifs, err := bsky.NotificationListNotifications(context.TODO(), xrpcc, "", 50)
	if err != nil {
		return err
	}

	if cCtx.Bool("json") {
		for _, n := range notifs.Notifications {
			json.NewEncoder(os.Stdout).Encode(n)
		}
		return nil
	}

	for _, n := range notifs.Notifications {
		if !cCtx.Bool("a") && n.IsRead {
			continue
		}
		color.Set(color.FgHiRed)
		fmt.Print(n.Author.Handle)
		color.Set(color.Reset)
		fmt.Printf(" [%s] ", stringp(n.Author.DisplayName))
		color.Set(color.FgBlue)
		fmt.Println(n.Author.Did)
		color.Set(color.Reset)

		switch v := n.Record.Val.(type) {
		case *bsky.FeedPost:
			fmt.Println(" " + n.Reason + " to " + n.Uri)
		case *bsky.FeedRepost:
			fmt.Printf(" reposted %s\n", v.Subject.Uri)
		case *bsky.FeedLike:
			fmt.Printf(" liked %s\n", v.Subject.Uri)
		case *bsky.GraphFollow:
			fmt.Println(" followed you")
		}

		bsky.NotificationUpdateSeen(context.TODO(), xrpcc, &bsky.NotificationUpdateSeen_Input{
			SeenAt: time.Now().Local().Format(time.RFC3339),
		})
	}

	return nil
}
