package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bluesky-social/indigo/api/bsky"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"os"
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

func DoFollows(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := MakeXRPCC(cCtx)
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
				fmt.Printf(" [%s] ", Stringp(f.DisplayName))
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
