package cmd

import (
	"fmt"
	"os"

	"github.com/nimaaskarian/goje/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var config_file string
func init() {
  rootCmd.PersistentFlags().StringVarP(&config_file, "config", "c", "", "path to config file. uses default if not specified")
}

var rootCmd = &cobra.Command{
	Use:   "goje",
	Short: "a server-client pomodoro timer",
	Long:  "goje is a mpd-like, client server pomodoro timer with modern features, suitable for both everyday users and computer nerds",
  PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    if config_file != "" {
      viper.SetConfigFile(config_file)
    } else {
      viper.SetConfigName("config")
      viper.SetConfigType("toml")
      viper.AddConfigPath(utils.ConfigDir())
    }
    return viper.ReadInConfig()
  },
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
