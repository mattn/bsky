package pkg

import (
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
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
