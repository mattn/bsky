package config

import (
	"encoding/json"
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Note: The application uses viper for configuration management. Viper merges configurations from various sources
//such as files, environment variables, and command line flags. After merging, viper unmarshals the configuration into the Configuration struct, which is then used throughout the application.

const (
	ConfigFlagName = "config"
	LevelFlagName  = "level"
	AppName        = "bsctl"
	ConfigDir      = "." + AppName
)

var (
	// globalV is the global instance of viper
	globalV *viper.Viper
)

// TODO(jeremy): It might be better to put the datastructures defining the configuration into the API package.
// The reason being we might want to share those data structures withother parts of the API (e.g. RPCs).
// However, we should keep the api package free of other dpendencies (e.g. viper, cobra, etc.). So that might
// necessitate some refactoring. We might want to use a separate struct defined here as a wrapper around
// the underlying data structure.

// Config represents the persistent configuration data.
//
// Currently, the format of the data on disk and in memory is identical. In the future, we may modify this to simplify
// changes to the disk format and to store in-memory values that should not be written to disk. Could that be achieved
// by embedding it in a different struct which contains values that shouldn't be serialized?
type Config struct {
	APIVersion string `json:"apiVersion" yaml:"apiVersion" yamltags:"required"`
	Kind       string `json:"kind" yaml:"kind" yamltags:"required"`

	Logging Logging `json:"logging" yaml:"logging"`

	Bgs      string `json:"bgs" yaml:"bgs"`
	Host     string `json:"host" yaml:"host"`
	Handle   string `json:"handle" yaml:"handle"`
	Password string `json:"password" yaml:"password"`
	Prefix   string `json:"prefix" yaml:"prefix"`

	// configFile is the configuration file used
	configFile string
}

type Logging struct {
	Level string `json:"level,omitempty" yaml:"level,omitempty"`
	// Use JSON logging
	JSON bool `json:"json,omitempty" yaml:"json,omitempty"`
}

type LogSink struct {
	// Set to true to write logs in JSON format
	JSON bool `json:"json,omitempty" yaml:"json,omitempty"`
	// Path is the path to write logs to. Use "stderr" to write to stderr.
	// Use gcplogs:///projects/${PROJECT}/logs/${LOGNAME} to write to Google Cloud Logging
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

func (c *Config) GetLogLevel() string {
	if c.Logging.Level == "" {
		return "info"
	}
	return c.Logging.Level
}

// GetConfigFile returns the configuration file
func (c *Config) GetConfigFile() string {
	if c.configFile == "" {
		c.configFile = DefaultConfigFile()
	}
	return c.configFile
}

// GetConfigDir returns the configuration directory
func (c *Config) GetConfigDir() string {
	configFile := c.GetConfigFile()
	if configFile != "" {
		return filepath.Dir(configFile)
	}

	// Since there is no config file we will use the default config directory.
	return binHome()
}

// GetAuthFile returns the file to persist auth information to
func (c *Config) GetAuthFile() string {
	return filepath.Join(c.GetConfigDir(), c.Handle+".auth.json")
}

// IsValid validates the configuration and returns any errors.
func (c *Config) IsValid() []string {
	problems := make([]string, 0, 1)
	return problems
}

// DeepCopy returns a deep copy.
func (c *Config) DeepCopy() Config {
	b, err := json.Marshal(c)
	if err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to marshal config")
		panic(err)
	}
	var copy Config
	if err := json.Unmarshal(b, &copy); err != nil {
		log := zapr.NewLogger(zap.L())
		log.Error(err, "Failed to unmarshal config")
		panic(err)
	}
	return copy
}

// InitViper function is responsible for reading the configuration file and environment variables, if they are set.
// The results are stored in viper. To retrieve a configuration, use the GetConfig function.
// The function accepts a cmd parameter which allows binding to command flags.
func InitViper(cmd *cobra.Command) error {
	// N.B. we need to set globalV because the subsequent call GetConfig will use that viper instance.
	// Would it make sense to combine InitViper and Get into one command that returns a config object?
	// TODO(jeremy): Could we just use viper.GetViper() to get the global instance?
	globalV = viper.New()
	return InitViperInstance(globalV, cmd)
}

// InitViperInstance function is responsible for reading the configuration file and environment variables, if they are set.
// The results are stored in viper. To retrieve a configuration, use the GetConfig function.
// The function accepts a cmd parameter which allows binding to command flags.
func InitViperInstance(v *viper.Viper, cmd *cobra.Command) error {
	// Ref https://github.com/spf13/viper#establishing-defaults
	v.SetEnvPrefix(AppName)

	if v.ConfigFileUsed() == "" {
		// If ConfigFile isn't already set then configure the search parameters.
		// The most likely scenario for it already being set is tests.

		// name of config file (without extension)
		v.SetConfigName("config")
		// make home directory the first search path
		v.AddConfigPath("$HOME/." + AppName)
	}

	// Without the replacer overriding with environment variables doesn't work
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv() // read in environment variables that match

	v.SetDefault("host", "https://bsky.social")
	v.SetDefault("bgs", "https://bsky.social")
	// We need to attach to the command line flag if it was specified.
	keyToflagName := map[string]string{
		ConfigFlagName:             ConfigFlagName,
		"logging." + LevelFlagName: LevelFlagName,
	}

	if cmd != nil {
		for key, flag := range keyToflagName {
			if err := v.BindPFlag(key, cmd.Flags().Lookup(flag)); err != nil {
				return err
			}
		}
	}

	// Ensure the path for the config file path is set
	// Required since we use viper to persist the location of the config file so can save to it.
	// This allows us to overwrite the config file location with the --config flag.
	cfgFile := v.GetString(ConfigFlagName)
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log := zapr.NewLogger(zap.L())
			log.Error(err, "config file not found", "file", cfgFile)
			return nil
		}
		if _, ok := err.(*fs.PathError); ok {
			log := zapr.NewLogger(zap.L())
			log.Error(err, "config file not found", "file", cfgFile)
			return nil
		}
		return err
	}
	return nil
}

// GetConfig returns a configuration created from the viper configuration.
func GetConfig() *Config {
	if globalV == nil {
		// TODO(jeremy): Using a global variable to pass state between InitViper and GetConfig is wonky.
		// It might be better to combine InitViper and GetConfig into a single command that returns a config object.
		// This would also make viper an implementation detail of the config.
		panic("globalV is nil; was InitViper called before calling GetConfig?")
	}
	// We do this as a way to load the configuration while still allowing values to be overwritten by viper
	cfg, err := getConfigFromViper(globalV)
	if err != nil {
		panic(err)
	}
	return cfg
}

func getConfigFromViper(v *viper.Viper) (*Config, error) {
	// We do this as a way to load the configuration while still allowing values to be overwritten by viper
	cfg := &Config{}

	if err := v.Unmarshal(cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal configuration; error %v", err)
	}

	// Set the configFileUsed
	cfg.configFile = v.ConfigFileUsed()
	return cfg, nil
}

func binHome() string {
	log := zapr.NewLogger(zap.L())
	usr, err := user.Current()
	homeDir := ""
	if err != nil {
		log.Error(err, "failed to get current user; falling back to temporary directory for homeDir", "homeDir", os.TempDir())
		homeDir = os.TempDir()
	} else {
		homeDir = usr.HomeDir
	}
	p := filepath.Join(homeDir, ConfigDir)

	return p
}

// Write saves the configuration to a file.
func (c *Config) Write(cfgFile string) error {
	log := zapr.NewLogger(zap.L())
	if cfgFile == "" {
		return errors.Errorf("no config file specified")
	}
	configDir := filepath.Dir(cfgFile)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		log.Info("creating config directory", "dir", configDir)
		if err := os.Mkdir(configDir, 0700); err != nil {
			return errors.Wrapf(err, "Ffailed to create config directory %s", configDir)
		}
	}

	f, err := os.Create(cfgFile)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(f).Encode(c)
}

func DefaultConfigFile() string {
	return binHome() + "/config.yaml"
}
