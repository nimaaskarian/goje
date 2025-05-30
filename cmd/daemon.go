package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/nimaaskarian/goje/activitywatch"
	"github.com/nimaaskarian/goje/httpd"
	"github.com/nimaaskarian/goje/tcpd"
	"github.com/nimaaskarian/goje/timer"
	"github.com/nimaaskarian/goje/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// server flags
type DaemonConfig struct {
	TcpAddress    string `mapstructure:"tcp-address"`
	HttpAddress   string `mapstructure:"http-address"`
	BuffSize      uint   `mapstructure:"buff-size"`
	Print         bool
	NoWebgui      bool   `mapstructure:"no-webgui"`
	WritePath     string `mapstructure:"write-path"`
	Activitywatch bool
	Timer   timer.TimerConfig
}

var config = DaemonConfig{Timer: timer.DefaultConfig}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.PersistentFlags().DurationVarP(
		&config.Timer.Duration[timer.Pomodoro],
		"pomodoro-duration",
		"p",
		timer.DefaultConfig.Duration[timer.Pomodoro],
		"duration of pomodoro sections of the timer",
	)
	daemonCmd.PersistentFlags().DurationVarP(
		&config.Timer.Duration[timer.ShortBreak],
		"short-break-duration",
		"s",
		timer.DefaultConfig.Duration[timer.ShortBreak],
		"duration of short break sections of the timer",
	)
	daemonCmd.PersistentFlags().DurationVarP(
		&config.Timer.Duration[timer.LongBreak],
		"long-break-duration",
		"l",
		timer.DefaultConfig.Duration[timer.LongBreak],
		"duration of long break sections of the timer",
	)
	viper.SetDefault("tcp-address", "localhost:7800")
	viper.SetDefault("http-address", "localhost:7900")
	viper.SetDefault("buff-size", 1024)
	daemonCmd.PersistentFlags().StringVarP(&config.TcpAddress, "tcp-address", "a", "", "address:[port] for tcp pomodoro daemon (doesn't run when empty)")
	daemonCmd.PersistentFlags().StringVarP(&config.HttpAddress, "http-address", "A", "", "address:[port] for http pomodoro api (doesn't run when empty)")
	daemonCmd.PersistentFlags().BoolVar(&config.NoWebgui, "no-webgui", false, "don't run webgui. webgui can't be run without the json server")
	daemonCmd.PersistentFlags().UintVar(&config.BuffSize, "buff-size", 0, "size of buffer that messages are parsed with")
	daemonCmd.PersistentFlags().BoolVar(&config.Print, "print", false, "the daemon prints current duration to stderr on ticks when this option is present")
	daemonCmd.PersistentFlags().BoolVarP(&config.Activitywatch, "activitywatch", "w", false, "daemon send's pomodoro data to activitywatch if is present")
	daemonCmd.PersistentFlags().StringVar(&config.WritePath, "write-file", "", "write last timer in a file at given path. empty means no write")
	viper.BindPFlags(daemonCmd.PersistentFlags())
}

var daemonCmd = &cobra.Command{
	Aliases: []string{
		"daemon",
		"server",
		"d",
		"s",
	},
	Use:   "daemon",
	Short: "run a pomodoro tcp daemon",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return viper.Unmarshal(&config)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		config.Timer.OnChange = afterTick
		if config.Activitywatch {
			activitywatch.SetupTimerConfig(&config.Timer)
		}
		tomato := timer.Timer{}
		if config.TcpAddress != "" {
			tcp_daemon := tcpd.Daemon{
				Timer:    &tomato,
				Buffsize: config.BuffSize,
			}
			if err := tcp_daemon.InitializeListener(config.TcpAddress); err != nil {
				return err
			}
			go tcp_daemon.Run()
		}
		if config.HttpAddress != "" {
			json_deamon := httpd.Daemon{
				Timer:   &tomato,
				Clients: &sync.Map{},
			}
			config.Timer.OnChange = func(t *timer.Timer) {
				afterTick(t)
				json_deamon.UpdateClients(json_deamon.TimerEvent())
			}
			json_deamon.Init()
			json_deamon.JsonRoutes()
			if !config.NoWebgui {
				if strings.HasPrefix(config.HttpAddress, "http://") {
					utils.OpenURL(config.HttpAddress)
				} else {
					if config.HttpAddress[0] == ':' {
						utils.OpenURL("http://localhost" + config.HttpAddress)
					} else {
						utils.OpenURL("http://" + config.HttpAddress)
					}
				}
				json_deamon.WebguiRoutes()
			}
			go func() {
				if err := json_deamon.Run(config.HttpAddress); err != nil {
					log.Fatalln(err)
				}
			}()
		}
		tomato.SetConfig(config.Timer)
		tomato.Init()
		tomato.Loop()
		return nil
	},
}

func afterTick(timer *timer.Timer) {
	if config.Print {
		fmt.Fprintln(os.Stderr, timer)
	}
	if config.WritePath != "" {
		if err := os.WriteFile(config.WritePath, []byte(timer.String()), 0644); err != nil {
			log.Fatalln(err)
		}
	}
}
