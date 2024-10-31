package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/bluesky-social/indigo/xrpc"
	"io"
)

// FollowList is a data structure to hold a list of folks to follow
type FollowList struct {
	APIVersion string    `json:"apiVersion" yaml:"apiVersion"`
	Kind       string    `json:"kind" yaml:"kind"`
	Accounts   []Account `json:"accounts" yaml:"accounts"`
}

type Account struct {
	Handle string `json:"handle" yaml:"handle"`
}

func DoFollows(client *xrpc.Client, handle string, w io.Writer) error {
	var cursor string
	for {
		follows, err := bsky.GraphGetFollows(context.TODO(), client, handle, cursor, 100)
		if err != nil {
			return fmt.Errorf("getting record: %w", err)
		}

		for _, f := range follows.Follows {
			json.NewEncoder(w).Encode(f)
		}
		if follows.Cursor == nil {
			break
		}
		cursor = *follows.Cursor
	}
	return nil
}
