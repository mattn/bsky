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
	if cCtx.Args().Len() != 2 {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}
	name := cCtx.Args().Get(0)
	desc := cCtx.Args().Get(1)

	var avatar *lexutil.Blob
	avatarFn := cCtx.String("avatar")
	if avatarFn != "" {
		b, err := os.ReadFile(avatarFn)
		if err != nil {
			return fmt.Errorf("cannot read image file: %w", err)
		}
		resp, err := comatproto.BlobUpload(context.TODO(), xrpcc, bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("cannot upload image file: %w", err)
		}
		avatar = &lexutil.Blob{
			Cid:      resp.Cid,
			MimeType: http.DetectContentType(b),
		}
	}
	var banner *lexutil.Blob
	bannerFn := cCtx.String("banner")
	if bannerFn != "" {
		b, err := os.ReadFile(bannerFn)
		if err != nil {
			return fmt.Errorf("cannot read image file: %w", err)
		}
		resp, err := comatproto.BlobUpload(context.TODO(), xrpcc, bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("cannot upload image file: %w", err)
		}
		banner = &lexutil.Blob{
			Cid:      resp.Cid,
			MimeType: http.DetectContentType(b),
		}
	}

	_, err = bsky.ActorUpdateProfile(context.TODO(), xrpcc, &bsky.ActorUpdateProfile_Input{
		Description: &desc,
		DisplayName: &name,
		Avatar:      avatar,
		Banner:      banner,
	})
	return fmt.Errorf("cannot update profile: %w", err)
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
		profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, arg)
		if err != nil {
			return fmt.Errorf("cannot get profile: %w", err)
		}

		follow := bsky.GraphFollow{
			LexiconTypeID: "app.bsky.graph.follow",
			CreatedAt:     time.Now().Format(time.RFC3339),
			Subject: &bsky.ActorRef{
				DeclarationCid: profile.Declaration.Cid,
				Did:            profile.Did,
			},
		}

		resp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.graph.follow",
			Did:        xrpcc.Auth.Did,
			Record: lexutil.LexiconTypeDecoder{
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
		follows, err := bsky.GraphGetFollows(context.TODO(), xrpcc, cursor, 100, arg)
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
		followers, err := bsky.GraphGetFollowers(context.TODO(), xrpcc, cursor, 100, arg)
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
	var cfg Config
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
