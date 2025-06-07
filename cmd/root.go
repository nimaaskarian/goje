package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

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
	Timer       timer.TimerConfig
	http_deamon *httpd.Daemon
}

var config = AppConfig{Timer: timer.DefaultConfig}

func init() {
	rootCmd.PersistentFlags().StringVarP(&config_file, "config", "c", "", "path to config file. uses default if not specified")
	for mode := range timer.MODE_MAX {
		lower := strings.ToLower(mode.String())
		worm_case := strings.ReplaceAll(lower, " ", "-")
		rootCmd.Flags().DurationVarP(
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
	rootCmd.Flags().Bool("no-open-browser", false, "don't open the browser when running webgui")
	rootCmd.Flags().BoolP("activitywatch", "", false, "daemon send's pomodoro data to activitywatch if is present")
	rootCmd.Flags().StringP("write-file", "w", "", "write timer events in a file at given path")
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
	SilenceUsage:  true,
  PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
    if err := readConfig(); err != nil {
      return err
    }
    viper.SetEnvPrefix("goje")
    viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
    viper.AutomaticEnv()
    return viper.BindPFlags(cmd.Flags())
  },
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runDaemons(); err != nil {
			return err
		}
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGHUP)
    for {
      <-sig
      config.http_deamon.UpdateClients(SigEvent{
        name: "restart",
      })
    }
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
	if err := viper.ReadInConfig(); err != nil {
    _, ok := err.(viper.ConfigFileNotFoundError)
    if !ok {
      return err
    }
	}
	return nil
}

func runDaemons() (errout error) {
	if path := viper.GetString("exec-start"); path != "" {
		config.Timer.OnModeStart.Append(func(t *timer.Timer) {
			content, _ := json.Marshal(t)
			if err := exec.Command(path, string(content)).Run(); err != nil {
				errout = err
			}
		})
	}
	if path := viper.GetString("exec-end"); path != "" {
		config.Timer.OnModeEnd.Append(func(t *timer.Timer) {
			content, _ := json.Marshal(t)
			exec.Command(path, string(content)).Run()
		})
	}
	if path := viper.GetString("exec-pause"); path != "" {
		config.Timer.OnPause.Append(func(t *timer.Timer) {
			content, _ := json.Marshal(t)
			exec.Command(path, string(content)).Run()
		})
	}
	if path := viper.GetString("write-file"); path != "" {
		writeChanges := func(event_callback func(payload any) timer.Event) func(t *timer.Timer) {
			return func(t *timer.Timer) {
				content, _ := json.Marshal(event_callback(t))
				errout = os.WriteFile(path, append(content, '\n'), 0644)
			}
		}
		config.Timer.OnChange.Append(writeChanges(timer.OnChangeEvent))
		config.Timer.OnModeEnd.Append(writeChanges(timer.OnModeEndEvent))
		config.Timer.OnModeStart.Append(writeChanges(timer.OnModeStartEvent))
	}
	if viper.GetBool("activitywatch") {
		aw := activitywatch.Watcher{}
		aw.Init()
		aw.AddEventWatchers(&config.Timer)
	}
	t := timer.Timer{}
  config.Timer.Duration = [...]time.Duration{
    viper.GetDuration("pomodoro-duration"),
    viper.GetDuration("short-break-duration"),
    viper.GetDuration("long-break-duration"),
  }
  config.Timer.Paused = viper.GetBool("paused");

	t.Config = &config.Timer

	if address := viper.GetString("tcp-address"); address != "" {
		tcp_daemon := tcpd.Daemon{
			Timer:    &t,
		}
		if err := tcp_daemon.InitializeListener(address); err != nil {
			return err
		}
		go tcp_daemon.Run()
	}
	if address := viper.GetString("http-address"); address != "" {
		config.http_deamon = &httpd.Daemon{
			Timer:   &t,
			Clients: &sync.Map{},
		}
		config.Timer.OnChange.Append(config.http_deamon.UpdateAllChangeEvent)
		config.http_deamon.SetupEndStartEvents()
		config.http_deamon.Init()
		config.http_deamon.JsonRoutes()
		if !viper.GetBool("no-webgui") {
			go runWebgui(address)
		}
		go func() {
			errout = config.http_deamon.Run(address)
		}()
	}
	t.Init()
	go t.Loop()
	return nil
}

func runWebgui(address string) {
	config.http_deamon.WebguiRoutes(viper.GetString("custom-css"))
	if !viper.GetBool("no-open-browser") {
		if strings.HasPrefix(address, "http://") {
			utils.OpenURL(address)
		} else {
			if address[0] == ':' {
				utils.OpenURL("http://localhost" + address)
			} else {
				utils.OpenURL("http://" + address)
			}
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
    os.Exit(1)
	}
}
