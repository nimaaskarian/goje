package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
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
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var config_file string

// global flag of quitting. turns on when a quitting precedure starts
var quitting = false

type AppConfig struct {
	Timer                timer.TimerConfig
	CustomCss            string        `mapstructure:"custom-css,omitempty"`
	Activitywatch        bool          `mapstructure:"activitywatch,omitempty"`
	NoWebgui             bool          `mapstructure:"no-webgui,omitempty"`
	NoOpenBrowser        bool          `mapstructure:"no-open-browser,omitempty"`
	ExecStart            string        `mapstructure:"exec-start,omitempty"`
	ExecEnd              string        `mapstructure:"exec-end,omitempty"`
	ExecPause            string        `mapstructure:"exec-pause,omitempty"`
	HttpAddress          string        `mapstructure:"http-address,omitempty"`
	TcpAddress           string        `mapstructure:"tcp-address,omitempty"`
	Fifo                 string        `mapstructure:"fifo,omitempty"`
	Loglevel             string        `mapstructure:"loglevel,omitempty"`
	Certfile             string        `mapstructure:"certfile,omitempty"`
	Keyfile              string        `mapstructure:"keyfile,omitempty"`
	Statefile            string        `mapstructure:"statefile,omitempty"`
  NtfyAddress          string        `mapstructure:"ntfy-address,omitempty"`
	StatefileKeepUpdated bool          `mapstructure:"statefile-keep-updated,omitempty"`
	Version              bool          `mapstructure:"version,omitempty"`
	Help                 bool          `mapstructure:"help,omitempty"`
	httpDaemon           *httpd.Daemon `mapstructure:"-"`
}

var ctx context.Context
var cancel context.CancelFunc

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

// flags that are shared between client and root
func rootFlags() *pflag.FlagSet {
	flagset := pflag.NewFlagSet("roots", pflag.ExitOnError)
	flagset.DurationSliceP("duration", "D", timer.DefaultConfig.Duration[:], "duration of timer as pomodoro,short break,long break")
	flagset.BoolP("not-paused", "P", false, "timer is not paused by default")
	flagset.UintP("sessions", "s", timer.DefaultConfig.Sessions, "count of sessions in timer")
	flagset.BoolP("paused", "p", false, "timer is paused by default")
	flagset.DurationP("duration-per-tick", "d", time.Second, "duration per each tick, that determines the accuracy of timer")
	flagset.String("custom-css", "", "a custom css file to load on the website")
	flagset.String("exec-start", "", "command to run when any timer mode starts (run's the script with json of timer as the first arguemnt)")
	flagset.String("exec-end", "", "command to run when any timer mode ends (run's the script with json of timer as the first arguemnt)")
	flagset.String("exec-pause", "", "command to run when timer (un)pauses")
	flagset.StringP("tcp-address", "a", "localhost:7800", "address:[port] for tcp pomodoro daemon (doesn't run when empty)")
	flagset.StringP("http-address", "A", "localhost:7900", "address:[port] for http pomodoro api (doesn't run when empty)")
	flagset.Bool("no-webgui", false, "don't run webgui. webgui can't be run without the json server")
	flagset.Bool("no-open-browser", false, "don't open the browser when running webgui")
	flagset.BoolP("activitywatch", "", false, "daemon send's pomodoro data to activitywatch if is present")
	flagset.StringP("fifo", "f", "", "write timer events in a fifo at given path")
	flagset.String("certfile", "", "path to ssl certificate's cert file")
	flagset.String("keyfile", "", "path to ssl certificate's key file")
	flagset.String("statefile", "", "path a file that goje writes its state on when quitting, and recovering it on startup")
	flagset.String("ntfy-address", "", "address to ntfy server")
	flagset.Bool("statefile-keep-updated", false, "keep state file updated; updating it on every kind of change (don't recommend this on a file on a SSD)")
	return flagset
}

func init() {
	rootCmd.Flags().AddFlagSet(rootFlags())
	rootCmd.MarkFlagsMutuallyExclusive("paused", "not-paused")

	rootCmd.PersistentFlags().StringVarP(&config_file, "config", "c", "", "path to config file. uses default if not specified")
	rootCmd.PersistentFlags().String("loglevel", "error", "log level of goje")
	rootCmd.RegisterFlagCompletionFunc("loglevel", func(cmd *cobra.Command, args []string, to_complete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return []string{"error", "warn", "debug", "info"}, cobra.ShellCompDirectiveNoFileComp
	})
}

var rootCmd = &cobra.Command{
	Use:          "goje",
	Short:        "a pomodoro timer server",
	Long:         "goje is a collaborative pomodoro timer with modern features, suitable for both everyday users and computer nerds",
	Version:      timer.VERSION,
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return setupConfigForCmd(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		t := timer.PomodoroTimer{
			Config: &config.Timer,
		}
		return setupServerAndSignalWatcher(&t)
	},
}

func setupServerAndSignalWatcher(t *timer.PomodoroTimer) error {
	if config.Statefile != "" {
		content, err := os.ReadFile(config.Statefile)
		if err == nil {
			if err := json.Unmarshal(content, &t.State); err != nil {
				return err
			}
		}
	}
	for !quitting {
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
			syscall.SIGABRT,
		)
		if t.State.IsZero() {
			slog.Debug("state is zero.")
			t.Init()
		} else {
			slog.Debug("state is NOT zero.")
		}
		if err := setupDaemons(t); err != nil {
			return err
		}
		go t.Loop(ctx)
		go func() {
			slog.Debug("clean-up goroutine started")
			select {
			case <-ctx.Done():
				slog.Debug("clean-up goroutine quitted")
				return
			case <-sigc:
			}
			quitting = true
			slog.Info("caught deadly signal")
			t.Config.OnQuit.RunSync(t)
			slog.Info("clean up finished. quitting")
			os.Exit(0)
		}()
		restartSig := make(chan os.Signal, 1)
		signal.Notify(restartSig, syscall.SIGHUP)
		select {
		case <-restartSig:
			slog.Info("restart signal (SIGHUP) caught. restarting...")
		case <-ctx.Done():
		}
	}
	return nil
}

func setupConfigForCmd(cmd *cobra.Command) error {
	slog.Info("running setup config")
	if err := readConfig(cmd); err != nil {
		return err
	}

	viper.OnConfigChange(func(e fsnotify.Event) {
		if e.Has(fsnotify.Write) {
			slog.Info("config changed", "path", e.Name, "event", e)
			if err := readConfig(cmd); err != nil {
				slog.Error("invalid config", "err", err)
				os.Exit(1)
			}
			cancel()
		}
	})
	viper.WatchConfig()
	return nil
}

func readConfig(cmd *cobra.Command) error {
	config = AppConfig{Timer: timer.DefaultConfig}
	if config_file != "" {
		// if the config_file arg is passed and doesn't exist
		if _, err := os.Stat(config_file); os.IsNotExist(err) {
			return err
		}
		viper.SetConfigFile(config_file)
	} else {
		slog.Info("using the default config path")
		viper.SetConfigName("config")
		viper.SetConfigType("toml")
		dir := utils.ConfigDir()
		viper.AddConfigPath(dir)
		slog.Info("searching for config.toml in", "dir", dir)
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
	if err := viper.BindPFlags(cmd.LocalFlags()); err != nil {
		return err
	}
	viper.BindPFlag("loglevel", cmd.PersistentFlags().Lookup("loglevel"))
	if err := viper.Unmarshal(&config); err != nil {
		return err
	}
	timer_viper := viper.Sub("timer")
	if timer_viper != nil {
		timer_viper.SetEnvPrefix("goje")
		timer_viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		timer_viper.AutomaticEnv()
		if err := timer_viper.BindPFlags(cmd.LocalFlags()); err != nil {
			return err
		}
		if err := timer_viper.Unmarshal(&config.Timer); err != nil {
			return err
		}
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

func setupDaemons(t *timer.PomodoroTimer) error {
	slog.Info("setting up daemons...")
	if config.Fifo != "" {
		slog.Info("using fifo", "path", config.Fifo)
		utils.Mkfifo(config.Fifo)
		writeToFifo := func(t *timer.PomodoroTimer) {
			if quitting {
				return
			}
			content, _ := json.Marshal(t)
			go func() {
				if err := os.WriteFile(config.Fifo, append(content, '\n'), 0644); err != nil {
					slog.Error("writing to fifo failed", "err", err)
				}
			}()
			slog.Debug("writing to fifo finished")
		}
		slog.Debug("setting up fifo events")
		config.Timer.OnQuit.Append(func(*timer.PomodoroTimer) {
			slog.Debug("removing fifo", "path", config.Fifo)
			if err := os.Remove(config.Fifo); err != nil {
				slog.Error("remove fifo failed", "err", err)
			}
		})
		// initially write to fifo. for times that timer is loaded from a state and
		// OnInit wouldn't fire
		writeToFifo(t)
		config.Timer.OnInit.Append(writeToFifo)
		config.Timer.OnChange.Append(writeToFifo)
		config.Timer.OnModeEnd.Append(writeToFifo)
		config.Timer.OnModeStart.Append(writeToFifo)
	}
	if config.NtfyAddress != "" {

		config.Timer.OnInit.Append((func(pt *timer.PomodoroTimer) {
			reader:=strings.NewReader("Timer started!")
			if _, err := http.Post(config.NtfyAddress, "text/plain" , reader); err != nil {
				log.Fatalln(err)
			}
		}))
		config.Timer.OnModeStart.Append((func(pt *timer.PomodoroTimer) {
			var reader *strings.Reader
			switch pt.State.Mode {
			case 0:
				reader=strings.NewReader("Pomodoro started!")
			case 1:
				reader=strings.NewReader("Short break!")
			case 2:
				reader=strings.NewReader("Long break!")
			}
			http.Post(config.NtfyAddress, "text/plain", reader)
		}))
		config.Timer.OnPause.Append(func(pt *timer.PomodoroTimer) {
			var reader *strings.Reader
			if pt.State.Paused {
				reader=strings.NewReader("Timer paused!")
			} else {
				reader=strings.NewReader("Timer unpaused!")
			}
			http.Post(config.NtfyAddress, "text/plain", reader)
		})
		config.Timer.OnModeEnd.Append((func(pt *timer.PomodoroTimer) {
			if pt.State.Mode  == 2 {
				reader:=strings.NewReader("Long break ended!")
				http.Post(config.NtfyAddress, "text/plain", reader)
			}
		}))
	}
	if config.Statefile != "" {
		slog.Debug("appending statefile")
		write_to_state_file := func(pt *timer.PomodoroTimer) {
			slog.Debug("writing in state file", "statefile", config.Statefile)
			content, _ := json.Marshal(pt.State)
			if err := os.WriteFile(config.Statefile, content, 0644); err != nil {
				slog.Error("write state file failed", "err", err)
			}
		}
		config.Timer.OnQuit.Append(write_to_state_file)
		if config.StatefileKeepUpdated {
			config.Timer.OnInit.Append(write_to_state_file)
			config.Timer.OnChange.Append(write_to_state_file)
			config.Timer.OnModeEnd.Append(write_to_state_file)
			config.Timer.OnModeStart.Append(write_to_state_file)
		}
	}
	if config.Activitywatch {
		aw := activitywatch.Watcher{}
		aw.Init()
		aw.AddEventWatchers(&config.Timer)
	}

	t.Config = &config.Timer

	if config.ExecStart != "" {
		config.Timer.OnModeStart.Append(func(t *timer.PomodoroTimer) {
			runSystemCommand(t, config.ExecStart)
		})
	}
	if config.ExecEnd != "" {
		config.Timer.OnModeEnd.Append(func(t *timer.PomodoroTimer) {
			runSystemCommand(t, config.ExecEnd)
		})
	}
	if config.ExecPause != "" {
		config.Timer.OnPause.Append(func(t *timer.PomodoroTimer) {
			runSystemCommand(t, config.ExecPause)
		})
	}
	if config.TcpAddress != "" {
		tcp_daemon := tcpd.Daemon{
			Timer: t,
		}
		if err := tcp_daemon.InitializeListener(config.TcpAddress); err != nil {
			return err
		}
		slog.Info("running tcp daemon", "address", config.TcpAddress)
		go tcp_daemon.Run(ctx)
	}
	if config.HttpAddress != "" {
		config.httpDaemon = &httpd.Daemon{
			Timer:   t,
			Clients: &sync.Map{},
		}
		config.httpDaemon.Init()
		config.httpDaemon.SetupEvents()
		config.httpDaemon.JsonRoutes()
		if !config.NoWebgui {
			go runWebgui(config.HttpAddress)
		}
		go config.httpDaemon.Run(config.HttpAddress, config.Certfile, config.Keyfile, ctx)
	}
	return nil
}

func runWebgui(address string) {
	slog.Debug("setting up webgui routes")
	config.httpDaemon.WebguiRoutes(config.CustomCss)
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

func runSystemCommand(t *timer.PomodoroTimer, cmd string) {
	content, _ := json.Marshal(t)
	go func() {
		if err := exec.Command(cmd, string(content)).Run(); err != nil {
			slog.Error("running system command failed", "cmd", cmd, "err", err, "content", content)
		}
	}()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
