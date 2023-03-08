package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	cliutil "github.com/bluesky-social/indigo/cmd/gosky/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/fatih/color"

	"github.com/urfave/cli/v2"
)

func printPost(p *bsky.FeedPost_View) {
	rec := p.Record.Val.(*bsky.FeedPost)
	color.Set(color.FgHiRed)
	fmt.Print(p.Author.Handle)
	color.Set(color.Reset)
	fmt.Printf(" [%s]", stringp(p.Author.DisplayName))
	fmt.Printf(" (%v)\n", ltime(rec.CreatedAt))
	if rec.Entities != nil {
		sort.Slice(rec.Entities, func(i, j int) bool {
			return rec.Entities[i].Index.Start < rec.Entities[j].Index.Start
		})
		rs := []rune(rec.Text)
		off := int64(0)
		for _, e := range rec.Entities {
			if e.Index.Start < 0 {
				e.Index.Start = 0
			}
			if e.Index.End > int64(len(rs)) {
				e.Index.End = int64(len(rs))
			}
			fmt.Print(string(rs[off:e.Index.Start]))
			if e.Type == "mention" {
				color.Set(color.Bold)
			} else {
				color.Set(color.Underline)
			}
			fmt.Print(string(rs[e.Index.Start:e.Index.End]))
			color.Set(color.Reset)
			off = e.Index.End
		}
		if off < int64(len(rs)) {
			fmt.Print(string(rs[off:]))
		}
		fmt.Println()
		//for _, e := range rec.Entities {
		//	fmt.Printf(" {%s}\n", e.Value)
		//}
	} else {
		fmt.Println(rec.Text)
	}
	if p.Embed != nil {
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
	if rec.Reply != nil && rec.Reply.Parent != nil {
		fmt.Print(" > ")
		color.Set(color.FgBlue)
		fmt.Println(rec.Reply.Parent.Uri)
		color.Set(color.Reset)
	}
	fmt.Print(" - ")
	color.Set(color.FgBlue)
	fmt.Println(p.Uri)
	color.Set(color.Reset)
	fmt.Println()
}

func ltime(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.000Z", s)
	if err != nil {
		return time.Now()
	}
	return t.Local()
}

func stringp(s *string) string {
	if s == nil {
		return ""
	}
	return *s
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
