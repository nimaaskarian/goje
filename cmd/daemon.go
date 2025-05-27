package cmd

import (
	"fmt"

	"github.com/nimaaskarian/tom/tcp"
	"github.com/nimaaskarian/tom/timer"
	"github.com/spf13/cobra"
)

// server flags
var (
	address  string
	buffsize uint
)

var tomato = timer.Timer{
	Config: timer.DefaultConfig,
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.PersistentFlags().DurationVarP(
		&tomato.Config.Duration[timer.Pomodoro],
		"pomodoro-duration",
		"p",
		timer.DefaultConfig.Duration[timer.Pomodoro],
		"duration of pomodoro sections of the timer",
	)
	daemonCmd.PersistentFlags().DurationVarP(
		&tomato.Config.Duration[timer.ShortBreak],
		"short-break-duration",
		"s",
		timer.DefaultConfig.Duration[timer.ShortBreak],
		"duration of short break sections of the timer",
	)
	daemonCmd.PersistentFlags().DurationVarP(
		&tomato.Config.Duration[timer.LongBreak],
		"long-break-duration",
		"l",
		timer.DefaultConfig.Duration[timer.LongBreak],
		"duration of long break sections of the timer",
	)
	daemonCmd.PersistentFlags().StringVarP(&address, "address", "a", ":8088", "address:[port] for the daemon to listen to")
	daemonCmd.PersistentFlags().UintVar(&buffsize, "buff-size", 1024, "size of buffer that messages are parsed with")
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
		tomato.Config.OnTick = func(*timer.Timer) {
			fmt.Println(tomato.String())
		}
		tomato.Config.OnSeek = tomato.Config.OnTick
		tomato.Init()
		go tomato.Loop()
		tcpln, err := tcp.NetListener(address)
		if err != nil {
			return err
		}
    tcp.RunTcpDaemon(&tomato, tcpln, buffsize)
		return nil
	},
}
