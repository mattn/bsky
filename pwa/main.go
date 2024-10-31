package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

// CommandApp represents the chat-like application.
type CommandApp struct {
	app.Compo
	commands []string
	input    string
}

func (a *CommandApp) Render() app.UI {
	return app.Div().Body(
		// Display area for previous commands
		app.Div().
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
					OnChange(a.OnInputChange),

				app.Button().
					Text("Enter").
					OnClick(a.OnEnterCommand).
					Style("margin-left", "10px").
					Style("padding", "10px"),
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
		if len(parts) >= 3 && parts[0] == "login" {
			handle := parts[1]
			password := parts[2]

			// Store handle and password in local storage
			ctx.LocalStorage().Set("handle", handle)
			ctx.LocalStorage().Set("password", password)

			output := fmt.Sprintf("Command: %s\nOutput: Login credentials stored", a.input)
			a.commands = append(a.commands, output)
		} else {
			// Original behavior for other commands
			output := fmt.Sprintf("Command: %s\nOutput: %s", a.input, fakeCommandExecution(a.input))
			a.commands = append(a.commands, output)
		}

		output := fmt.Sprintf("Command: %s\nOutput: %s", a.input, fakeCommandExecution(a.input))
		a.commands = append(a.commands, output)
		a.input = ""
		a.Update()
	}
}

// fakeCommandExecution simulates executing a command and returns a response.
func fakeCommandExecution(command string) string {
	time.Sleep(500 * time.Millisecond) // Simulate some processing delay
	return fmt.Sprintf("Executed command '%s' successfully.", command)
}

func main() {
	// Register the root component.
	app.Route("/", &CommandApp{})
	app.RunWhenOnBrowser()

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
	http.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		//Resources:   app.CustomProvider("", "/viewer"),
		//Styles: []string{
		//	"/web/table.css",
		//	"/web/viewer.css",
		//},
		//Env: map[string]string{
		//	logsviewer.APIPrefixEnvVar: "api",
		//},
	})

	if err := http.ListenAndServe(":8000", nil); err != nil {
		//log.Fatal(err)
		fmt.Printf("Error starting server: %v\n", err)
	}
}
