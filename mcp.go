package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	cliutil "github.com/bluesky-social/indigo/util/cliutil"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/urfave/cli/v2"
)

func makeXRPCCFromConfig(cfg *config) (*xrpc.Client, error) {
	xrpcc := &xrpc.Client{
		Client: cliutil.NewHttpClient(),
		Host:   cfg.Host,
		Auth:   &xrpc.AuthInfo{Handle: cfg.Handle},
	}

	auth, err := cliutil.ReadAuth(filepath.Join(cfg.dir, cfg.prefix+cfg.Handle+".auth"))
	if err == nil {
		xrpcc.Auth = auth
		xrpcc.Auth.AccessJwt = xrpcc.Auth.RefreshJwt
		refresh, err2 := comatproto.ServerRefreshSession(context.TODO(), xrpcc)
		if err2 != nil {
			err = err2
		} else {
			xrpcc.Auth.Did = refresh.Did
			xrpcc.Auth.AccessJwt = refresh.AccessJwt
			xrpcc.Auth.RefreshJwt = refresh.RefreshJwt
		}
	}
	if err != nil {
		input := &comatproto.ServerCreateSession_Input{
			Identifier: xrpcc.Auth.Handle,
			Password:   cfg.Password,
		}
		auth, err := comatproto.ServerCreateSession(context.TODO(), xrpcc, input)
		if err != nil {
			return nil, fmt.Errorf("cannot create session: %w", err)
		}
		xrpcc.Auth.Did = auth.Did
		xrpcc.Auth.AccessJwt = auth.AccessJwt
		xrpcc.Auth.RefreshJwt = auth.RefreshJwt
	}

	return xrpcc, nil
}

func doMcp(cCtx *cli.Context) error {
	cfg := cCtx.App.Metadata["config"].(*config)

	s := server.NewMCPServer(name, version)

	// timeline
	s.AddTool(mcp.NewTool("bluesky_timeline",
		mcp.WithDescription("Show timeline posts from Bluesky"),
		mcp.WithNumber("n",
			mcp.Description("Number of posts to retrieve"),
			mcp.DefaultNumber(30),
		),
		mcp.WithString("handle",
			mcp.Description("User handle to get timeline for (empty for own timeline)"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		n := mcp.ParseInt64(request, "n", 30)
		handle := mcp.ParseString(request, "handle", "")

		var feed []*bsky.FeedDefs_FeedViewPost
		var cursor string
		for {
			if handle != "" {
				if handle == "self" {
					handle = xrpcc.Auth.Did
				}
				resp, err := bsky.FeedGetAuthorFeed(ctx, xrpcc, handle, cursor, "", false, n)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("cannot get author feed: %v", err)), nil
				}
				feed = append(feed, resp.Feed...)
				if resp.Cursor != nil {
					cursor = *resp.Cursor
				} else {
					cursor = ""
				}
			} else {
				resp, err := bsky.FeedGetTimeline(ctx, xrpcc, "reverse-chronological", cursor, n)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("cannot get timeline: %v", err)), nil
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

		if int64(len(feed)) > n {
			feed = feed[len(feed)-int(n):]
		}

		var posts []map[string]any
		for _, p := range feed {
			rec := p.Post.Record.Val.(*bsky.FeedPost)
			post := map[string]any{
				"uri":        p.Post.Uri,
				"author":     p.Post.Author.Handle,
				"displayName": stringp(p.Post.Author.DisplayName),
				"text":       rec.Text,
				"createdAt":  rec.CreatedAt,
				"likeCount":  int64p(p.Post.LikeCount),
				"repostCount": int64p(p.Post.RepostCount),
				"replyCount": int64p(p.Post.ReplyCount),
			}
			posts = append(posts, post)
		}

		b, err := json.MarshalIndent(posts, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	})

	// post
	s.AddTool(mcp.NewTool("bluesky_post",
		mcp.WithDescription("Create a new post on Bluesky"),
		mcp.WithString("text",
			mcp.Description("Text content of the post"),
			mcp.Required(),
		),
		mcp.WithString("reply",
			mcp.Description("URI of the post to reply to"),
		),
		mcp.WithString("quote",
			mcp.Description("URI of the post to quote"),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		text := mcp.ParseString(request, "text", "")
		if strings.TrimSpace(text) == "" {
			return mcp.NewToolResultError("text is required"), nil
		}

		post := &bsky.FeedPost{
			Text:      text,
			CreatedAt: time.Now().Local().Format(time.RFC3339),
		}

		// reply
		replyTo := mcp.ParseString(request, "reply", "")
		if replyTo != "" {
			parts := strings.Split(replyTo, "/")
			if len(parts) < 3 {
				return mcp.NewToolResultError(fmt.Sprintf("invalid post uri: %q", replyTo)), nil
			}
			rkey := parts[len(parts)-1]
			collection := parts[len(parts)-2]
			did := parts[2]

			resp, err := comatproto.RepoGetRecord(ctx, xrpcc, "", collection, did, rkey)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("cannot get record: %v", err)), nil
			}
			orig := resp.Value.Val.(*bsky.FeedPost)
			reply := &bsky.FeedPost_ReplyRef{
				Root:   &comatproto.RepoStrongRef{Cid: *resp.Cid, Uri: resp.Uri},
				Parent: &comatproto.RepoStrongRef{Cid: *resp.Cid, Uri: resp.Uri},
			}
			if orig.Reply != nil && orig.Reply.Root != nil {
				reply.Root = &comatproto.RepoStrongRef{Cid: orig.Reply.Root.Cid, Uri: orig.Reply.Root.Uri}
			}
			post.Reply = reply
		}

		// quote
		quoteTo := mcp.ParseString(request, "quote", "")
		if quoteTo != "" {
			parts := strings.Split(quoteTo, "/")
			if len(parts) < 3 {
				return mcp.NewToolResultError(fmt.Sprintf("invalid post uri: %q", quoteTo)), nil
			}
			rkey := parts[len(parts)-1]
			collection := parts[len(parts)-2]
			did := parts[2]

			resp, err := comatproto.RepoGetRecord(ctx, xrpcc, "", collection, did, rkey)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("cannot get record: %v", err)), nil
			}
			if post.Embed == nil {
				post.Embed = &bsky.FeedPost_Embed{}
			}
			post.Embed.EmbedRecord = &bsky.EmbedRecord{
				Record: &comatproto.RepoStrongRef{Cid: *resp.Cid, Uri: resp.Uri},
			}
		}

		// facets
		for _, entry := range extractLinksBytes(text) {
			post.Facets = append(post.Facets, &bsky.RichtextFacet{
				Features: []*bsky.RichtextFacet_Features_Elem{
					{
						RichtextFacet_Link: &bsky.RichtextFacet_Link{
							Uri: entry.text,
						},
					},
				},
				Index: &bsky.RichtextFacet_ByteSlice{
					ByteStart: entry.start,
					ByteEnd:   entry.end,
				},
			})
		}

		for _, entry := range extractMentionsBytes(text) {
			profile, err := bsky.ActorGetProfile(ctx, xrpcc, entry.text)
			if err != nil {
				continue
			}
			post.Facets = append(post.Facets, &bsky.RichtextFacet{
				Features: []*bsky.RichtextFacet_Features_Elem{
					{
						RichtextFacet_Mention: &bsky.RichtextFacet_Mention{
							Did: profile.Did,
						},
					},
				},
				Index: &bsky.RichtextFacet_ByteSlice{
					ByteStart: entry.start,
					ByteEnd:   entry.end,
				},
			})
		}

		for _, entry := range extractTagsBytes(text) {
			post.Facets = append(post.Facets, &bsky.RichtextFacet{
				Features: []*bsky.RichtextFacet_Features_Elem{
					{
						RichtextFacet_Tag: &bsky.RichtextFacet_Tag{
							Tag: entry.text,
						},
					},
				},
				Index: &bsky.RichtextFacet_ByteSlice{
					ByteStart: entry.start,
					ByteEnd:   entry.end,
				},
			})
		}

		resp, err := comatproto.RepoCreateRecord(ctx, xrpcc, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.post",
			Repo:       xrpcc.Auth.Did,
			Record:     &lexutil.LexiconTypeDecoder{Val: post},
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot create post: %v", err)), nil
		}

		result := map[string]string{
			"uri": resp.Uri,
			"cid": resp.Cid,
		}
		b, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	// search
	s.AddTool(mcp.NewTool("bluesky_search",
		mcp.WithDescription("Search posts on Bluesky"),
		mcp.WithString("query",
			mcp.Description("Search terms"),
			mcp.Required(),
		),
		mcp.WithNumber("n",
			mcp.Description("Maximum number of results"),
			mcp.DefaultNumber(30),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		query := mcp.ParseString(request, "query", "")
		n := mcp.ParseInt64(request, "n", 30)

		var results []*bsky.FeedDefs_PostView
		var cursor string
		for {
			resp, err := bsky.FeedSearchPosts(ctx, xrpcc, "", cursor, "", "", 100, "", query, "", "", nil, "", "")
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("cannot search: %v", err)), nil
			}
			if resp.Cursor != nil {
				cursor = *resp.Cursor
			} else {
				cursor = ""
			}
			results = append(results, resp.Posts...)
			if cursor == "" || int64(len(results)) > n {
				break
			}
		}

		if int64(len(results)) > n {
			results = results[:n]
		}

		var posts []map[string]any
		for _, p := range results {
			rec := p.Record.Val.(*bsky.FeedPost)
			post := map[string]any{
				"uri":        p.Uri,
				"author":     p.Author.Handle,
				"displayName": stringp(p.Author.DisplayName),
				"text":       rec.Text,
				"createdAt":  rec.CreatedAt,
				"likeCount":  int64p(p.LikeCount),
				"repostCount": int64p(p.RepostCount),
				"replyCount": int64p(p.ReplyCount),
			}
			posts = append(posts, post)
		}

		b, err := json.MarshalIndent(posts, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	})

	// show-profile
	s.AddTool(mcp.NewTool("bluesky_show_profile",
		mcp.WithDescription("Show a Bluesky user profile"),
		mcp.WithString("handle",
			mcp.Description("User handle (empty for own profile)"),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		handle := mcp.ParseString(request, "handle", "")
		if handle == "" {
			handle = xrpcc.Auth.Handle
		}

		profile, err := bsky.ActorGetProfile(ctx, xrpcc, handle)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot get profile: %v", err)), nil
		}

		result := map[string]any{
			"did":         profile.Did,
			"handle":      profile.Handle,
			"displayName": stringp(profile.DisplayName),
			"description": stringp(profile.Description),
			"follows":     int64p(profile.FollowsCount),
			"followers":   int64p(profile.FollowersCount),
			"avatar":      stringp(profile.Avatar),
			"banner":      stringp(profile.Banner),
		}

		b, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	})

	// search-actors
	s.AddTool(mcp.NewTool("bluesky_search_actors",
		mcp.WithDescription("Search for Bluesky users"),
		mcp.WithString("query",
			mcp.Description("Search terms"),
			mcp.Required(),
		),
		mcp.WithNumber("n",
			mcp.Description("Maximum number of results"),
			mcp.DefaultNumber(50),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		query := mcp.ParseString(request, "query", "")
		n := mcp.ParseInt64(request, "n", 50)

		result, err := bsky.ActorSearchActors(ctx, xrpcc, "", n, query, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot search actors: %v", err)), nil
		}

		var actors []map[string]any
		for _, actor := range result.Actors {
			a := map[string]any{
				"did":         actor.Did,
				"handle":      actor.Handle,
				"displayName": stringp(actor.DisplayName),
				"description": stringp(actor.Description),
			}
			actors = append(actors, a)
		}

		b, err := json.MarshalIndent(actors, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	})

	// thread
	s.AddTool(mcp.NewTool("bluesky_thread",
		mcp.WithDescription("Show a post thread on Bluesky"),
		mcp.WithString("uri",
			mcp.Description("URI of the post"),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(true),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		uri := mcp.ParseString(request, "uri", "")
		if !strings.HasPrefix(uri, "at://did:plc:") {
			uri = "at://did:plc:" + uri
		}

		resp, err := bsky.FeedGetPostThread(ctx, xrpcc, 0, 30, uri)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot get thread: %v", err)), nil
		}

		var posts []map[string]any
		if resp.Thread.FeedDefs_ThreadViewPost != nil {
			posts = append(posts, threadPostToMap(resp.Thread.FeedDefs_ThreadViewPost.Post))
			for _, r := range resp.Thread.FeedDefs_ThreadViewPost.Replies {
				if r.FeedDefs_ThreadViewPost != nil {
					posts = append(posts, threadPostToMap(r.FeedDefs_ThreadViewPost.Post))
				}
			}
		}

		b, err := json.MarshalIndent(posts, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	})

	// notification
	s.AddTool(mcp.NewTool("bluesky_notification",
		mcp.WithDescription("Show Bluesky notifications"),
		mcp.WithReadOnlyHintAnnotation(true),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		notifs, err := bsky.NotificationListNotifications(ctx, xrpcc, "", 50, false, nil, "")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot get notifications: %v", err)), nil
		}

		var items []map[string]any
		for _, n := range notifs.Notifications {
			item := map[string]any{
				"author":     n.Author.Handle,
				"displayName": stringp(n.Author.DisplayName),
				"reason":     n.Reason,
				"uri":        n.Uri,
				"isRead":     n.IsRead,
				"indexedAt":  n.IndexedAt,
			}
			items = append(items, item)
		}

		b, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(string(b)), nil
	})

	// like
	s.AddTool(mcp.NewTool("bluesky_like",
		mcp.WithDescription("Like a post on Bluesky"),
		mcp.WithString("uri",
			mcp.Description("URI of the post to like"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		uri := mcp.ParseString(request, "uri", "")

		parts := strings.Split(uri, "/")
		if len(parts) < 3 {
			return mcp.NewToolResultError(fmt.Sprintf("invalid post uri: %q", uri)), nil
		}
		rkey := parts[len(parts)-1]
		collection := parts[len(parts)-2]
		did := parts[2]

		resp, err := comatproto.RepoGetRecord(ctx, xrpcc, "", collection, did, rkey)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot get record: %v", err)), nil
		}

		likeResp, err := comatproto.RepoCreateRecord(ctx, xrpcc, &comatproto.RepoCreateRecord_Input{
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
			return mcp.NewToolResultError(fmt.Sprintf("cannot like: %v", err)), nil
		}

		result := map[string]string{"uri": likeResp.Uri}
		b, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	// repost
	s.AddTool(mcp.NewTool("bluesky_repost",
		mcp.WithDescription("Repost a post on Bluesky"),
		mcp.WithString("uri",
			mcp.Description("URI of the post to repost"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		uri := mcp.ParseString(request, "uri", "")

		parts := strings.Split(uri, "/")
		if len(parts) < 3 {
			return mcp.NewToolResultError(fmt.Sprintf("invalid post uri: %q", uri)), nil
		}
		rkey := parts[len(parts)-1]
		collection := parts[len(parts)-2]
		did := parts[2]

		resp, err := comatproto.RepoGetRecord(ctx, xrpcc, "", collection, did, rkey)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot get record: %v", err)), nil
		}

		repostResp, err := comatproto.RepoCreateRecord(ctx, xrpcc, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.feed.repost",
			Repo:       xrpcc.Auth.Did,
			Record: &lexutil.LexiconTypeDecoder{
				Val: &bsky.FeedRepost{
					CreatedAt: time.Now().Local().Format(time.RFC3339),
					Subject:   &comatproto.RepoStrongRef{Uri: resp.Uri, Cid: *resp.Cid},
				},
			},
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot repost: %v", err)), nil
		}

		result := map[string]string{"uri": repostResp.Uri}
		b, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	// follow
	s.AddTool(mcp.NewTool("bluesky_follow",
		mcp.WithDescription("Follow a user on Bluesky"),
		mcp.WithString("handle",
			mcp.Description("Handle of the user to follow"),
			mcp.Required(),
		),
	), func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		xrpcc, err := makeXRPCCFromConfig(cfg)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		handle := mcp.ParseString(request, "handle", "")

		profile, err := bsky.ActorGetProfile(ctx, xrpcc, handle)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot get profile: %v", err)), nil
		}

		followResp, err := comatproto.RepoCreateRecord(ctx, xrpcc, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.graph.follow",
			Repo:       xrpcc.Auth.Did,
			Record: &lexutil.LexiconTypeDecoder{
				Val: &bsky.GraphFollow{
					LexiconTypeID: "app.bsky.graph.follow",
					CreatedAt:     time.Now().Local().Format(time.RFC3339),
					Subject:       profile.Did,
				},
			},
		})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("cannot follow: %v", err)), nil
		}

		result := map[string]string{"uri": followResp.Uri}
		b, _ := json.MarshalIndent(result, "", "  ")
		return mcp.NewToolResultText(string(b)), nil
	})

	return server.ServeStdio(s)
}

func threadPostToMap(p *bsky.FeedDefs_PostView) map[string]any {
	rec := p.Record.Val.(*bsky.FeedPost)
	return map[string]any{
		"uri":         p.Uri,
		"author":      p.Author.Handle,
		"displayName": stringp(p.Author.DisplayName),
		"text":        rec.Text,
		"createdAt":   rec.CreatedAt,
		"likeCount":   int64p(p.LikeCount),
		"repostCount": int64p(p.RepostCount),
		"replyCount":  int64p(p.ReplyCount),
	}
}
