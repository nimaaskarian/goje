package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:   "goje",
	Short: "a server-client pomodoro timer",
	Long:  "goje is a mpd-like, client server pomodoro timer with modern features",
}

func Execute() {
	cmd, _, err := rootCmd.Find(os.Args[1:])
	if err == nil && cmd.Use == rootCmd.Use && cmd.Flags().Parse(os.Args[1:]) != pflag.ErrHelp {
		args := append([]string{daemonCmd.Use}, os.Args[1:]...)
		rootCmd.SetArgs(args)
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
