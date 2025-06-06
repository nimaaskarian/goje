package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
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

var config_file string

type AppConfig struct {
	TcpAddress    string `mapstructure:"tcp-address"`
	HttpAddress   string `mapstructure:"http-address"`
	BuffSize      uint   `mapstructure:"buff-size"`
	NoWebgui      bool   `mapstructure:"no-webgui"`
	NoOpenBrowser bool   `mapstructure:"no-open-browser"`
	WriteFile     string `mapstructure:"write-path"`
	Activitywatch bool
	ExecEnd       string `mapstructure:"exec-end"`
	ExecStart     string `mapstructure:"exec-start"`
	ExecPause     string `mapstructure:"exec-pause"`
	CustomCss     string `mapstructure:"custom-css"`
	Timer         timer.TimerConfig
	http_deamon   *httpd.Daemon
}

var config = AppConfig{Timer: timer.DefaultConfig}

func init() {
	rootCmd.PersistentFlags().StringVarP(&config_file, "config", "c", "", "path to config file. uses default if not specified")
	for mode := range timer.MODE_MAX {
		lower := strings.ToLower(mode.String())
		worm_case := strings.ReplaceAll(lower, " ", "-")
		rootCmd.PersistentFlags().DurationVarP(
			&config.Timer.Duration[mode],
			worm_case+"-duration",
			worm_case[0:1],
			timer.DefaultConfig.Duration[mode],
			"duration of "+lower+" sections of the timer",
		)
	}
	rootCmd.Flags().String("custom-css", "", "a custom css file to load on the website")
	rootCmd.Flags().String("exec-start", "", "command to run when any timer mode starts (run's the script with json of timer as the first arguemnt)")
	rootCmd.Flags().String("exec-end", "", "command to run when any timer mode ends (run's the script with json of timer as the first arguemnt)")
	rootCmd.Flags().String("exec-pause", "", "command to run when timer (un)pauses")
	rootCmd.Flags().StringP("tcp-address", "a", "localhost:7800", "address:[port] for tcp pomodoro daemon (doesn't run when empty)")
	rootCmd.Flags().StringP("http-address", "A", "localhost:7900", "address:[port] for http pomodoro api (doesn't run when empty)")
	rootCmd.Flags().Bool("no-webgui", false, "don't run webgui. webgui can't be run without the json server")
	rootCmd.Flags().Uint("buff-size", 1024, "size of buffer that tcp messages are parsed with")
	rootCmd.Flags().Bool("no-open-browser", false, "don't open the browser when running webgui")
	rootCmd.Flags().BoolP("activitywatch", "", false, "daemon send's pomodoro data to activitywatch if is present")
	rootCmd.Flags().StringP("write-file", "w", "", "write timer events in a file at given path")
	readConfig()
	viper.BindPFlags(rootCmd.Flags())
	viper.Unmarshal(&config)
}

type SigEvent struct {
	name string
}

func (e SigEvent) Payload() any {
	return nil
}
func (e SigEvent) Name() string {
	return e.name
}

var rootCmd = &cobra.Command{
	Use:           "goje",
	Short:         "a pomodoro timer server",
	Long:          "goje is a pomodoro timer server with modern features, suitable for both everyday users and computer nerds",
	SilenceErrors: true,
	SilenceUsage:  true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runDaemons(); err != nil {
			return err
		}
    return nil
	},
}

func readConfig() error {
	if config_file != "" {
		viper.SetConfigFile(config_file)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		viper.AddConfigPath(utils.ConfigDir())
	}
	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, viper.ConfigFileNotFoundError{}) {
		return err
	}
  return nil
}

func runDaemons() (errout error) {
	if config.ExecStart != "" {
		config.Timer.OnModeStart.Append(func(t *timer.Timer) {
			content, _ := json.Marshal(t)
			if err := exec.Command(config.ExecStart, string(content)).Run(); err != nil {
				errout = err
			}
		})
	}
	if config.ExecEnd != "" {
		config.Timer.OnModeEnd.Append(func(t *timer.Timer) {
			content, _ := json.Marshal(t)
			exec.Command(config.ExecEnd, string(content)).Run()
		})
	}
	if config.ExecPause != "" {
		config.Timer.OnPause.Append(func(t *timer.Timer) {
			content, _ := json.Marshal(t)
			exec.Command(config.ExecPause, string(content)).Run()
		})
	}
	if config.WriteFile != "" {
		writeChanges := func(event_callback func(payload any) timer.Event) func(t *timer.Timer) {
			return func(t *timer.Timer) {
				content, _ := json.Marshal(event_callback(t))
				errout = os.WriteFile(config.WriteFile, append(content, '\n'), 0644)
			}
		}
		config.Timer.OnChange.Append(writeChanges(timer.OnChangeEvent))
		config.Timer.OnModeEnd.Append(writeChanges(timer.OnModeEndEvent))
		config.Timer.OnModeStart.Append(writeChanges(timer.OnModeStartEvent))
	}
	if config.Activitywatch {
		aw := activitywatch.Watcher{}
		aw.Init()
		aw.AddEventWatchers(&config.Timer)
	}
	t := timer.Timer{}
	t.SetConfig(&config.Timer)

	if config.TcpAddress != "" {
		tcp_daemon := tcpd.Daemon{
			Timer:    &t,
			Buffsize: config.BuffSize,
		}
		if err := tcp_daemon.InitializeListener(config.TcpAddress); err != nil {
			return err
		}
		go tcp_daemon.Run()
	}
	if config.HttpAddress != "" {
		config.http_deamon = &httpd.Daemon{
			Timer:   &t,
			Clients: &sync.Map{},
		}
		config.Timer.OnChange.Append(config.http_deamon.UpdateAllChangeEvent)
		config.http_deamon.SetupEndStartEvents()
		config.http_deamon.Init()
		config.http_deamon.JsonRoutes()
		if !config.NoWebgui {
			go runWebgui()
		}
		go func() {
			errout = config.http_deamon.Run(config.HttpAddress)
		}()
	}
	t.Init()
	go t.Loop()
	return nil
}

func runWebgui() {
	config.http_deamon.WebguiRoutes(config.CustomCss)
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

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
  sig := make(chan os.Signal, 1)
  signal.Notify(sig, syscall.SIGHUP)
  for {
    <-sig
    readConfig()
    viper.BindPFlags(rootCmd.Flags())
    viper.Unmarshal(&config)
    config.http_deamon.UpdateClients(SigEvent{
      name: "restart",
    })
  }
}
