package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/urfave/cli/v2"
)

func doSearch(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	n := cCtx.Int64("n")

	terms := strings.Join(cCtx.Args().Slice(), " ")

	var results []*bsky.FeedDefs_PostView

	var cursor string

	for {
		resp, err := bsky.FeedSearchPosts(context.TODO(), xrpcc, "", cursor, "", "", 100, "", terms, "", "", nil, "", "")
		if err != nil {
			return fmt.Errorf("cannot perform search: %w", err)
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

	sort.Slice(results, func(i, j int) bool {
		ri := timep(results[i].Record.Val.(*bsky.FeedPost).CreatedAt)
		rj := timep(results[j].Record.Val.(*bsky.FeedPost).CreatedAt)
		return ri.Before(rj)
	})
	if int64(len(results)) > n {
		results = results[len(results)-int(n):]
	}

	if cCtx.Bool("json") {
		for _, p := range results {
			checkError(json.NewEncoder(os.Stdout).Encode(p), "Could not encode search results properly")
		}
	} else {
		for _, p := range results {
			printPost(p)
		}
	}

	return nil
}
