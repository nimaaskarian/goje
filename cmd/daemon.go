package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/nimaaskarian/tom/activitywatch"
	"github.com/nimaaskarian/tom/httpd"
	"github.com/nimaaskarian/tom/tcpd"
	"github.com/nimaaskarian/tom/timer"
	"github.com/spf13/cobra"
)

// server flags
var (
	tcp_address       string
	json_address      string
	buffsize          uint
	should_print      bool
	no_webgui      bool
	write_path        string
	run_activitywatch bool
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
	daemonCmd.PersistentFlags().BoolVar(&no_webgui, "no-webgui", false, "don't run webgui. webgui can't be run without the json server")
	daemonCmd.PersistentFlags().UintVar(&buffsize, "buff-size", 1024, "size of buffer that messages are parsed with")
	daemonCmd.PersistentFlags().BoolVar(&should_print, "print", false, "the daemon prints current duration to stderr on ticks when this option is present")
	daemonCmd.PersistentFlags().BoolVarP(&run_activitywatch, "activitywatch", "w", false, "activitywatch port. doesn't send pomodoro data to activitywatch if is empty")
	daemonCmd.PersistentFlags().StringVar(&write_path, "write-file", "", "write last timer in a file at given path. empty means no write")
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
		config.AfterTick = afterTick
		config.AfterSeek = config.AfterTick
		if run_activitywatch {
			activitywatch.SetupTimerConfig(&config)
		}
		tomato := timer.Timer{}
		if tcp_address != "" {
			tcp_daemon := tcpd.Daemon{
				Timer:    &tomato,
				Buffsize: buffsize,
			}
			if err := tcp_daemon.InitializeListener(tcp_address); err != nil {
				return err
			}
			go tcp_daemon.Run()
		}
		if json_address != "" {
			timerchan := make(chan string)
			config.AfterTick = func(t *timer.Timer) {
				afterTick(t)
				bytes, _ := json.Marshal(t)
				timerchan <- string(bytes)
			}
			config.AfterSeek = config.AfterTick
			json_deamon := httpd.Daemon{
				Timer: &tomato,
        TimerJsonChan: timerchan,
			}
			json_deamon.Init()
      json_deamon.JsonRoutes()
      if !no_webgui {
        json_deamon.WebguiRoutes()
      }
			go func() {
				if err := json_deamon.Run(json_address); err != nil {
					log.Fatalln(err)
				}
			}()
		}
		tomato.SetConfig(config)
		tomato.Init()
		tomato.Loop()
		return nil
	},
}

func afterTick(timer *timer.Timer) {
	if should_print {
		fmt.Fprintln(os.Stderr, timer)
	}
	if write_path != "" {
		if err := os.WriteFile(write_path, []byte(timer.String()), 0644); err != nil {
			log.Fatalln(err)
		}
	}
}
