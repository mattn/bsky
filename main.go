package main

import (
	"fmt"
	"github.com/jlewi/goapp-template/cmd"
	"os"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Command failed with error: %+v", err)
		os.Exit(1)
	}
}
