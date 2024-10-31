package pkg

import (
	"encoding/json"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/pkg/errors"
	"os"
)

// AuthManager is an interface that defines methods for managing authentication.
type AuthManager interface {
	// ReadAuth reads the authorization information.
	ReadAuth() (*xrpc.AuthInfo, error)
	// WriteAuth writes the authorization information.
	WriteAuth(*xrpc.AuthInfo) error
}

// AuthLocalStorage reads/writes authorization information to local storage in the browser.
type AuthLocalStorage struct {
	Ctx app.Context
}

const (
	xrpcAuthKey = "xrpcauth"
)

func (a *AuthLocalStorage) ReadAuth() (*xrpc.AuthInfo, error) {
	authInfo := &xrpc.AuthInfo{}
	storage := a.Ctx.LocalStorage()
	err := storage.Get(xrpcAuthKey, authInfo)
	return authInfo, err
}

func (a *AuthLocalStorage) WriteAuth(info *xrpc.AuthInfo) error {
	return a.Ctx.LocalStorage().Set(xrpcAuthKey, info)
}

type AuthLocalFile struct {
	Path string
}

func (a *AuthLocalFile) ReadAuth() (*xrpc.AuthInfo, error) {
	b, err := os.ReadFile(a.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load Auth file: %s", a.Path)
	}

	cfg := &xrpc.AuthInfo{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal Config file: %s", a.Path)
	}
	return cfg, nil
}

func (a *AuthLocalFile) WriteAuth(info *xrpc.AuthInfo) error {
	b, err := json.Marshal(info)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal AuthInfo")
	}
	// TODO(jeremy): Create the directory if it doesn't exist
	if err := os.WriteFile(a.Path, b, 0600); err != nil {
		return errors.Wrapf(err, "cannot write Config file: %s", a.Path)

	}
	return nil
}
