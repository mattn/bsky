package cmd

import (
	"fmt"
	"github.com/go-logr/zapr"
	"github.com/jlewi/bsctl/pkg/api/v1alpha1"
	"github.com/jlewi/bsctl/pkg/application"
	"github.com/jlewi/bsctl/pkg/util"
	"github.com/jlewi/bsctl/pkg/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"sort"
)

// NewDumpCmd creates a command to dump the feed terms so they can be copied and pasted into blueskyfeedcreator
func NewDumpCmd() *cobra.Command {
	// TODO(jeremy): We should update apply to support the image resource.
	cmd := &cobra.Command{
		Use:   "dump <resource.yaml>",
		Short: "Dump the include terms in a feed resource as text so they can be copied into blueskyfeedcreator",
		Run: func(cmd *cobra.Command, args []string) {
			err := func() error {
				app := application.NewApp()
				defer app.Shutdown()
				if err := app.LoadConfig(cmd); err != nil {
					return err
				}
				if err := app.SetupLogging(); err != nil {
					return err
				}
				log := zapr.NewLogger(zap.L())
				if len(args) == 0 {
					log.Info("apply takes at least one argument which should be the file or directory YAML to apply.")
					return errors.New("apply takes at least one argument which should be the file or directory YAML to apply.")
				}
				version.Log()

				for _, arg := range args {
					nodes, err := util.ReadYaml(arg)
					if err != nil {
						return errors.Wrapf(err, "Failed to read YAML file; %v", arg)
					}

					for _, node := range nodes {
						if node.GetKind() != v1alpha1.FeedKind {
							log.Info("Skipping resource; not a feed", "kind", node.GetKind(), "name", node.GetName())
							continue
						}

						f := &v1alpha1.Feed{}

						if err := node.YNode().Decode(f); err != nil {
							return errors.Wrapf(err, "Failed to decode feed; %v", node.GetName())
						}

						sort.Strings(f.Include)
						fmt.Fprintf(os.Stdout, "Feed:%v\n", f.Metadata.Name)
						fmt.Fprintf(os.Stdout, "Include Terms:\n")
						for _, term := range f.Include {
							fmt.Fprintf(os.Stdout, "%v\n", term)
						}
					}
				}
				return nil
			}()
			if err != nil {
				fmt.Printf("Error running apply;\n %+v\n", err)
				os.Exit(1)
			}
		},
	}

	return cmd
}
