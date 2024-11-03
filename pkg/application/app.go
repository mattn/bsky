package application

import (
	"context"
	"fmt"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/go-logr/zapr"
	"github.com/jlewi/bsctl/pkg/api/v1alpha1"
	"github.com/jlewi/bsctl/pkg/config"
	"github.com/jlewi/bsctl/pkg/controllers"
	"github.com/jlewi/bsctl/pkg/lists"
	"github.com/jlewi/bsctl/pkg/util"
	"github.com/jlewi/monogo/gcp/logging"
	"github.com/jlewi/monogo/helpers"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"os"
	"strings"
)

// TODO(jeremy): How do we turn App into an interface such that we could simultaneously support a CLI and WebCLI

type App struct {
	Config   *config.Config
	Registry *controllers.Registry
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

// SetupRegistry sets up the registry with a list of registered controllers
//
// TODO(jeremy): How do we avoid circular dependencies? We'd like to be able to use
// the Application class in manual tests that live side by side with the relevant code.
// see also: https://github.com/jlewi/goapp-template/issues/1
func (a *App) SetupRegistry() error {
	if a.Config == nil {
		return errors.New("Config is nil; call LoadConfig first")
	}
	a.Registry = &controllers.Registry{}

	client, err := a.GetXRPCClient()
	if err != nil {
		return err
	}
	// Register controllers
	listController, err := lists.NewAccountListController(client)
	if err != nil {
		return err
	}
	if err := a.Registry.Register(v1alpha1.AccountListGVK, listController); err != nil {
		return err
	}

	return nil
}

// ApplyPaths applies the resources in the specified paths.
// Paths can be files or directories.
func (a *App) ApplyPaths(ctx context.Context, inPaths []string) error {
	log := util.LogFromContext(ctx)

	paths := make([]string, 0, len(inPaths))
	for _, resourcePath := range inPaths {
		newPaths, err := util.FindYamlFiles(resourcePath)
		if err != nil {
			log.Error(err, "Failed to find YAML files", "path", resourcePath)
			return err
		}

		paths = append(paths, newPaths...)
	}

	for _, path := range paths {
		err := a.apply(ctx, path)
		if err != nil {
			log.Error(err, "Apply failed", "path", path)
		}
	}

	return nil
}

func (a *App) apply(ctx context.Context, path string) error {
	if a.Registry == nil {
		return errors.New("Registry is nil; call SetupRegistry first")
	}

	log := zapr.NewLogger(zap.L())
	log.Info("Reading file", "path", path)
	rNodes, err := util.ReadYaml(path)
	if err != nil {
		return err
	}

	allErrors := &helpers.ListOfErrors{
		Causes: []error{},
	}

	for _, n := range rNodes {
		m, err := n.GetMeta()
		if err != nil {
			log.Error(err, "Failed to get metadata", "n", n)
			continue
		}
		log.Info("Read resource", "meta", m)

		gvk := schema.FromAPIVersionAndKind(m.APIVersion, m.Kind)
		controller, err := a.Registry.GetController(gvk)
		if err != nil {
			log.Error(err, "Unsupported kind", "gvk", gvk)
			allErrors.AddCause(err)
			continue
		}

		if err := controller.ReconcileNode(ctx, n); err != nil {
			log.Error(err, "Failed to reconcile resource", "name", m.Name, "namespace", m.Namespace, "gvk", gvk)
			allErrors.AddCause(err)
		}

	}

	if len(allErrors.Causes) == 0 {
		return nil
	}
	allErrors.Final = fmt.Errorf("failed to apply one or more resources")
	return allErrors
}

func (a *App) GetXRPCClient() (*xrpc.Client, error) {
	if a.Config == nil {
		return nil, errors.WithStack(errors.New("Config is nil; call LoadConfig first"))
	}
	authM, err := a.GetAuthManager()
	if err != nil {
		return nil, err
	}

	m := &XRPCManager{
		Config:      a.Config,
		AuthManager: authM,
	}

	return m.CreateClient(context.Background())
}

func (a *App) GetAuthManager() (AuthManager, error) {
	if a.Config == nil {
		return nil, errors.WithStack(errors.New("Config is nil; call LoadConfig first"))
	}
	m := &AuthLocalFile{
		Path: a.Config.GetAuthFile(),
	}
	return m, nil
}

func (a *App) Shutdown() error {
	// Any shutdown code goes here.
	return nil
}
