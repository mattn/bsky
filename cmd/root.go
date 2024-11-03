package cmd

import (
	"fmt"
	"github.com/jlewi/bsctl/pkg/config"
	"github.com/spf13/cobra"
	"os"
)

func NewRootCmd() *cobra.Command {
	var cfgFile string
	var level string
	var jsonLog bool
	rootCmd := &cobra.Command{
		Short: config.AppName,
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, config.ConfigFlagName, "", fmt.Sprintf("config file (default is $HOME/.%s/config.yaml)", config.AppName))
	rootCmd.PersistentFlags().StringVarP(&level, config.LevelFlagName, "", "info", "The logging level.")
	rootCmd.PersistentFlags().BoolVarP(&jsonLog, "json-logs", "", false, "Enable json logging.")

	rootCmd.AddCommand(NewVersionCmd(os.Stdout))
	rootCmd.AddCommand(NewConfigCmd())
	//rootCmd.AddCommand(NewRunCmd())

	return rootCmd
}
