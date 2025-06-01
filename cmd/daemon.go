package cmd

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

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
	NoWebgui      bool   `mapstructure:"no-webgui"`
	NoOpenBrowser bool   `mapstructure:"no-open-browser"`
	WriteFile     string `mapstructure:"write-path"`
	Activitywatch bool
	ExecEnd       string `mapstructure:"exec-end"`
	ExecStart     string `mapstructure:"exec-start"`
	Timer         timer.TimerConfig
}

var config = DaemonConfig{Timer: timer.DefaultConfig}

func init() {
	rootCmd.AddCommand(daemonCmd)
	for mode := range timer.MODE_MAX {
		lower := strings.ToLower(mode.String())
		worm_case := strings.ReplaceAll(lower, " ", "-")
		daemonCmd.PersistentFlags().DurationVarP(
			&config.Timer.Duration[mode],
			worm_case+"-duration",
			worm_case[0:1],
			timer.DefaultConfig.Duration[mode],
			"duration of "+lower+" sections of the timer",
		)
	}
	daemonCmd.PersistentFlags().StringVar(&config.ExecEnd, "exec-start", "", "script to run when any timer mode starts (run's the script with json of timer as the first arguemnt)")
	daemonCmd.PersistentFlags().StringVar(&config.ExecStart, "exec-end", "", "script to run when any timer mode ends (run's the script with json of timer as the first arguemnt)")
	viper.SetDefault("tcp-address", "localhost:7800")
	viper.SetDefault("http-address", "localhost:7900")
	viper.SetDefault("buff-size", 1024)
	daemonCmd.PersistentFlags().StringVarP(&config.TcpAddress, "tcp-address", "a", "", "address:[port] for tcp pomodoro daemon (doesn't run when empty)")
	daemonCmd.PersistentFlags().StringVarP(&config.HttpAddress, "http-address", "A", "", "address:[port] for http pomodoro api (doesn't run when empty)")
	daemonCmd.PersistentFlags().BoolVar(&config.NoWebgui, "no-webgui", false, "don't run webgui. webgui can't be run without the json server")
	daemonCmd.PersistentFlags().UintVar(&config.BuffSize, "buff-size", 0, "size of buffer that tcp messages are parsed with")
	daemonCmd.PersistentFlags().BoolVar(&config.NoOpenBrowser, "no-open-browser", false, "don't open the browser when running webgui")
	daemonCmd.PersistentFlags().BoolVarP(&config.Activitywatch, "activitywatch", "", false, "daemon send's pomodoro data to activitywatch if is present")
	daemonCmd.PersistentFlags().StringVar(&config.WriteFile, "write-file", "", "write timer events in a file at given path")
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
	RunE: func(cmd *cobra.Command, args []string) (errout error) {
		if config.ExecStart != "" {
			config.Timer.OnModeStart = append(config.Timer.OnModeStart, func(t *timer.Timer) {
				content, _ := json.Marshal(t)
				if err := exec.Command(config.ExecStart, string(content)).Run(); err != nil {
					log.Fatalln(err)
				}
			})
		}
		if config.ExecEnd != "" {
			config.Timer.OnModeEnd = append(config.Timer.OnModeEnd, func(t *timer.Timer) {
				content, _ := json.Marshal(t)
				exec.Command(config.ExecStart, string(content)).Run()
			})
		}
		if config.WriteFile != "" {
			writeChanges := func(event_callback func(payload any) timer.Event) func(t *timer.Timer) {
				return func(t *timer.Timer) {
					content, _ := json.Marshal(event_callback(t))
					errout = os.WriteFile(config.WriteFile, append(content, '\n'), 0644)
				}
			}
			config.Timer.OnChange = append(config.Timer.OnChange, writeChanges(timer.OnChangeEvent))
			config.Timer.OnModeEnd = append(config.Timer.OnChange, writeChanges(timer.OnModeEndEvent))
			config.Timer.OnModeStart = append(config.Timer.OnChange, writeChanges(timer.OnModeStartEvent))
		}
		if config.Activitywatch {
			activitywatch.SetupTimerConfig(&config.Timer)
		}
		tomato := timer.Timer{}
		tomato.SetConfig(&config.Timer)

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
			http_deamon := httpd.Daemon{
				Timer:   &tomato,
				Clients: &sync.Map{},
			}
			config.Timer.OnChange = append(config.Timer.OnChange, http_deamon.UpdateAllChangeEvent)
			http_deamon.SetupEndStartEvents()
			http_deamon.Init()
			http_deamon.JsonRoutes()
			if !config.NoWebgui {
				go runWebgui(&http_deamon)
			}
			go func() {
				errout = http_deamon.Run(config.HttpAddress)
			}()
		}
		tomato.Init()
		go tomato.Loop()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for {
			<-sig
			if err := rootCmd.PersistentPreRunE(rootCmd, os.Args[1:]); err != nil {
				return err
			}
			viper.Unmarshal(&config)
		}
	},
}

func runWebgui(http_deamon *httpd.Daemon) {
	http_deamon.WebguiRoutes()
	if !config.NoOpenBrowser {
		if strings.HasPrefix(config.HttpAddress, "http://") {
			utils.OpenURL(config.HttpAddress)
		} else {
			if config.HttpAddress[0] == ':' {
				utils.OpenURL("http://localhost" + config.HttpAddress)
			} else {
				utils.OpenURL("http://" + config.HttpAddress)
			}
		}
	}
}
