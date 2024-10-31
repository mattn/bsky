package pkg

import (
	"context"
	"fmt"
	comatproto "github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/util/cliutil"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// XRPCManager is a struct that manages XRPC connections and requests.
type XRPCManager struct {
	AuthManager AuthManager
	Config      *Config
}

func (m *XRPCManager) MakeXRPCC(ctx context.Context) (*xrpc.Client, error) {
	log := zapr.NewLogger(zap.L())

	log.Info("Creating XRPC Client")
	xrpcc := &xrpc.Client{
		Client: cliutil.NewHttpClient(),
		Host:   m.Config.Host,
		Auth:   &xrpc.AuthInfo{Handle: m.Config.Handle},
	}

	auth, err := m.AuthManager.ReadAuth()
	if err == nil && auth.AccessJwt != "" && auth.RefreshJwt != "" {
		log.Info("Auth found, attempting to refresh session")
		xrpcc.Auth = auth
		xrpcc.Auth.AccessJwt = xrpcc.Auth.RefreshJwt
		refresh, err2 := comatproto.ServerRefreshSession(context.TODO(), xrpcc)
		if err2 != nil {
			err = err2
		} else {
			xrpcc.Auth.Did = refresh.Did
			xrpcc.Auth.AccessJwt = refresh.AccessJwt
			xrpcc.Auth.RefreshJwt = refresh.RefreshJwt

			log.Info("Persisting auth information")
			if err := m.AuthManager.WriteAuth(xrpcc.Auth); err != nil {
				return nil, errors.Wrapf(err, "cannot persist authorization information")
			}
		}
	}
	if err != nil || (xrpcc.Auth.AccessJwt == "" || xrpcc.Auth.RefreshJwt == "") {
		log.Info("Auth not found, creating new session")
		auth, err := comatproto.ServerCreateSession(context.TODO(), xrpcc, &comatproto.ServerCreateSession_Input{
			Identifier: xrpcc.Auth.Handle,
			Password:   m.Config.Password,
		})
		if err != nil {
			return nil, fmt.Errorf("cannot create session: %w", err)
		}
		xrpcc.Auth.Did = auth.Did
		xrpcc.Auth.AccessJwt = auth.AccessJwt
		xrpcc.Auth.RefreshJwt = auth.RefreshJwt

		log.Info("New session created, persisting auth information")
		if err := m.AuthManager.WriteAuth(xrpcc.Auth); err != nil {
			return nil, errors.Wrapf(err, "cannot persist authorization information")
		}
	}

	return xrpcc, nil
}
