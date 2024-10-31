package pkg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
