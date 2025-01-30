package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	comatproto "github.com/bluesky-social/indigo/api/atproto"

	"github.com/urfave/cli/v2"
)

func doListAppPasswords(cCtx *cli.Context) error {
	if cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	passwords, err := comatproto.ServerListAppPasswords(context.TODO(), xrpcc)
	if err != nil {
		return fmt.Errorf("cannot get profile: %w", err)
	}

	if cCtx.Bool("json") {
		for _, password := range passwords.Passwords {
			checkError(json.NewEncoder(os.Stdout).Encode(password), "Could not encode password properly")
		}
		return nil
	}

	for _, password := range passwords.Passwords {
		fmt.Printf("%s (%s)\n", password.Name, password.CreatedAt)
	}
	return nil
}

func doAddAppPassword(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
		input := &comatproto.ServerCreateAppPassword_Input{
			Name: arg,
		}
		password, err := comatproto.ServerCreateAppPassword(context.TODO(), xrpcc, input)
		if err != nil {
			return fmt.Errorf("cannot create app-password: %w", err)
		}

		if cCtx.Bool("json") {
			checkError(json.NewEncoder(os.Stdout).Encode(password), "Could not encode password properly")
		} else {
			fmt.Printf("%s: %s\n", password.Name, password.Password)
		}
	}
	return nil
}

func doRevokeAppPassword(cCtx *cli.Context) error {
	if !cCtx.Args().Present() {
		return cli.ShowSubcommandHelp(cCtx)
	}

	xrpcc, err := makeXRPCC(cCtx)
	if err != nil {
		return fmt.Errorf("cannot create client: %w", err)
	}

	for _, arg := range cCtx.Args().Slice() {
		input := &comatproto.ServerRevokeAppPassword_Input{
			Name: arg,
		}
		err := comatproto.ServerRevokeAppPassword(context.TODO(), xrpcc, input)
		if err != nil {
			return fmt.Errorf("cannot create app-password: %w", err)
		}
	}
	return nil
}
