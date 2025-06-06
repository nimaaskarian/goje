package cmd

import (
	"os"
	"path/filepath"

	"github.com/nimaaskarian/goje/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
  rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command {
  Use: "config",
  Short: "open config file of goje",
  RunE: func(cmd *cobra.Command, args []string) error {
    config_file := viper.ConfigFileUsed()
    if _, err := os.Stat(config_file); os.IsNotExist(err) {
      if err := viper.WriteConfig(); err != nil {
        os.MkdirAll(utils.ConfigDir(), 0755);
        config_file = filepath.Join(utils.ConfigDir(), "config.toml")
        viper.WriteConfigAs(config_file)
      }
    }
    c, err := utils.EditorCmd(config_file)
    if err != nil {
      return err
    }
    utils.CmdStdOs(c)
    return c.Run()
  },

}
