package cmd

import (
	"fmt"
	"github.com/jlewi/goapp-template/pkg/config"
	"github.com/jlewi/goapp-template/pkg/version"
	"io"

	"github.com/spf13/cobra"
)

func NewVersionCmd(w io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Return version",
		Example: fmt.Sprintf("%s  version", config.AppName),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(w, "%s %s, commit %s, built at %s by %s\n", config.AppName, version.Version, version.Commit, version.Date, version.BuiltBy)
		},
	}
	return cmd
}
