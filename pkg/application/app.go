package application

import (
	"fmt"
	"github.com/jlewi/goapp-template/pkg/config"
	"github.com/jlewi/monogo/gcp/logging"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"strings"
)

type App struct {
	Config *config.Config
}

// NewApp creates a new application. You should call one more setup/Load functions to properly set it up.
func NewApp() *App {
	return &App{}
}

// LoadConfig loads the config. It takes an optional command. The command allows values to be overwritten from
// the CLI.
func (a *App) LoadConfig(cmd *cobra.Command) error {
	// N.B. at this point we haven't configured any logging so zap just returns the default logger.
	// TODO(jeremy): Should we just initialize the logger without cfg and then reinitialize it after we've read the config?
	if err := config.InitViper(cmd); err != nil {
		return err
	}
	cfg := config.GetConfig()

	if problems := cfg.IsValid(); len(problems) > 0 {
		fmt.Fprintf(os.Stdout, "Invalid configuration; %s\n", strings.Join(problems, "\n"))
		return fmt.Errorf("invalid configuration; fix the problems and then try again")
	}
	a.Config = cfg

	return nil
}

func (a *App) SetupLogging() error {
	// Configure encoder for JSON format
	c := zap.NewDevelopmentConfig()

	if a.Config.Logging.JSON {
		c = zap.NewProductionConfig()
	}
	// Use the keys used by cloud logging
	// https://cloud.google.com/logging/docs/structured-logging
	c.EncoderConfig.LevelKey = logging.SeverityField
	c.EncoderConfig.TimeKey = logging.TimeField
	c.EncoderConfig.MessageKey = logging.MessageField
	// We attach the function key to the logs because that is useful for identifying the function that generated the log.
	c.EncoderConfig.FunctionKey = "function"

	lvl := a.Config.GetLogLevel()
	zapLvl := zap.NewAtomicLevel()

	if err := zapLvl.UnmarshalText([]byte(lvl)); err != nil {
		return errors.Wrapf(err, "Could not convert level %v to ZapLevel", lvl)
	}

	c.Level = zapLvl

	l, err := c.Build()
	if err != nil {
		return errors.Wrap(err, "failed to build logger")
	}
	zap.ReplaceGlobals(l)
	return nil
}

func (a *App) Shutdown() error {
	// Any shutdown code goes here.
	return nil
}
