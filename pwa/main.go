package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/jlewi/bsctl/pkg"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	outputBoxID = "output-box"
)

// CommandApp represents the chat-like application.
type CommandApp struct {
	app.Compo
	commands []string
	input    string
	manager  *pkg.XRPCManager
}

func (a *CommandApp) Render() app.UI {
	return app.Div().Body(
		// New top toolbar
		app.Div().
			Style("background-color", "#1DA1F2"). // Twitter blue color
			Style("color", "white").
			Style("padding", "10px").
			Style("font-size", "18px").
			Style("font-weight", "bold").
			Style("text-align", "center").
			Text("bsctl: Bluesky Web CLI"),

		// Display area for previous commands
		app.Div().
			ID(outputBoxID).
			Style("border", "1px solid #ddd").
			Style("height", "400px").
			Style("overflow-y", "auto").
			Style("padding", "10px").
			Body(
				app.Range(a.commands).Slice(func(i int) app.UI {
					return app.Div().Text(a.commands[i])
				}),
			),

		// Input area
		app.Div().
			Style("display", "flex").
			Style("margin-top", "10px").
			Body(
				app.Input().
					Type("text").
					Value(a.input).
					Placeholder("Enter command...").
					Style("flex", "1").
					Style("padding", "10px").
					OnChange(a.OnInputChange).OnKeyDown(func(ctx app.Context, e app.Event) {
					if e.Get("key").String() == "Enter" {
						a.OnEnterCommand(ctx, e)
					}
				}),
				app.Button().
					Text("Enter").
					OnClick(a.OnEnterCommand).
					Style("margin-left", "10px").
					Style("padding", "10px"),
			),
		// Warning box
		app.Div().
			Style("border", "2px solid red").
			Style("color", "red").
			Style("padding", "10px").
			Style("margin-top", "20px").
			Style("text-align", "center").
			Body(
				app.Text("Warning: This is a janky POC for distributing CLIs as webapps. For a list of issues check "),
				app.A().Href("https://github.com/jlewi/bsctl/issues").Text("GitHub."),
			),

		// Footer bar
		app.Div().
			Style("background-color", "#f0f0f0").
			Style("padding", "10px").
			Style("margin-top", "20px").
			Style("display", "flex").
			Style("justify-content", "space-between").
			Style("align-items", "center").
			Body(
				// Left empty space
				app.Div(),

				// Center version, commit, and build date information
				app.Div().
					Style("text-align", "center").
					Style("color", "#666").
					Style("font-size", "12px").
					Body(
						app.Div().Text("version: "+pkg.Version),
						app.Div().Text("commit: "+pkg.Commit),
						app.Div().Text("buildDate: "+pkg.Date),
					),
				// Right-aligned links
				app.Div().
					Style("display", "flex").
					Style("align-items", "center").
					Body(
						app.A().
							Href("https://github.com/jlewi/bsctl").
							Target("_blank").
							Style("margin-right", "10px").
							Body(
								app.Img().
									Src("/web/github.svg").
									Alt("GitHub").
									Style("width", "24px").
									Style("height", "24px"),
							),
						app.A().
							Href("https://bsky.app/profile/jeremylewi.bsky.social").
							Target("_blank").
							Body(
								app.Img().
									Src("/web/bluesky_Logo.svg").
									Alt("jeremylewi.bsky.social").
									Style("width", "24px").
									Style("height", "24px"),
							),
					),
			),
	)
}

// OnInputChange updates the input text when the user types.
func (a *CommandApp) OnInputChange(ctx app.Context, e app.Event) {
	a.input = ctx.JSSrc().Get("value").String()
	a.Update()
}

// OnEnterCommand handles command submission.
func (a *CommandApp) OnEnterCommand(ctx app.Context, e app.Event) {
	if a.input != "" {
		// Append the command to the list of commands with a fake output
		// TODO(jeremy): This seems to add extra quotes and doesn't handle the case where we have spaces in
		// the password
		parts := strings.Fields(a.input)
		command := parts[0]
		subCommand := ""
		if len(parts) > 1 {
			subCommand = parts[1]
		}
		err := func() error {
			switch command {
			case "config":
				switch subCommand {
				case "set":
					return a.handleSetConfig(ctx)
				case "get":
					return a.handleGetConfig(ctx)
				default:
					output := fmt.Sprintf("Invalid subcommand %s; must be config get or config set", subCommand)
					a.commands = append(a.commands, output)
					return nil
				}

			case "follow":
				return a.handleFollow(ctx)
			case "follows":
				return a.handleFollows(ctx)
			default:
				// Original behavior for other commands
				output := fmt.Sprintf("Unrecognized command %s", command)
				a.commands = append(a.commands, output)
			}
			return nil
		}()

		if err != nil {
			a.commands = append(a.commands, fmt.Sprintf("Error: %+v", err))
		}

		//output := fmt.Sprintf("Command: %s\nOutput: %s", a.input, fakeCommandExecution(a.input))
		//a.commands = append(a.commands, output)
		a.input = ""
		a.scrollToBottom(ctx)
		a.Update()
	}
}

func (a *CommandApp) handleSetConfig(ctx app.Context) error {
	parts := strings.SplitN(a.input, " ", 3)
	if len(parts) != 3 {
		output := fmt.Sprintf("Invalid command format. Use:\nconfig set <key>=<value>")
		a.commands = append(a.commands, output)
		return nil
	}

	keyValue := parts[2]
	pieces := strings.SplitN(keyValue, "=", 2)
	if len(pieces) != 2 {
		output := fmt.Sprintf("Invalid command format. Use:\nconfig set <key>=<value>")
		a.commands = append(a.commands, output)
		return nil
	}

	key := pieces[0]
	value := pieces[1]

	cfg := &pkg.Config{}
	if err := ctx.LocalStorage().Get("config", cfg); err != nil {
		return errors.Wrapf(err, "failed to read config from local storage")
	}

	switch key {
	case "handle":
		cfg.Handle = value
	case "password":
		cfg.Password = value
	case "bgs":
		cfg.Bgs = value
	case "host":
		cfg.Host = value

	default:
		output := fmt.Sprintf("Invalid key %s; must be handle, password, bgs, or host", key)
		a.commands = append(a.commands, output)
		return nil
	}

	// Set defaults if they aren't set
	if cfg.Bgs == "" {
		cfg.Bgs = "https://bsky.network"
	}

	if cfg.Host == "" {
		cfg.Host = "https://bsky.social"
	}

	if err := ctx.LocalStorage().Set("config", cfg); err != nil {
		return errors.Wrapf(err, "failed to write config to local storage")
	}
	return nil
}

func (a *CommandApp) handleGetConfig(ctx app.Context) error {
	cfg := &pkg.Config{}
	if err := ctx.LocalStorage().Get("config", cfg); err != nil {
		return errors.Wrapf(err, "failed to read config from local storage")
	}

	j, err := json.Marshal(cfg)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal config")
	}
	output := "Config: " + string(j)
	a.commands = append(a.commands, output)
	return nil
}

func (a *CommandApp) handleFollow(ctx app.Context) error {
	m, err := a.getXRPCManager(ctx)
	if err != nil {
		return err
	}

	client, err := m.MakeXRPCC(context.Background())
	if err != nil {
		return err
	}

	parts := strings.Fields(a.input)

	if len(parts) != 2 {
		output := fmt.Sprintf("Invalid command format. Use: follow <URI>")
		a.commands = append(a.commands, output)
		return nil
	}

	var w strings.Builder
	if err := pkg.DoFollow(client, parts[1], &w); err != nil {
		output := fmt.Sprintf("Failed to DoFollows: %+v", err)
		a.commands = append(a.commands, output)
		return nil
	}

	output := fmt.Sprintf("Command: %s\nOutput: %s", a.input, w.String())
	a.commands = append(a.commands, output)
	return nil
}

func (a *CommandApp) handleFollows(ctx app.Context) error {
	m, err := a.getXRPCManager(ctx)
	if err != nil {
		return err
	}

	client, err := m.MakeXRPCC(context.Background())
	if err != nil {
		return err
	}

	var w strings.Builder
	if err := pkg.DoFollows(client, m.Config.Handle, &w); err != nil {
		output := fmt.Sprintf("Failed to DoFollows: %+v", err)
		a.commands = append(a.commands, output)
		return nil
	}

	output := fmt.Sprintf("Command: %s\nOutput: %s", a.input, w.String())
	a.commands = append(a.commands, output)
	return nil
}

func (a *CommandApp) getConfig(ctx app.Context) (*pkg.Config, error) {
	cfg := &pkg.Config{}
	if err := ctx.LocalStorage().Get("config", cfg); err != nil {
		return nil, errors.Wrapf(err, "failed to read config from local storage")
	}
	// Set defaults if they aren't set
	if cfg.Bgs == "" {
		cfg.Bgs = "https://bsky.network"
	}

	if cfg.Host == "" {
		cfg.Host = "https://bsky.social"
	}

	return cfg, nil
}

func (a *CommandApp) getXRPCManager(ctx app.Context) (*pkg.XRPCManager, error) {
	if a.manager != nil {
		return a.manager, nil
	}
	log := zapr.NewLogger(zap.L())
	log.Info("Creating xRPCManager")

	cfg, err := a.getConfig(ctx)

	if err != nil {
		return nil, err
	}

	if cfg.Handle == "" {
		return nil, errors.New("host not set. Run config set handle=<handle> to set it")
	}

	if cfg.Password == "" {
		return nil, errors.New("password not set. Run config set password=password to set it")
	}

	m := pkg.XRPCManager{
		AuthManager: &pkg.AuthLocalStorage{
			Ctx: ctx,
		},
		Config: cfg,
	}

	a.manager = &m
	return &m, nil
}

func (a *CommandApp) scrollToBottom(ctx app.Context) {
	ctx.Async(func() {
		element := app.Window().GetElementByID(outputBoxID)
		if element.Truthy() {
			scrollHeight := element.Get("scrollHeight").Int()
			element.Set("scrollTop", scrollHeight)
		}
	})
}

func main() {
	// We need to configure a logger so that messages will be logged to the console.
	c := zap.NewDevelopmentConfig()
	c.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	newLogger, err := c.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize zap logger; error %v", err))
	}

	zap.ReplaceGlobals(newLogger)
	log := zapr.NewLogger(newLogger)

	// Register the root component.
	bucketName := "/bsctl"
	pkg.LogVersion()
	// N.B. if we run it locally we will serve it on "/"
	// But when we run it on GCS we will serve it on the bucket name. so we add a second route
	log.Info("Registering path", "path", "/")
	app.Route("/", &CommandApp{})
	log.Info("Registering path", "path", bucketName)
	app.Route(bucketName, &CommandApp{})
	log.Info("Registering path", "path", "index.html")
	app.Route("/index.html", &CommandApp{})
	app.RunWhenOnBrowser()

	log.Info("Running code path for server")
	// Once the routes set up, the next thing to do is to either launch the app
	// or the server that serves the app.
	//
	// When executed on the client-side, the RunWhenOnBrowser() function
	// launches the app,  starting a loop that listens for app events and
	// executes client instructions. Since it is a blocking call, the code below
	// it will never be executed.
	//
	// When executed on the server-side, RunWhenOnBrowser() does nothing, which
	// lets room for server implementation without the need for precompiling
	// instructions.
	handler := &app.Handler{
		Name:        "bsctl",
		Description: "WebCLI for BlueSky",
		//Resources:   app.CustomProvider("", "/viewer"),
		//Styles: []string{
		//	"/web/table.css",
		//	"/web/viewer.css",
		//},
		//Env: map[string]string{
		//	logsviewer.APIPrefixEnvVar: "api",
		//},
	}
	buildStatic := os.Getenv("BUILD_STATIC")

	if buildStatic == "" {
		http.Handle("/", handler)

		if err := http.ListenAndServe(":8000", nil); err != nil {
			//log.Fatal(err)
			fmt.Printf("Error starting server: %v\n", err)
		}
	} else {
		// Generate a static website for serving
		// N.B. We need to use a CustomProvider because all the resources will be on
		// https://storage.googleapis.com/bsctl

		handler.Resources = app.CustomProvider("", bucketName)
		// Does GenerateStaticWebsite require absolute paths?
		buildStatic, err = filepath.Abs(buildStatic)
		if err != nil {
			fmt.Printf("Error getting absolute path: %v\n", err)
			return
		}
		if err := app.GenerateStaticWebsite(buildStatic, handler); err != nil {
			fmt.Printf("Error generating static website: %v\n", err)
			return
		}

		fmt.Printf("Static website generated in %s\n", buildStatic)
	}
}
