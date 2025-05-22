package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/nimaaskarian/tom/requests"
	"github.com/nimaaskarian/tom/timer"
	"github.com/spf13/cobra"
)

// server flags
var (
	address  string
	buffsize int
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
		timer.DefaultConfig.Duration[timer.LongBreak],
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
	daemonCmd.PersistentFlags().IntVar(&buffsize, "buff-size", 1024, "size of buffer that messages are parsed with")
}

var daemonCmd = &cobra.Command{
	Aliases: []string{
		"server",
		"d",
		"s",
	},
	Use:   "daemon",
	Short: "run a tom daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		ln, err := net.Listen("tcp", address)
		if err != nil {
			return err
		}
		tomato.Config.OnTick = func(*timer.Timer) {
			fmt.Println(tomato.String())
		}
		tomato.Config.OnSeek = tomato.Config.OnTick
		tomato.Init()
		go tomato.Loop()
		buff := make([]byte, buffsize)
		for {
			conn, err := ln.Accept()
			if err != nil {
				slog.Warn("connection throw error", "err", err)
				continue
			}
			conn.Write([]byte("OK tom 0.0.1\n"))
			for {
				n, err := conn.Read(buff)
				if err == io.EOF {
					break
				} else if err != nil {
					slog.Warn("read throw error", "err", err)
					continue
				}
				cmd, out, err := requests.ParseInput(&tomato, string(bytes.TrimSpace(buff[:n])))
				if err != nil {
					slog.Error("command throw error", "err", err)
					conn.Write(fmt.Appendf(nil, "ACK {%s} %s\n", cmd, err))
				} else {
					conn.Write([]byte(out))
					conn.Write([]byte("OK\n"))
				}
			}
		}
	},
}
