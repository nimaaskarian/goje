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

	"github.com/fsnotify/fsnotify"
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
	Timer         timer.TimerConfig
	CustomCss     string        `mapstructure:"custom-css"`
	Activitywatch bool          `mapstructure:"activitywatch"`
	NoWebgui      bool          `mapstructure:"no-webgui"`
	NoOpenBrowser bool          `mapstructure:"no-open-browser"`
	ExecStart     string        `mapstructure:"exec-start"`
	ExecEnd       string        `mapstructure:"exec-end"`
	ExecPause     string        `mapstructure:"exec-pause"`
	HttpAddress   string        `mapstructure:"http-address"`
	TcpAddress    string        `mapstructure:"tcp-address"`
	Fifo          string        `mapstructure:"fifo"`
	Loglevel      string        `mapstructure:"loglevel"`
	http_deamon   *httpd.Daemon `mapstructure:"-"`
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
	rootCmd.Flags().DurationSliceP("duration", "D", timer.DefaultConfig.Duration[:], "duration of timer as pomodoro,short break,long break")
	rootCmd.Flags().BoolP("not-paused", "P", false, "timer is not paused by default")
	rootCmd.Flags().UintP("sessions", "s", timer.DefaultConfig.Sessions, "count of sessions in timer")
	rootCmd.Flags().BoolP("paused", "p", false, "timer is paused by default")
	rootCmd.MarkFlagsMutuallyExclusive("paused", "not-paused")
	rootCmd.Flags().DurationP("duration-per-tick", "d", time.Second, "duration per each tick, that determines the accuracy of timer")
	rootCmd.Flags().String("loglevel", "error", "log level of goje")
	rootCmd.RegisterFlagCompletionFunc("loglevel", func(cmd *cobra.Command, args []string, to_complete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []string{"error", "warn", "debug", "info"}, cobra.ShellCompDirectiveNoFileComp
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (errout error) {
		if err := readConfig(cmd); err != nil {
			return err
		}
		viper.OnConfigChange(func(e fsnotify.Event) {
			slog.Info("config changed", "path", e.Name)
			if err := readConfig(cmd); err != nil {
				errout = err
			}
		})
		viper.WatchConfig()
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

func readConfig(cmd *cobra.Command) error {
	if config_file != "" {
		// if the config_file arg is passed and doesn't exist
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
	viper.SetEnvPrefix("goje")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	if err := viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	if err := viper.Unmarshal(&config); err != nil {
		return err
	}
	timer_viper := viper.Sub("timer")
	timer_viper.SetEnvPrefix("goje")
	timer_viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	timer_viper.AutomaticEnv()
	if err := timer_viper.BindPFlags(cmd.Flags()); err != nil {
		return err
	}
	if err := timer_viper.Unmarshal(&config.Timer); err != nil {
		return err
	}
	if ok, err := cmd.Flags().GetBool("not-paused"); ok && err == nil {
		config.Timer.Paused = false
	}
	if err := loglevel.Set(config.Loglevel); err != nil {
		return err
	}
	slog.SetLogLoggerLevel(slog.Level(loglevel))
	slog.Info("using configuration", "path", viper.ConfigFileUsed())
	return nil
}

func runDaemons(t *timer.Timer) (errout error) {
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
	if config.Fifo != "" {
		utils.Mkfifo(config.Fifo)
		writeToFile := func(t *timer.Timer) {
			content, _ := json.Marshal(t)
			go func() {
				if err := os.WriteFile(config.Fifo, append(content, '\n'), 0644); err != nil {
					errout = err
				}
			}()
			slog.Debug("writing to fifo finished")
		}
		slog.Debug("setting up fifo events")
		config.Timer.OnQuit.Append(func(*timer.Timer) {
			slog.Debug("removing fifo", "path", config.Fifo)
			os.Remove(config.Fifo)
		})
		config.Timer.OnChange.Append(writeToFile)
		config.Timer.OnModeEnd.Append(writeToFile)
		config.Timer.OnModeStart.Append(writeToFile)
	}
	if config.Activitywatch {
		aw := activitywatch.Watcher{}
		aw.Init()
		aw.AddEventWatchers(&config.Timer)
	}

	t.Config = &config.Timer

	if config.TcpAddress != "" {
		tcp_daemon := tcpd.Daemon{
			Timer: t,
		}
		if err := tcp_daemon.InitializeListener(config.TcpAddress); err != nil {
			return err
		}
		go tcp_daemon.Run()
	}
	if config.HttpAddress != "" {
		config.http_deamon = &httpd.Daemon{
			Timer:   t,
			Clients: &sync.Map{},
		}
		config.Timer.OnChange.Append(config.http_deamon.UpdateAllChangeEvent)
		config.http_deamon.SetupEndStartEvents()
		config.http_deamon.Init()
		config.http_deamon.JsonRoutes()
		if !config.NoWebgui {
			go runWebgui(config.HttpAddress)
		}
		go func() {
			errout = config.http_deamon.Run(config.HttpAddress)
		}()
	}
	t.Init()
	go t.Loop()
	return nil
}

func runWebgui(address string) {
	config.http_deamon.WebguiRoutes(config.CustomCss)
	if !config.NoOpenBrowser {
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
