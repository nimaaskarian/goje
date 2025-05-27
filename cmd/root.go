package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tom",
	Short: "tom is a server-client pomodoro timer",
	Long: "tom is a mpd-like, client server pomodoro timer with modern features",
}

func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}
