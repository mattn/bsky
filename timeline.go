package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/events"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/repomgr"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/fatih/color"
	cid "github.com/ipfs/go-cid"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/websocket"
	"github.com/urfave/cli/v2"
)

func doThread(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()
	if !strings.HasPrefix(arg, "at://did:plc:") {
		arg = "at://did:plc:" + arg
	}

	n := cCtx.Int64("n")
	resp, err := bsky.FeedGetPostThread(context.TODO(), xrpcc, n, arg)
	if err != nil {
		return fmt.Errorf("cannot get post thread: %w", err)
	}

	replies := resp.Thread.FeedDefs_ThreadViewPost.Replies
	if cCtx.Bool("json") {
		json.NewEncoder(os.Stdout).Encode(resp.Thread.FeedDefs_ThreadViewPost)
		for _, p := range replies {
			json.NewEncoder(os.Stdout).Encode(p)
		}
		return nil
	}

	for i := 0; i < len(replies)/2; i++ {
		replies[i], replies[len(replies)-i-1] = replies[len(replies)-i-1], replies[i]
	}
	printPost(resp.Thread.FeedDefs_ThreadViewPost.Post)
	for _, r := range replies {
		printPost(r.FeedDefs_ThreadViewPost.Post)
	}
	return nil
}

func doTimeline(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	var feed []*bsky.FeedDefs_FeedViewPost

	n := cCtx.Int64("n")
	handle := cCtx.String("handle")

	var cursor string

	for {
		if handle != "" {
			if handle == "self" {
				handle = xrpcc.Auth.Did
			}
			resp, err := bsky.FeedGetAuthorFeed(context.TODO(), xrpcc, handle, cursor, n)
			if err != nil {
				return fmt.Errorf("cannot get author feed: %w", err)
			}
			feed = append(feed, resp.Feed...)
			if resp.Cursor != nil {
				cursor = *resp.Cursor
			} else {
				cursor = ""
			}
		} else {
			resp, err := bsky.FeedGetTimeline(context.TODO(), xrpcc, "reverse-chronological", cursor, n)
			if err != nil {
				return fmt.Errorf("cannot get timeline: %w", err)
			}
			feed = append(feed, resp.Feed...)
			if resp.Cursor != nil {
				cursor = *resp.Cursor
			} else {
				cursor = ""
			}
		}
		if cursor == "" || int64(len(feed)) > n {
			break
		}
	}

	sort.Slice(feed, func(i, j int) bool {
		ri := timep(feed[i].Post.Record.Val.(*bsky.FeedPost).CreatedAt)
		rj := timep(feed[j].Post.Record.Val.(*bsky.FeedPost).CreatedAt)
		return ri.Before(rj)
	})
	if int64(len(feed)) > n {
		feed = feed[len(feed)-int(n):]
	}
	if cCtx.Bool("json") {
		for _, p := range feed {
			json.NewEncoder(os.Stdout).Encode(p)
		}
	} else {
		for _, p := range feed {
			//if p.Reason != nil {
			//continue
			//}
			printPost(p.Post)
		}
	}

	return nil
}

func doDelete(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
		if !strings.HasPrefix(arg, "at://did:plc:") {
			arg = "at://did:plc:" + arg
		}
		parts := strings.Split(arg, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid post uri: %q", arg)
		}
		rkey := parts[len(parts)-1]
		schema := parts[len(parts)-2]

		err = comatproto.RepoDeleteRecord(context.TODO(), xrpcc, &comatproto.RepoDeleteRecord_Input{
			Repo:       xrpcc.Auth.Did,
			Collection: schema,
			Rkey:       rkey,
		})
		if err != nil {
			return fmt.Errorf("cannot delete post: %w", err)
		}
	}
	return nil
}

func addLink(xrpcc *xrpc.Client, post *bsky.FeedPost, link string) {
	doc, _ := goquery.NewDocument(link)
	var title string
	var description string
	var imgURL string
	if doc != nil {
		title = doc.Find(`title`).Text()
		description, _ = doc.Find(`meta[property="description"]`).Attr("content")
		imgURL, _ = doc.Find(`meta[property="og:image"]`).Attr("content")
		if title == "" {
			title, _ = doc.Find(`meta[property="og:title"]`).Attr("content")
			if title == "" {
				title = link
			}
		}
		if description == "" {
			description, _ = doc.Find(`meta[property="og:description"]`).Attr("content")
			if description == "" {
				description = link
			}
		}
		post.Embed.EmbedExternal = &bsky.EmbedExternal{
			External: &bsky.EmbedExternal_External{
				Description: description,
				Title:       title,
				Uri:         link,
			},
		}
	}
	if imgURL != "" && post.Embed.EmbedExternal != nil {
		resp, err := http.Get(imgURL)
		if err == nil {
			defer resp.Body.Close()
			b, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				resp, err := comatproto.RepoUploadBlob(context.TODO(), xrpcc, bytes.NewReader(b))
				if err == nil {
					post.Embed.EmbedExternal.External.Thumb = &lexutil.LexBlob{
						Ref:      resp.Blob.Ref,
						MimeType: http.DetectContentType(b),
						Size:     resp.Blob.Size,
					}
				}
			}
		}
	}
}

func doPost(cCtx *cli.Context) error {
	stdin := cCtx.Bool("stdin")
	if !stdin && !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}
	text := strings.Join(cCtx.Args().Slice(), " ")
	if stdin {
		b, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		text = string(b)
	}
	if strings.TrimSpace(text) == "" {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	// reply
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

		resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, did, rkey)
		if err != nil {
			return fmt.Errorf("cannot get record: %w", err)
		}
		orig := resp.Value.Val.(*bsky.FeedPost)
		reply = &bsky.FeedPost_ReplyRef{
			Root:   &comatproto.RepoStrongRef{Cid: *resp.Cid, Uri: resp.Uri},
			Parent: &comatproto.RepoStrongRef{Cid: *resp.Cid, Uri: resp.Uri},
		}
		if orig.Reply != nil && orig.Reply.Root != nil {
			reply.Root = &comatproto.RepoStrongRef{Cid: orig.Reply.Root.Cid, Uri: orig.Reply.Root.Uri}
		} else {
			reply.Root = &comatproto.RepoStrongRef{Cid: *resp.Cid, Uri: resp.Uri}
		}
	}

	post := &bsky.FeedPost{
		Text:      text,
		CreatedAt: time.Now().Local().Format(time.RFC3339),
		Reply:     reply,
	}

	// quote
	quoteTo := cCtx.String("q")
	if quoteTo != "" {
		parts := strings.Split(quoteTo, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid post uri: %q", replyTo)
		}
		rkey := parts[len(parts)-1]
		collection := parts[len(parts)-2]
		did := parts[2]

		resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, did, rkey)
		if err != nil {
			return fmt.Errorf("cannot get record: %w", err)
		}

		if post.Embed == nil {
			post.Embed = &bsky.FeedPost_Embed{}
		}
		post.Embed.EmbedRecord = &bsky.EmbedRecord{
			//LexiconTypeID: "app.bsky.feed.post",
			Record: &comatproto.RepoStrongRef{Cid: *resp.Cid, Uri: resp.Uri},
		}
	}

	for _, entry := range extractLinks(text) {
		post.Entities = append(post.Entities, &bsky.FeedPost_Entity{
			Index: &bsky.FeedPost_TextSlice{
				Start: entry.start,
				End:   entry.end,
			},
			Type:  "link",
			Value: entry.text,
		})
		if post.Embed == nil {
			post.Embed = &bsky.FeedPost_Embed{}
		}
		if post.Embed.EmbedExternal == nil {
			addLink(xrpcc, post, entry.text)
		}
	}

	for _, entry := range extractMentions(text) {
		profile, err := bsky.ActorGetProfile(context.TODO(), xrpcc, entry.text)
		if err != nil {
			return err
		}
		post.Entities = append(post.Entities, &bsky.FeedPost_Entity{
			Index: &bsky.FeedPost_TextSlice{
				Start: entry.start,
				End:   entry.end,
			},
			Type:  "mention",
			Value: profile.Did,
		})
	}

	// embeded images
	imageFn := cCtx.StringSlice("image")
	if len(imageFn) > 0 {
		var images []*bsky.EmbedImages_Image
		for _, fn := range imageFn {
			b, err := os.ReadFile(fn)
			if err != nil {
				return fmt.Errorf("cannot read image file: %w", err)
			}
			resp, err := comatproto.RepoUploadBlob(context.TODO(), xrpcc, bytes.NewReader(b))
			if err != nil {
				return fmt.Errorf("cannot upload image file: %w", err)
			}
			images = append(images, &bsky.EmbedImages_Image{
				Alt: filepath.Base(fn),
				Image: &lexutil.LexBlob{
					Ref:      resp.Blob.Ref,
					MimeType: http.DetectContentType(b),
					Size:     resp.Blob.Size,
				},
			})
		}
		if post.Embed == nil {
			post.Embed = &bsky.FeedPost_Embed{}
		}
		post.Embed.EmbedImages = &bsky.EmbedImages{
			Images: images,
		}
	}

	resp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       xrpcc.Auth.Did,
		Record: &lexutil.LexiconTypeDecoder{
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
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
		if !strings.HasPrefix(arg, "at://did:plc:") {
			arg = "at://did:plc:" + arg
		}
		parts := strings.Split(arg, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid post uri: %q", arg)
		}
		rkey := parts[len(parts)-1]
		collection := parts[len(parts)-2]
		did := parts[2]

		resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, did, rkey)
		if err != nil {
			return fmt.Errorf("getting record: %w", err)
		}

		voteResp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.like",
			Repo:       xrpcc.Auth.Did,
			Record: &lexutil.LexiconTypeDecoder{
				Val: &bsky.FeedLike{
					CreatedAt: time.Now().Format("2006-01-02T15:04:05.000Z"),
					Subject:   &comatproto.RepoStrongRef{Uri: resp.Uri, Cid: *resp.Cid},
				},
			},
		})

		if err != nil {
			return fmt.Errorf("cannot create vote: %w", err)
		}
		fmt.Println(voteResp.Uri)
	}

	return nil
}

func doVotes(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()
	if !strings.HasPrefix(arg, "at://did:plc:") {
		arg = "at://did:plc:" + arg
	}
	parts := strings.Split(arg, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid post uri: %q", arg)
	}
	rkey := parts[len(parts)-1]
	collection := parts[len(parts)-2]
	did := parts[2]

	resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, did, rkey)
	if err != nil {
		return fmt.Errorf("getting record: %w", err)
	}

	votes, err := bsky.FeedGetLikes(context.TODO(), xrpcc, *resp.Cid, "", 50, resp.Uri)
	if err != nil {
		return fmt.Errorf("getting votes: %w", err)
	}

	if cCtx.Bool("json") {
		for _, v := range votes.Likes {
			json.NewEncoder(os.Stdout).Encode(v)
		}
		return nil
	}

	for _, v := range votes.Likes {
		fmt.Print("👍 ")
		color.Set(color.FgHiRed)
		fmt.Print(v.Actor.Handle)
		color.Set(color.Reset)
		fmt.Printf(" [%s]", stringp(v.Actor.DisplayName))
		fmt.Printf(" (%v)\n", timep(v.CreatedAt))
	}

	return nil
}

func doRepost(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
		if !strings.HasPrefix(arg, "at://did:plc:") {
			arg = "at://did:plc:" + arg
		}
		parts := strings.Split(arg, "/")
		if len(parts) < 3 {
			return fmt.Errorf("invalid post uri: %q", arg)
		}
		rkey := parts[len(parts)-1]
		collection := parts[len(parts)-2]
		did := parts[2]

		resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, did, rkey)
		if err != nil {
			return fmt.Errorf("getting record: %w", err)
		}

		repost := &bsky.FeedRepost{
			CreatedAt: time.Now().Local().Format(time.RFC3339),
			Subject: &comatproto.RepoStrongRef{
				Uri: resp.Uri,
				Cid: *resp.Cid,
			},
		}
		repostResp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.repost",
			Repo:       xrpcc.Auth.Did,
			Record: &lexutil.LexiconTypeDecoder{
				Val: repost,
			},
		})
		if err != nil {
			return fmt.Errorf("cannot create repost: %w", err)
		}
		fmt.Println(repostResp.Uri)
	}

	return nil
}

func doReposts(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	arg := cCtx.Args().First()
	if !strings.HasPrefix(arg, "at://did:plc:") {
		arg = "at://did:plc:" + arg
	}
	parts := strings.Split(arg, "/")
	if len(parts) < 3 {
		return fmt.Errorf("invalid post uri: %q", arg)
	}
	rkey := parts[len(parts)-1]
	collection := parts[len(parts)-2]
	did := parts[2]

	resp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", collection, did, rkey)
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

func doStream(cCtx *cli.Context) error {
	var host string
	if cCtx.Args().Present() {
		host = cCtx.Args().First()
	} else {
		cfg := cCtx.App.Metadata["config"].(*config)
		u, err := url.Parse(cfg.Host)
		if err != nil {
			return err
		}
		u.Scheme = "wss"
		u.Path = "/xrpc/com.atproto.sync.subscribeRepos"
		host = u.String()
	}
	pattern := cCtx.String("pattern")
	reply := cCtx.String("reply")

	var re *regexp.Regexp
	if pattern != "" {
		var err error
		re, err = regexp.Compile(pattern)
		if err != nil {
			return err
		}
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)

	d := websocket.DefaultDialer
	con, _, err := d.Dial(host, http.Header{})
	if err != nil {
		return fmt.Errorf("dial failure: %w", err)
	}

	defer func() {
		_ = con.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-ch
		cancel()
		con.Close()
	}()

	enc := json.NewEncoder(os.Stdout)
	events.ConsumeRepoStreamLite(ctx, con, func(op repomgr.EventKind, seq int64, path string, did string, rcid *cid.Cid, rec any) error {
		type Rec struct {
			Op   repomgr.EventKind `json:"op"`
			Seq  int64             `json:"seq"`
			Path string            `json:"path"`
			Did  string            `json:"did"`
			Rcid *cid.Cid          `json:"rcid"`
			Rec  any               `json:"rec"`
		}

		orig, isPost := rec.(*bsky.FeedPost)

		if re != nil {
			if !isPost || !re.MatchString(orig.Text) {
				return nil
			}
		}
		if cCtx.Bool("json") {
			enc.Encode(Rec{
				Op:   op,
				Seq:  seq,
				Path: path,
				Did:  did,
				Rcid: rcid,
				Rec:  rec,
			})
		} else if isPost {
			xrpcc, err := makeXRPCC(cCtx)
			if err != nil {
				return fmt.Errorf("cannot create client: %w", err)
			}
			var post bsky.FeedDefs_PostView
			if author, err := bsky.ActorGetProfile(context.TODO(), xrpcc, did); err == nil {
				post.Author = &bsky.ActorDefs_ProfileViewBasic{
					Avatar:      author.Avatar,
					Did:         author.Did,
					DisplayName: author.DisplayName,
					Handle:      author.Handle,
					Labels:      author.Labels,
					Viewer:      author.Viewer,
				}
				post.Record = &lexutil.LexiconTypeDecoder{
					Val: orig,
				}
				printPost(&post)
			}
		}
		if orig != nil && reply != "" {
			xrpcc, err := makeXRPCC(cCtx)
			if err != nil {
				return fmt.Errorf("cannot create client: %w", err)
			}
			parts := strings.Split(path, "/")
			getResp, err := comatproto.RepoGetRecord(context.TODO(), xrpcc, "", parts[0], did, parts[1])
			if err != nil {
				return fmt.Errorf("cannot get record: %w", err)
			}

			orig := getResp.Value.Val.(*bsky.FeedPost)
			replyTo := &bsky.FeedPost_ReplyRef{
				Root:   &comatproto.RepoStrongRef{Cid: *getResp.Cid, Uri: getResp.Uri},
				Parent: &comatproto.RepoStrongRef{Cid: *getResp.Cid, Uri: getResp.Uri},
			}
			if orig.Reply != nil && orig.Reply.Root != nil {
				replyTo.Root = &comatproto.RepoStrongRef{Cid: orig.Reply.Root.Cid, Uri: orig.Reply.Root.Uri}
			} else {
				replyTo.Root = &comatproto.RepoStrongRef{Cid: *getResp.Cid, Uri: getResp.Uri}
			}
			post := &bsky.FeedPost{
				Text:      reply,
				CreatedAt: time.Now().Local().Format(time.RFC3339),
				Reply:     replyTo,
			}

			resp, err := comatproto.RepoCreateRecord(context.TODO(), xrpcc, &comatproto.RepoCreateRecord_Input{
				Collection: "app.bsky.feed.post",
				Repo:       xrpcc.Auth.Did,
				Record: &lexutil.LexiconTypeDecoder{
					Val: post,
				},
			})
			if err != nil {
				log.Println(err, resp.Uri)
			}
		}
		return nil
	})

	return nil
}
