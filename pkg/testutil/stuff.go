package testutil

import (
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/jlewi/bsctl/pkg/application"
	"github.com/pkg/errors"
)

type TestStuff struct {
	App    *application.App
	Client *xrpc.Client
}

func New() (*TestStuff, error) {
	app := application.NewApp()

	if err := app.LoadConfig(nil); err != nil {
		return nil, errors.Wrapf(err, "Failed to loadconfig")
	}

	if err := app.SetupLogging(); err != nil {
		return nil, errors.Wrapf(err, "Failed to setup logging")
	}

	client, err := app.GetXRPCClient()

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to make XRPC client")
	}

	return &TestStuff{
		App:    app,
		Client: client,
	}, nil
}
