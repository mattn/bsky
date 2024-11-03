package pkg

import (
	"context"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type TestStuff struct {
	Manager *XRPCManager
	Client  *xrpc.Client
}

func testSetup() (*TestStuff, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get user home directory")
	}

	configPath := filepath.Join(homeDir, ".config/bsky/config.json")

	cManager := &LocalFileConfigManager{
		Path: configPath,
	}

	config, err := cManager.LoadConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to Load configuration")
	}

	auth := &AuthLocalFile{
		Path: filepath.Join(homeDir, ".config/bsky/jeremylewi.bsky.social.auth"),
	}

	m := &XRPCManager{
		AuthManager: auth,
		Config:      config,
	}

	client, err := m.MakeXRPCC(context.Background())

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to make XRPC client")
	}

	l, err := zap.NewDevelopmentConfig().Build()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create logger")
	}

	zap.ReplaceGlobals(l)

	return &TestStuff{
		Manager: m,
		Client:  client,
	}, nil
}

func Test_Follows(t *testing.T) {
	if os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skipf("Test_StarterPacks is a manual test that is skipped in CICD")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("os.UserHomeDir() = %v, wanted nil", err)
	}

	configPath := filepath.Join(homeDir, ".config/bsky/config.json")

	cManager := &LocalFileConfigManager{
		Path: configPath,
	}

	config, err := cManager.LoadConfig()
	if err != nil {
		t.Fatalf("cManager.LoadConfig() = %v, wanted nil", err)
	}

	auth := &AuthLocalFile{
		Path: filepath.Join(homeDir, ".config/bsky/jeremylewi.bsky.social.auth"),
	}

	m := &XRPCManager{
		AuthManager: auth,
		Config:      config,
	}

	client, err := m.MakeXRPCC(context.Background())
	if err != nil {
		t.Fatalf("m.MakeXRPCC() = %v, wanted nil", err)
	}
	var out strings.Builder
	if err := DoFollows(client, config.Handle, &out); err != nil {
		t.Fatalf("DoFollows() = %v, wanted nil", err)
	}

	t.Logf("Follows:\n%s", out.String())
}
