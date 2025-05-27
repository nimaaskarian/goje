package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/nimaaskarian/tom/json"
	"github.com/nimaaskarian/tom/tcp"
	"github.com/nimaaskarian/tom/timer"
	"github.com/spf13/cobra"
)

// server flags
var (
	tcp_address  string
	json_address string
	buffsize     uint
	print        bool
)

var config = timer.DefaultConfig

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.PersistentFlags().DurationVarP(
		&config.Duration[timer.Pomodoro],
		"pomodoro-duration",
		"p",
		timer.DefaultConfig.Duration[timer.Pomodoro],
		"duration of pomodoro sections of the timer",
	)
	daemonCmd.PersistentFlags().DurationVarP(
		&config.Duration[timer.ShortBreak],
		"short-break-duration",
		"s",
		timer.DefaultConfig.Duration[timer.ShortBreak],
		"duration of short break sections of the timer",
	)
	daemonCmd.PersistentFlags().DurationVarP(
		&config.Duration[timer.LongBreak],
		"long-break-duration",
		"l",
		timer.DefaultConfig.Duration[timer.LongBreak],
		"duration of long break sections of the timer",
	)
	daemonCmd.PersistentFlags().StringVarP(&tcp_address, "tcp-address", "a", ":8088", "address:[port] for tcp pomodoro daemon")
	daemonCmd.PersistentFlags().StringVarP(&json_address, "json-address", "j", "", "address:[port] for http json pomodoro daemon (doesn't run when empty)")
	daemonCmd.PersistentFlags().UintVar(&buffsize, "buff-size", 1024, "size of buffer that messages are parsed with")
	daemonCmd.PersistentFlags().BoolVar(&print, "print", false, "the daemon prints current duration to stderr on ticks when this option is present")
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
	RunE: func(cmd *cobra.Command, args []string) error {
    if print {
      config.AfterTick = func(timer *timer.Timer) {
        fmt.Fprintln(os.Stderr, timer)
      }
      config.AfterSeek = config.AfterTick
    }
		tomato := timer.Timer{}
		tomato.SetConfig(config)
		tomato.Init()
		go tomato.Loop()
		tcp_daemon := tcp.Daemon{
			Timer:    &tomato,
			Buffsize: buffsize,
		}
		if err := tcp_daemon.InitializeListener(tcp_address); err != nil {
			return err
		}
		if json_address != "" {
			json_deamon := json.Daemon{
				Timer: &tomato,
			}
			json_deamon.Init()
			go func() {
				if err := json_deamon.Run(json_address); err != nil {
					log.Fatalln(err)
				}
			}()
		}
		tcp_daemon.Run()
		return nil
	},
}
