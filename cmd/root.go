package cmd

import (
	"encoding/json"
	"errors"
	"log/slog"
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

type LogLevel slog.Level

func (level *LogLevel) Set(v string) error {
	switch v {
	case "warning", "warn":
		*level = LogLevel(slog.LevelWarn)
	case "info":
		*level = LogLevel(slog.LevelInfo)
	case "debug":
		*level = LogLevel(slog.LevelDebug)
	case "error", "":
		*level = LogLevel(slog.LevelError)
	default:
		return errors.New(`loglevel must be one of "info", "warn", "debug, or "error"`)
	}
	return nil
}

var config = AppConfig{Timer: timer.DefaultConfig}
var loglevel LogLevel

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
	rootCmd.Flags().String("loglevel", "error", "log level of goje")
  rootCmd.RegisterFlagCompletionFunc("loglevel", func(cmd *cobra.Command, args []string, to_complete string) ([] cobra.Completion, cobra.ShellCompDirective){
    return []string{"error", "warn", "debug", "info"}, cobra.ShellCompDirectiveNoFileComp;
  })
	rootCmd.Flags().String("custom-css", "", "a custom css file to load on the website")
	rootCmd.Flags().String("exec-start", "", "command to run when any timer mode starts (run's the script with json of timer as the first arguemnt)")
	rootCmd.Flags().String("exec-end", "", "command to run when any timer mode ends (run's the script with json of timer as the first arguemnt)")
	rootCmd.Flags().String("exec-pause", "", "command to run when timer (un)pauses")
	rootCmd.Flags().StringP("tcp-address", "a", "localhost:7800", "address:[port] for tcp pomodoro daemon (doesn't run when empty)")
	rootCmd.Flags().StringP("http-address", "A", "localhost:7900", "address:[port] for http pomodoro api (doesn't run when empty)")
	rootCmd.Flags().Bool("no-webgui", false, "don't run webgui. webgui can't be run without the json server")
	rootCmd.Flags().Bool("no-open-browser", false, "don't open the browser when running webgui")
	rootCmd.Flags().BoolP("activitywatch", "", false, "daemon send's pomodoro data to activitywatch if is present")
	rootCmd.Flags().StringP("fifo", "f", "", "write timer events in a fifo at given path")
	rootCmd.Flags().BoolP("paused", "P", false, "whether the timer is initially paused or not")
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
	Use:          "goje",
	Short:        "a pomodoro timer server",
	Long:         "goje is a pomodoro timer server with modern features, suitable for both everyday users and computer nerds",
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := readConfig(); err != nil {
			return err
		}
		viper.SetEnvPrefix("goje")
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.AutomaticEnv()
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return err
		}
		if err := loglevel.Set(viper.GetString("loglevel")); err != nil {
			return err
		}
		slog.SetLogLoggerLevel(slog.Level(loglevel))
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)
		t := timer.Timer{}
		go func() {
			sig := <-sigc
			slog.Debug("caught deadly signal", "signal", sig)
			t.Config.OnQuit.Run(&t)
			slog.Debug("clean up finished. quitting")
			os.Exit(1)
		}()
		if err := runDaemons(&t); err != nil {
			return err
		}
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP)
		for {
			<-sig
			cmd.PersistentPreRun(cmd, args)
			config.http_deamon.UpdateClients(SigEvent{
				name: "restart",
			})
		}
	},
}

func readConfig() error {
	if config_file != "" {
		// if the config_file arg doesn't exist
		if _, err := os.Stat(config_file); os.IsNotExist(err) {
			return err
		}
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
		slog.Debug("default config not found. using the default values")
	}
	return nil
}

func runDaemons(t *timer.Timer) (errout error) {
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
	if path := viper.GetString("fifo"); path != "" {
    utils.Mkfifo(path)
    writeToFile := func(t *timer.Timer) {
      content, _ := json.Marshal(t)
      go func() {
        if err := os.WriteFile(path, append(content, '\n'), 0644); err != nil {
          errout = err
        }
      }()
      slog.Debug("writing to fifo finished")
    }
    slog.Debug("setting up fifo events")
    config.Timer.OnQuit.Append(func(*timer.Timer) {
      os.Remove(path)
    })
    // initially write to fifo
    writeToFile(t)
    config.Timer.OnChange.Append(writeToFile)
    config.Timer.OnModeEnd.Append(writeToFile)
    config.Timer.OnModeStart.Append(writeToFile)
	}
	if viper.GetBool("activitywatch") {
		aw := activitywatch.Watcher{}
		aw.Init()
		aw.AddEventWatchers(&config.Timer)
	}
	config.Timer.Duration = [...]time.Duration{
		viper.GetDuration("pomodoro-duration"),
		viper.GetDuration("short-break-duration"),
		viper.GetDuration("long-break-duration"),
	}
	config.Timer.Paused = viper.GetBool("paused")

	t.Config = &config.Timer

	if address := viper.GetString("tcp-address"); address != "" {
		tcp_daemon := tcpd.Daemon{
			Timer: t,
		}
		if err := tcp_daemon.InitializeListener(address); err != nil {
			return err
		}
		go tcp_daemon.Run()
	}
	if address := viper.GetString("http-address"); address != "" {
		config.http_deamon = &httpd.Daemon{
			Timer:   t,
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
