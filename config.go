package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func configDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, ".config"), nil
	default:
		return os.UserConfigDir()

	}
}

func loadConfig(profile string) (*config, string, error) {
	dir, err := configDir()
	if err != nil {
		return nil, "", err
	}
	dir = filepath.Join(dir, "bsky")

	var fp string
	if profile == "" {
		fp = filepath.Join(dir, "config.json")
	} else if profile == "?" {
		names, err := filepath.Glob(filepath.Join(dir, "config-*.json"))
		if err != nil {
			return nil, "", err
		}
		for _, name := range names {
			name = filepath.Base(name)
			name = strings.TrimLeft(name[6:len(name)-5], "-")
			fmt.Println(name)
		}
		os.Exit(0)
	} else {
		fp = filepath.Join(dir, "config-"+profile+".json")
	}
	os.MkdirAll(filepath.Dir(fp), 0700)

	b, err := os.ReadFile(fp)
	if err != nil {
		return nil, fp, fmt.Errorf("cannot load config file: %w", err)
	}
	var cfg config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, fp, fmt.Errorf("cannot load config file: %w", err)
	}
	if cfg.Host == "" {
		cfg.Host = "https://bsky.social"
	}
	cfg.dir = dir
	return &cfg, fp, nil
}
