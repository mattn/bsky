package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	cliutil "github.com/bluesky-social/indigo/cmd/gosky/util"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const name = "bsky"

const version = "0.0.2"

var revision = "HEAD"

type Config struct {
	Host     string `json:"host"`
	Handle   string `json:"handle"`
	Password string `json:"password"`
	dir      string
	verbose  bool
}

func loadConfig(profile string) (*Config, string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, "", err
	}
	dir = filepath.Join(dir, "bsky")

	var fp string
	if profile == "" {
		fp = filepath.Join(dir, "config.json")
	} else if profile == "?" {
		names, err := filepath.Glob(filepath.Join(dir, "config-*.json"))
		if err != nil {
			return nil, "", err
		}
		for _, name := range names {
			name = filepath.Base(name)
			name = strings.TrimLeft(name[6:len(name)-5], "-")
			fmt.Println(name)
		}
		os.Exit(0)
	} else {
		fp = filepath.Join(dir, "config-"+profile+".json")
	}
	os.MkdirAll(filepath.Dir(fp), 0700)

	b, err := os.ReadFile(fp)
	if err != nil {
		return nil, fp, fmt.Errorf("cannot load config file: %w", err)
	}
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, fp, fmt.Errorf("cannot load config file: %w", err)
	}
	if cfg.Host == "" {
		cfg.Host = "https://bsky.social"
	}
	cfg.dir = dir
	return &cfg, fp, nil
}

func makeXRPCC(cCtx *cli.Context) (*xrpc.Client, error) {
	cfg := cCtx.App.Metadata["config"].(*Config)

	xrpcc := &xrpc.Client{
		Client: cliutil.NewHttpClient(),
		Host:   cfg.Host,
		Auth:   &xrpc.AuthInfo{Handle: cfg.Handle},
	}

	auth, err := cliutil.ReadAuth(filepath.Join(cfg.dir, cfg.Handle+".auth"))
	if err == nil {
		xrpcc.Auth = auth
		refresh, err2 := comatproto.SessionRefresh(context.TODO(), xrpcc)
		if err2 != nil {
			err = err2
		} else {
			xrpcc.Auth.Did = refresh.Did
			xrpcc.Auth.AccessJwt = refresh.AccessJwt
			xrpcc.Auth.RefreshJwt = refresh.RefreshJwt

			b, err := json.Marshal(xrpcc.Auth)
			if err == nil {
				if err := os.WriteFile(filepath.Join(cfg.dir, cfg.Handle+".auth"), b, 0600); err != nil {
					return nil, fmt.Errorf("cannot write auth file: %w", err)
				}
			}
		}
	}
	if err != nil {
		auth, err := comatproto.SessionCreate(context.TODO(), xrpcc, &comatproto.SessionCreate_Input{
			Identifier: &xrpcc.Auth.Handle,
			Password:   cfg.Password,
		})
		if err != nil {
			return nil, fmt.Errorf("cannot create session: %w", err)
		}
		xrpcc.Auth.Did = auth.Did
		xrpcc.Auth.AccessJwt = auth.AccessJwt
		xrpcc.Auth.RefreshJwt = auth.RefreshJwt

		b, err := json.Marshal(xrpcc.Auth)
		if err == nil {
			if err := os.WriteFile(filepath.Join(cfg.dir, cfg.Handle+".auth"), b, 0600); err != nil {
				return nil, fmt.Errorf("cannot write auth file: %w", err)
			}
		}
	}

	return xrpcc, nil
}

func doLogin(cCtx *cli.Context) error {
	fp, _ := cCtx.App.Metadata["path"].(string)
	var cfg Config
	cfg.Host = cCtx.String("host")
	cfg.Handle = cCtx.Args().Get(0)
	cfg.Password = cCtx.Args().Get(1)
	if cfg.Handle == "" || cfg.Password == "" {
		return errors.New("handle and pasword are required")
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

func doPost(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	var images []*bsky.EmbedImages_Image
	imageFn := cCtx.StringSlice("image")
	if len(imageFn) > 0 {
		for _, fn := range imageFn {
			b, err := os.ReadFile(fn)
			if err != nil {
				return fmt.Errorf("cannot read image file: %w", err)
			}
			resp, err := comatproto.BlobUpload(context.TODO(), xrpcc, bytes.NewReader(b))
			if err != nil {
				return fmt.Errorf("cannot upload image file: %w", err)
			}
			images = append(images, &bsky.EmbedImages_Image{
				Alt: filepath.Base(fn),
				Image: &lexutil.Blob{
					Cid:      resp.Cid,
					MimeType: http.DetectContentType(b),
				},
			})
		}
	}

	var reply *bsky.FeedPost_ReplyRef
	replyTo := cCtx.String("r")
	if replyTo != "" {
		parts := strings.Split(replyTo, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid post uri: %q", replyTo)
		}
		rkey := parts[len(parts)-1]
		collection := parts[len(parts)-2]
		did := parts[2]

		resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, rkey, did)
		if err != nil {
			return fmt.Errorf("cannot get record: %w", err)
		}
		reply = &bsky.FeedPost_ReplyRef{
			LexiconTypeID: resp.LexiconTypeID,
			Parent:        &comatproto.RepoStrongRef{Cid: *resp.Cid, Uri: resp.Uri},
		}
		post := resp.Value.Val.(*bsky.FeedPost)
		if post.Reply != nil && post.Reply.Root != nil {
			reply.Root = &comatproto.RepoStrongRef{Cid: post.Reply.Root.Cid, Uri: post.Reply.Root.Uri}
		} else {
			reply.Root = reply.Parent
		}
	}

	text := strings.Join(cCtx.Args().Slice(), " ")
	post := &bsky.FeedPost{
		Text:      text,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05.000Z"),
		Reply:     reply,
	}
	if len(images) > 0 {
		post.Embed = &bsky.FeedPost_Embed{
			EmbedImages: &bsky.EmbedImages{
				Images: images,
			},
		}
	}
	resp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Did:        xrpcc.Auth.Did,
		Record: lexutil.LexiconTypeDecoder{
			Val: post,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	fmt.Println(resp.Uri)

	return nil
}

func doVote(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()

	parts := strings.Split(arg, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid post uri: %q", arg)
	}
	rkey := parts[len(parts)-1]
	collection := parts[len(parts)-2]
	did := parts[2]

	dir := cCtx.Args().Get(1)
	if dir == "" {
		dir = "up"
	}

	resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, rkey, did)
	if err != nil {
		return fmt.Errorf("getting record: %w", err)
	}

	_, err = bsky.FeedSetVote(context.TODO(), xrpcc, &bsky.FeedSetVote_Input{
		Subject:   &comatproto.RepoStrongRef{Uri: resp.Uri, Cid: *resp.Cid},
		Direction: dir,
	})
	if err != nil {
		return fmt.Errorf("cannot create vote: %w", err)
	}

	return nil
}

func doVotes(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()

	parts := strings.Split(arg, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid post uri: %q", arg)
	}
	rkey := parts[len(parts)-1]
	collection := parts[len(parts)-2]
	did := parts[2]

	resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, rkey, did)
	if err != nil {
		return fmt.Errorf("getting record: %w", err)
	}

	votes, err := bsky.FeedGetVotes(context.TODO(), xrpcc, "", *resp.Cid, "", 50, resp.Uri)
	if err != nil {
		return fmt.Errorf("getting votes: %w", err)
	}

	if cCtx.Bool("json") {
		for _, v := range votes.Votes {
			json.NewEncoder(os.Stdout).Encode(v)
		}
		return nil
	}

	for _, v := range votes.Votes {
		if v.Direction == "up" {
			fmt.Print("👍 ")
		} else {
			fmt.Print("👎 ")
		}
		color.Set(color.FgHiRed)
		fmt.Print(v.Actor.Handle)
		color.Set(color.Reset)
		fmt.Printf(" [%s]", stringp(v.Actor.DisplayName))
		fmt.Printf(" (%v)\n", ltime(v.CreatedAt))
	}

	return nil
}

func doRepost(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()

	parts := strings.Split(arg, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid post uri: %q", arg)
	}
	rkey := parts[len(parts)-1]
	collection := parts[len(parts)-2]
	did := parts[2]

	resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, rkey, did)
	if err != nil {
		return fmt.Errorf("getting record: %w", err)
	}

	repost := &bsky.FeedRepost{
		CreatedAt: time.Now().Format(time.RFC3339),
		Subject: &comatproto.RepoStrongRef{
			Uri: resp.Uri,
			Cid: *resp.Cid,
		},
	}
	_, err = comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.repost",
		Did:        xrpcc.Auth.Did,
		Record: lexutil.LexiconTypeDecoder{
			Val: repost,
		},
	})
	if err != nil {
		return fmt.Errorf("cannot create repost: %w", err)
	}

	return nil
}

func doReposts(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()

	parts := strings.Split(arg, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid post uri: %q", arg)
	}
	rkey := parts[len(parts)-1]
	collection := parts[len(parts)-2]
	did := parts[2]

	resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, rkey, did)
	if err != nil {
		return fmt.Errorf("getting record: %w", err)
	}

	reposts, err := bsky.FeedGetRepostedBy(context.TODO(), xrpcc, "", *resp.Cid, 50, resp.Uri)
	if err != nil {
		return fmt.Errorf("getting reposts: %w", err)
	}

	if cCtx.Bool("json") {
		for _, r := range reposts.RepostedBy {
			json.NewEncoder(os.Stdout).Encode(r)
		}
		return nil
	}

	for _, r := range reposts.RepostedBy {
		fmt.Printf("⚡ ")
		color.Set(color.FgHiRed)
		fmt.Print(r.Handle)
		color.Set(color.Reset)
		fmt.Printf(" [%s]\n", stringp(r.DisplayName))
	}

	return nil
}

func doFollow(cCtx *cli.Context) error {
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

func doDelete(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()

	parts := strings.Split(arg, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid post uri: %q", arg)
	}
	rkey := parts[len(parts)-1]
	schema := parts[len(parts)-2]

	return comatproto.RepoDeleteRecord(context.TODO(), xrpcc, &comatproto.RepoDeleteRecord_Input{
		Did:        xrpcc.Auth.Did,
		Collection: schema,
		Rkey:       rkey,
	})
}

func doTimeline(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	var feed []*bsky.FeedFeedViewPost

	n := cCtx.Int64("n")
	handle := cCtx.String("handle")
	if handle != "" {
		if handle == "self" {
			handle = xrpcc.Auth.Did
		}
		resp, err := bsky.FeedGetAuthorFeed(context.TODO(), xrpcc, handle, "", n)
		if err != nil {
			return fmt.Errorf("cannot get author feed: %w", err)
		}
		feed = resp.Feed
	} else {
		handle = "reverse-chronological"
		resp, err := bsky.FeedGetTimeline(context.TODO(), xrpcc, handle, "", n)
		if err != nil {
			return fmt.Errorf("cannot get timeline: %w", err)
		}
		feed = resp.Feed
	}

	if cCtx.Bool("json") {
		for _, p := range feed {
			json.NewEncoder(os.Stdout).Encode(p)
		}
		return nil
	}

	for i := 0; i < len(feed)/2; i++ {
		feed[i], feed[len(feed)-i-1] = feed[len(feed)-i-1], feed[i]
	}
	for _, p := range feed {
		if p.Reason != nil {
			continue
		}
		printPost(p.Post)
		if p.Reply != nil {
			fmt.Print(" > ")
			color.Set(color.FgBlue)
			fmt.Println(p.Reply.Parent.Uri)
			color.Set(color.Reset)
		}
		fmt.Println()
	}

	return nil
}

func printPost(p *bsky.FeedPost_View) {
	rec := p.Record.Val.(*bsky.FeedPost)
	color.Set(color.FgHiRed)
	fmt.Print(p.Author.Handle)
	color.Set(color.Reset)
	fmt.Printf(" [%s]", stringp(p.Author.DisplayName))
	fmt.Printf(" (%v)\n", ltime(rec.CreatedAt))
	fmt.Println(rec.Text)
	if rec.Entities != nil {
		for _, e := range rec.Entities {
			fmt.Printf(" {%s}\n", e.Value)
		}
	}
	if rec.Embed != nil {
		if p.Embed.EmbedImages_Presented != nil {
			for _, i := range p.Embed.EmbedImages_Presented.Images {
				fmt.Println(" {" + i.Fullsize + "}")
			}
		}
	}
	fmt.Printf(" 👍(%d)👎(%d)⚡(%d)↩️ (%d)\n",
		p.UpvoteCount,
		p.DownvoteCount,
		p.RepostCount,
		p.ReplyCount,
	)
	fmt.Print(" - ")
	color.Set(color.FgBlue)
	fmt.Println(p.Uri)
	color.Set(color.Reset)
}

func doThread(cCtx *cli.Context) error {
	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	n := cCtx.Int64("n")
	resp, err := bsky.FeedGetPostThread(context.TODO(), xrpcc, n, cCtx.Args().First())
	if err != nil {
		return fmt.Errorf("cannot get post thread: %w", err)
	}

	replies := resp.Thread.FeedGetPostThread_ThreadViewPost.Replies
	if cCtx.Bool("json") {
		json.NewEncoder(os.Stdout).Encode(resp.Thread.FeedGetPostThread_ThreadViewPost.Post)
		for _, p := range replies {
			json.NewEncoder(os.Stdout).Encode(p)
		}
		return nil
	}

	for i := 0; i < len(replies)/2; i++ {
		replies[i], replies[len(replies)-i-1] = replies[len(replies)-i-1], replies[i]
	}
	printPost(resp.Thread.FeedGetPostThread_ThreadViewPost.Post)
	fmt.Println()
	for _, r := range replies {
		printPost(r.FeedGetPostThread_ThreadViewPost.Post)
		fmt.Println()
	}
	return nil
}

func ltime(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.000Z", s)
	if err != nil {
		return time.Now()
	}
	return t
}

func stringp(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

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

func main() {
	app := &cli.App{
		Description: "A cli application for bluesky",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "a", Usage: "profile name"},
			&cli.BoolFlag{Name: "V", Usage: "verbose"},
		},
		Commands: []*cli.Command{
			{
				Name:      "show-profile",
				Usage:     "show profile",
				UsageText: "bsky show-profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
				},
				Action: doShowProfile,
			},
			{
				Name:      "update-profile",
				Usage:     "update profile",
				UsageText: "bsky update-profile",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "avatar", Value: "", Usage: "avatar image", TakesFile: true},
					&cli.StringFlag{Name: "banner", Value: "", Usage: "banner image", TakesFile: true},
				},
				Action: doUpdateProfile,
			},
			{
				Name:      "timeline",
				Aliases:   []string{"tl"},
				Usage:     "show timeline",
				UsageText: "bsky timeline",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "handle", Aliases: []string{"H"}, Value: "", Usage: "user handle"},
					&cli.IntFlag{Name: "n", Value: 30, Usage: "number of items"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				Action: doTimeline,
			},
			{
				Name:      "thread",
				Usage:     "show thread",
				UsageText: "bsky thread [uri]",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "n", Value: 30, Usage: "number of items"},
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				Action: doThread,
			},
			{
				Name:      "post",
				Usage:     "post new text",
				UsageText: "bsky post [text]",
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
				Name:      "vote",
				Usage:     "vote the post",
				UsageText: "bsky vote [uri]",
				HelpName:  "vote",
				Action:    doVote,
			},
			{
				Name:      "votes",
				Usage:     "show votes of the post",
				UsageText: "bsky votes [uri]",
				HelpName:  "votes",
				Action:    doVotes,
			},
			{
				Name:      "repost",
				Usage:     "repost the post",
				UsageText: "bsky repost [uri]",
				HelpName:  "repost",
				Action:    doRepost,
			},
			{
				Name:      "reposts",
				Usage:     "show reposts of the post",
				UsageText: "bsky reposts [uri]",
				HelpName:  "reposts",
				Action:    doReposts,
			},
			{
				Name:      "follow",
				Usage:     "follow the handle",
				UsageText: "bsky follow [handle]",
				HelpName:  "follow",
				Action:    doFollow,
			},
			{
				Name:      "follows",
				Usage:     "show follows",
				UsageText: "bsky follows",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "follows",
				Action:   doFollows,
			},
			{
				Name:      "followers",
				Usage:     "show followers",
				UsageText: "bsky followres",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "json", Usage: "output JSON"},
				},
				HelpName: "followers",
				Action:   doFollowers,
			},
			{
				Name:      "delete",
				Usage:     "delete the note",
				UsageText: "bsky delete [cid]",
				HelpName:  "delete",
				Action:    doDelete,
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
