package pkg

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	Bgs      string `json:"bgs"`
	Host     string `json:"host"`
	Handle   string `json:"handle"`
	Password string `json:"password"`
	dir      string
	Verbose  bool
	Prefix   string
}

func ConfigDir() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, ".Config"), nil
	default:
		return os.UserConfigDir()

	}
}

func LoadConfig(profile string) (*Config, string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, "", err
	}
	dir = filepath.Join(dir, "bsky")

	var fp string
	if profile == "" {
		fp = filepath.Join(dir, "Config.json")
	} else if profile == "?" {
		names, err := filepath.Glob(filepath.Join(dir, "Config-*.json"))
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
		fp = filepath.Join(dir, "Config-"+profile+".json")
	}
	os.MkdirAll(filepath.Dir(fp), 0700)

	b, err := os.ReadFile(fp)
	if err != nil {
		return nil, fp, fmt.Errorf("cannot load Config file: %w", err)
	}
	var cfg Config
	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return nil, fp, fmt.Errorf("cannot load Config file: %w", err)
	}
	if cfg.Host == "" {
		cfg.Host = "https://bsky.social"
	}
	cfg.dir = dir
	return &cfg, fp, nil
}

type ConfigManager interface {
	LoadConfig() (*Config, error)
	SaveConfig(*Config) error
}

type LocalFileConfigManager struct {
	Path string
}

func (m *LocalFileConfigManager) LoadConfig() (*Config, error) {
	b, err := os.ReadFile(m.Path)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot load Config file: %s", m.Path)
	}

	cfg := &Config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, errors.Wrapf(err, "cannot unmarshal Config file: %s", m.Path)
	}
	return cfg, nil
}

func (m *LocalFileConfigManager) SaveConfig(cfg *Config) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrapf(err, "cannot marshal Config")
	}
	// TODO(jeremy): Create the directory if it doesn't exist
	if err := os.WriteFile(m.Path, b, 0600); err != nil {
		return errors.Wrapf(err, "cannot write Config file: %s", m.Path)

	}
	return nil
}
