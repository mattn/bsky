package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	lexutil "github.com/bluesky-social/indigo/lex/util"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
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

func DoFollow(client *xrpc.Client, filePath string, w io.Writer) error {
	var fContents []byte
	var err error

	handles := make([]string, 0)
	// Support reading YAML files directly from git.
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		// Handle URL
		resp, err := http.Get(filePath)
		if err != nil {
			return errors.Wrapf(err, "cannot fetch URL %s", filePath)
		}
		defer resp.Body.Close()

		fContents, err = io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrapf(err, "cannot read response body from URL %s", filePath)
		}
	} else {
		// Handle local file
		fContents, err = os.ReadFile(filePath)
		if err != nil {
			return errors.Wrapf(err, "cannot read file %s", filePath)
		}
	}
	list := &FollowList{}
	if err := yaml.Unmarshal(fContents, &list); err != nil {
		return errors.Wrapf(err, "cannot unmarshal FollowList from file %s", filePath)
	}

	for _, a := range list.Accounts {
		handles = append(handles, a.Handle)
	}

	if len(handles) == 0 {
		fmt.Fprintf(w, "No handles found in the follow list in: %s\n", filePath)
	}

	for _, arg := range handles {
		profile, err := bsky.ActorGetProfile(context.TODO(), client, arg)
		if err != nil {
			var xErr *xrpc.Error
			if errors.As(err, &xErr) {
				if 400 == xErr.StatusCode {
					fmt.Fprintf(w, "Profile not found for handle: %s\n", arg)
					continue
				}
			}
			return fmt.Errorf("cannot get profile: %w", err)
		}

		follow := bsky.GraphFollow{
			LexiconTypeID: "app.bsky.graph.follow",
			CreatedAt:     time.Now().Local().Format(time.RFC3339),
			Subject:       profile.Did,
		}

		resp, err := comatproto.RepoCreateRecord(context.TODO(), client, &comatproto.RepoCreateRecord_Input{
			Collection: "app.bsky.graph.follow",
			Repo:       client.Auth.Did,
			Record: &lexutil.LexiconTypeDecoder{
				Val: &follow,
			},
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "Followed: %s; uri: %s", arg, resp.Uri)
	}
	return nil
}
