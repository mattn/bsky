package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/xrpc"
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
	fmt.Printf("Follows: %d\n", int64p(profile.FollowsCount))
	fmt.Printf("Followers: %d\n", int64p(profile.FollowersCount))
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
	// read options
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

	currentProfile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, xrpcc.Auth.Handle)
	if err != nil {
		return fmt.Errorf("cannot get profile: %w", err)
	}
	if name == nil {
		name = currentProfile.DisplayName
	}
	if desc == nil {
		desc = currentProfile.Description
	}

	var avatar *lexutil.LexBlob
	if currentProfile.Avatar != nil {
		currentAvatarCid, currentAvatarType, err := ParseCid(currentProfile.Avatar)
		if err != nil {
			return fmt.Errorf("cannot parse avatar cid: %w", err)
		}
		avatar = &lexutil.LexBlob{
			Ref:      lexutil.LexLink(currentAvatarCid),
			MimeType: currentAvatarType,
		}
	}
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
	if currentProfile.Banner != nil {
		currentBannerCid, currentBannerType, err := ParseCid(currentProfile.Banner)
		if err != nil {
			return fmt.Errorf("cannot parse banner cid: %w", err)
		}
		banner = &lexutil.LexBlob{
			Ref:      lexutil.LexLink(currentBannerCid),
			MimeType: currentBannerType,
		}
	}

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
	currentRecord, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", "app.bsky.actor.profile", xrpcc.Auth.Did, "self")
	if err != nil {
		return fmt.Errorf("cannot get profile: %w", err)
	}

	updatedRecord := &lexutil.LexiconTypeDecoder{Val: &bsky.ActorProfile{
		Description: desc,
		DisplayName: name,
		Avatar:      avatar,
		Banner:      banner,
	}}

	_, err = comatproto.RepoPutRecord(context.TODO(), xrpcc, &comatproto.RepoPutRecord_Input{
		Repo:       xrpcc.Auth.Did,
		Collection: "app.bsky.actor.profile",
		Rkey:       "self",
		Record:     updatedRecord,
		SwapRecord: currentRecord.Cid,
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
		profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, arg)
		if err != nil {
			return fmt.Errorf("cannot get profile: %w", err)
		}

		follow := bsky.GraphFollow{
			LexiconTypeID: "app.bsky.graph.follow",
			CreatedAt:     time.Now().Local().Format(time.RFC3339),
			Subject:       profile.Did,
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

func doUnfollow(cCtx *cli.Context) error {
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

		if profile.Viewer.Following == nil {
			continue
		}

		parts := strings.Split(*profile.Viewer.Following, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid post uri: %q", arg)
		}
		rkey := parts[len(parts)-1]
		schema := parts[len(parts)-2]
		fmt.Println(stringp(profile.Viewer.Following))
		_, err = comatproto.RepoDeleteRecord(context.TODO(), xrpcc, &comatproto.RepoDeleteRecord_Input{
			Repo:       xrpcc.Auth.Did,
			Collection: schema,
			Rkey:       rkey,
		})
		if err != nil {
			return err
		}
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

func doBlock(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
	  var did *string
    if strings.HasPrefix(arg, "did:") {
      did = &arg
    } else {
      profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, arg)
      if err != nil {
        profileErr := err
        result, err := bsky.ActorSearchActors(context.TODO(), xrpcc, "", 50, arg, "")
        if err != nil {
          panic("Failed to search: " + err.Error())
        }
        for _, actor := range result.Actors {
          if err == nil && arg == actor.Handle {
            did = &actor.Did
          }
        }
        if did == nil {
          panic("Failed to get profile: " + profileErr.Error())
        }
      } else {
        did = &profile.Did
      }
    }

		block := bsky.GraphBlock{
			LexiconTypeID: "app.bsky.graph.block",
			CreatedAt:     time.Now().Local().Format(time.RFC3339),
			Subject:       *did,
		}

		resp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.graph.block",
			Repo:       xrpcc.Auth.Did,
			Record: &lexutil.LexiconTypeDecoder{
				Val: &block,
			},
		})
		if err != nil {
			return err
		}
		fmt.Println(resp.Uri)
	}
	return nil
}

func doMute(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
	  var did *string
    if strings.HasPrefix(arg, "did:") {
      did = &arg
    } else {
      profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, arg)
      if err != nil {
        profileErr := err
        result, err := bsky.ActorSearchActors(context.TODO(), xrpcc, "", 50, arg, "")
        if err != nil {
          panic("Failed to search: " + err.Error())
        }
        for _, actor := range result.Actors {
          if err == nil && arg == actor.Handle {
            did = &actor.Did
          }
        }
        if did == nil {
          panic("Failed to get profile: " + profileErr.Error())
        }
      } else {
        did = &profile.Did
      }
    }

		err = bsky.GraphMuteActor(context.TODO(), xrpcc, &bsky.GraphMuteActor_Input{Actor: *did})
		if err != nil {
			panic("Failed to mute user: " + err.Error())
		}
	}
	return nil
}

func doReport(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
	  var did *string
    if strings.HasPrefix(arg, "did:") {
      did = &arg
    } else {
      profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, arg)
      if err != nil {
        profileErr := err
        result, err := bsky.ActorSearchActors(context.TODO(), xrpcc, "", 50, arg, "")
        if err != nil {
          panic("Failed to search: " + err.Error())
        }
        for _, actor := range result.Actors {
          if err == nil && arg == actor.Handle {
            did = &actor.Did
          }
        }
        if did == nil {
          panic("Failed to get profile: " + profileErr.Error())
        }
      } else {
        did = &profile.Did
      }
    }

    comment := cCtx.String("comment")
		var reasonType string
		reasonType = "com.atproto.moderation.defs#reasonSpam"
		input := map[string]interface{}{
			"reasonType": reasonType,
			"subject": map[string]string{
				"$type": "com.atproto.admin.defs#repoRef",
				"did": *did,
			},
			"comment":   comment,
			"createdAt": time.Now().Format(time.RFC3339),
		}

		var response map[string]interface{}
		err = xrpcc.Do(context.TODO(), xrpc.Procedure, "application/json", "com.atproto.moderation.createReport", nil, input, &response)
		if err != nil {
			panic("Failed to create report: " + err.Error())
		}

		fmt.Println("Report created successfully:", response)
	}
	return nil
}

func doModList(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	name := cCtx.String("name")
	description := cCtx.String("description")

  var purpose string
  purpose =     "app.bsky.graph.defs#modlist"
  modList := bsky.GraphList{
      Name:        name,
      Purpose:     &purpose,
      Description: &description,
      CreatedAt:   time.Now().Format(time.RFC3339),
  }

  listResp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
      Repo:       xrpcc.Auth.Did,
      Collection: "app.bsky.graph.list",
      Record:     &lexutil.LexiconTypeDecoder{
        Val: &modList,
      },
  })
  if err != nil {
      panic(err)
  }

  listURI := listResp.Uri
  fmt.Println("List created successfully. URI:", listURI)

	for _, arg := range cCtx.Args().Slice() {
	  var did *string
    if strings.HasPrefix(arg, "did:") {
      did = &arg
    } else {
      profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, arg)
      if err != nil {
        profileErr := err
        result, err := bsky.ActorSearchActors(context.TODO(), xrpcc, "", 50, arg, "")
        if err != nil {
          panic("Failed to search: " + err.Error())
        }
        for _, actor := range result.Actors {
          if err == nil && arg == actor.Handle {
            did = &actor.Did
          }
        }
        if did == nil {
          panic("Failed to get profile: " + profileErr.Error())
        }
      } else {
        did = &profile.Did
      }
    }

		listItem := bsky.GraphListitem{
				Subject: *did,
				List:    listURI,
				CreatedAt: time.Now().Format(time.RFC3339),
		}

		_, err = comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
				Repo:       xrpcc.Auth.Did,
				Collection: "app.bsky.graph.listitem",
				Record:     &lexutil.LexiconTypeDecoder{
				  Val: &listItem,
				},
		})
		if err != nil {
				panic(err)
		}

		fmt.Println("User added to moderation list successfully.")
	}
	return nil
}

func doSearchActors(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	n := cCtx.Int64("n")

	for _, arg := range cCtx.Args().Slice() {
    result, err := bsky.ActorSearchActors(context.TODO(), xrpcc, "", n, arg, "")
    if err != nil {
      panic("Failed to search: " + err.Error())
    }
    for _, actor := range result.Actors {
      jsn, err := json.MarshalIndent(&actor, "", "  ")
      if err == nil {
        fmt.Println("Actor: ", string(jsn))
      }
    }
	}
	return nil
}

func doUnblock(cCtx *cli.Context) error {
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

		if profile.Viewer.Blocking == nil {
			continue
		}

		parts := strings.Split(*profile.Viewer.Blocking, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid post uri: %q", arg)
		}
		rkey := parts[len(parts)-1]
		schema := parts[len(parts)-2]
		fmt.Println(stringp(profile.Viewer.Blocking))
		_, err = comatproto.RepoDeleteRecord(context.TODO(), xrpcc, &comatproto.RepoDeleteRecord_Input{
			Repo:       xrpcc.Auth.Did,
			Collection: schema,
			Rkey:       rkey,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func doBlocks(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	var cursor string
	for {
		blocks, err := bsky.GraphGetBlocks(context.TODO(), xrpcc, cursor, 100)
		if err != nil {
			return fmt.Errorf("getting record: %w", err)
		}

		if cCtx.Bool("json") {
			for _, f := range blocks.Blocks {
				json.NewEncoder(os.Stdout).Encode(f)
			}
		} else {
			for _, f := range blocks.Blocks {
				color.Set(color.FgHiRed)
				fmt.Print(f.Handle)
				color.Set(color.Reset)
				fmt.Printf(" [%s] ", stringp(f.DisplayName))
				color.Set(color.FgBlue)
				fmt.Println(f.Did)
				color.Set(color.Reset)
			}
		}
		if blocks.Cursor == nil {
			break
		}
		cursor = *blocks.Cursor
	}
	return nil
}

func doLogin(cCtx *cli.Context) error {
	fp, _ := cCtx.App.Metadata["path"].(string)
	var cfg config
	cfg.Host = cCtx.String("host")
	cfg.Bgs = cCtx.String("bgs")
	cfg.Handle = cCtx.Args().Get(0)
	cfg.Password = cCtx.Args().Get(1)
	if cfg.Handle == "" || cfg.Password == "" {
		cli.ShowSubcommandHelpAndExit(cCtx, 1)
	}
	b, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot make config file: %w", err)
	}
	err = os.WriteFile(fp, b, 0644)
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

	notifs, err := bsky.NotificationListNotifications(context.TODO(), xrpcc, "", 50, false, "")
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

func doShowSession(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	session, err := comatproto.ServerGetSession(context.TODO(), xrpcc)
	if err != nil {
		return err
	}

	if cCtx.Bool("json") {
		json.NewEncoder(os.Stdout).Encode(session)
		return nil
	}

	fmt.Printf("Did: %s\n", session.Did)
	fmt.Printf("Email: %s\n", stringp(session.Email))
	fmt.Printf("Handle: %s\n", session.Handle)
	return nil
}

func doInviteCodes(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	includeUsed := cCtx.Bool("used")

	codes, err := comatproto.ServerGetAccountInviteCodes(context.TODO(), xrpcc, false, includeUsed)
	if err != nil {
		return err
	}

	if cCtx.Bool("json") {
		for _, c := range codes.Codes {
			json.NewEncoder(os.Stdout).Encode(c)
		}
		return nil
	}

	for _, c := range codes.Codes {
		if int64(len(c.Uses)) >= c.Available { // used
			color.Set(color.FgHiMagenta)
			fmt.Printf("%s (used)\n", c.Code)
			color.Set(color.Reset)
		} else {
			fmt.Println(c.Code)
		}
	}

	return nil
}
