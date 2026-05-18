package cmd

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/nimaaskarian/goje/activitywatch"
	"github.com/nimaaskarian/goje/httpd"
	"github.com/nimaaskarian/goje/mpris"
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
	CustomCss            string `mapstructure:"custom-css,omitempty"`
	Activitywatch        bool   `mapstructure:"activitywatch,omitempty"`
	NoWebgui             bool   `mapstructure:"no-webgui,omitempty"`
	NoOpenBrowser        bool   `mapstructure:"no-open-browser,omitempty"`
	ExecStart            string `mapstructure:"exec-start,omitempty"`
	ExecEnd              string `mapstructure:"exec-end,omitempty"`
	ExecPause            string `mapstructure:"exec-pause,omitempty"`
	ExecQuit             string `mapstructure:"exec-quit,omitempty"`
	SyncExec             bool   `mapstructure:"sync-exec,omitempty"`
	HttpAddress          string `mapstructure:"http-address,omitempty"`
	TcpAddress           string `mapstructure:"tcp-address,omitempty"`
	Fifo                 string `mapstructure:"fifo,omitempty"`
	Loglevel             string `mapstructure:"loglevel,omitempty"`
	Certfile             string `mapstructure:"certfile,omitempty"`
	Keyfile              string `mapstructure:"keyfile,omitempty"`
	Statefile            string `mapstructure:"statefile,omitempty"`
	NtfyAddress          string `mapstructure:"ntfy-address,omitempty"`
	NtfyClickUrl         string `mapstructure:"ntfy-click-url,omitempty"`
	NtfyAuth             string `mapstructure:"ntfy-auth,omitempty"`
	StatefileKeepUpdated bool   `mapstructure:"statefile-keep-updated,omitempty"`
	Version              bool   `mapstructure:"version,omitempty"`
	Help                 bool   `mapstructure:"help,omitempty"`
	Mpris                bool   `mapstructure:"mpris,omitempty"`
	MprisNoInstance      bool   `mapstructure:"mpris-no-instance,omitempty"`
}

var (
	httpDaemon    *httpd.Daemon
	webguiAddress string
)

// objects that define a path. later used for utils.ExpandUser to get applied on all paths
var filename_fields = []string{
	"fifo", "certfile", "keyfile", "statefile", "custom-css", "exec-start", "exec-end", "exec-pause", "exec-quit",
}

var ctx context.Context
var tcp_ctx context.Context
var http_ctx context.Context
var tcp_cancel context.CancelFunc
var http_cancel context.CancelFunc
var cancel context.CancelFunc

var config = AppConfig{Timer: timer.DefaultConfig}
var old_config = AppConfig{}

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
	flagset.String("exec-quit", "", "command to run when timer quit")
	flagset.Bool("sync-exec", false, "run exec-* hooks synchronously, pausing the timer instead of asynchronously (default)")
	flagset.StringP("tcp-address", "a", "localhost:7800", "address:[port] for tcp pomodoro daemon (doesn't run when empty)")
	flagset.StringP("http-address", "A", "localhost:7900", "address:[port] for http pomodoro api (doesn't run when empty)")
	flagset.Bool("no-webgui", false, "don't run webgui. webgui can't be run without the json server")
	flagset.Bool("no-open-browser", false, "don't open the browser when running webgui")
	flagset.Bool("activitywatch", false, "daemon send's pomodoro data to activitywatch if is present")
	flagset.StringP("fifo", "f", "", "write timer events in a fifo at given path")
	flagset.String("certfile", "", "path to ssl certificate's cert file")
	flagset.String("keyfile", "", "path to ssl certificate's key file")
	flagset.String("statefile", "", "path a file that goje writes its state on when quitting, and recovering it on startup")
	flagset.String("ntfy-address", "", "address to ntfy topic")
	flagset.String("ntfy-click-url", "", "address to open on notification click of subscribers")
	flagset.String("ntfy-auth", "", "username:password to access ntfy topic")
	flagset.Bool("mpris", false, "run a MPRIS interface for goje")
	flagset.Bool("mpris-no-instance", false, "don't append instance to MPRIS's name")
	flagset.Bool("statefile-keep-updated", false, "keep state file updated; updating it on every kind of change (don't recommend this on a file on a SSD)")
	return flagset
}

func init() {
	rootCmd.Flags().AddFlagSet(rootFlags())
	rootCmd.MarkFlagsMutuallyExclusive("paused", "not-paused")

	rootCmd.PersistentFlags().StringVarP(&config_file, "config", "c", "", "path to config file. uses default if not specified")
	rootCmd.PersistentFlags().Var(&loglevel, "loglevel", "log level of goje")
	rootCmd.RegisterFlagCompletionFunc("loglevel", logLevelCompletion)
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
	if config.Statefile != old_config.Statefile {
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
		if err := setupDaemons(t); err != nil {
			return err
		}
		if t.State.IsZero() {
			slog.Debug("state is zero.")
			t.Init()
		} else {
			slog.Debug("state is NOT zero.")
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
			t.Config.Hooks.OnQuit.RunSync(t)
			slog.Info("clean up finished. quitting")
			os.Exit(0)
		}()
		restartSig := make(chan os.Signal, 1)
		signal.Notify(restartSig, syscall.SIGHUP)
		select {
		case <-restartSig:
			old_config = config
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
			old_config = config
			if err := readConfig(cmd); err != nil {
				slog.Error("invalid config. using the old config.", "err", err)
				config = old_config
				return
			}
			temp := old_config.Timer.Hooks
			old_config.Timer.Hooks = timer.TimerConfigHooks{}
			if !reflect.DeepEqual(config, old_config) {
				slog.Info("configs aren't equal. canceling the timer", "old", old_config, "new", config)
				cancel()
			}
			old_config.Timer.Hooks = temp
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
	expanduser, err := utils.NewExpandUser()
	if err != nil {
		slog.Error("failed to initialize expanduser. probably couldn't find home directory")
	} else {
		for _, path_object := range filename_fields {
			expanded := expanduser.Expand(viper.GetString(path_object))
			viper.Set(path_object, expanded)
		}
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

	t.Config = &config.Timer

	for _, script := range []struct {
		command     string
		event       *timer.TimerConfigHook
		old_command string
	}{
		{config.ExecStart, &config.Timer.Hooks.OnModeStart, old_config.ExecStart},
		{config.ExecEnd, &config.Timer.Hooks.OnModeEnd, old_config.ExecEnd},
		{config.ExecPause, &config.Timer.Hooks.OnPause, old_config.ExecPause},
		{config.ExecQuit, &config.Timer.Hooks.OnQuit, old_config.ExecQuit},
	} {
		if script.command != "" {
			script.event.Append(func(pt *timer.PomodoroTimer) {
				var temp bool
				if config.SyncExec {
					pt.State.Mu.Lock()
					temp = pt.State.Paused
					pt.State.Paused = true
				}
				runSystemCommand(pt, script.command)
				if config.SyncExec {
					pt.State.Paused = temp
					pt.State.Mu.Unlock()
				}
			})
		}
	}

	if config.Fifo != old_config.Fifo {
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
		config.Timer.Hooks.OnQuit.Append(func(*timer.PomodoroTimer) {
			slog.Debug("removing fifo", "path", config.Fifo)
			if err := os.Remove(config.Fifo); err != nil {
				slog.Error("remove fifo failed", "err", err)
			}
		})
		// initially write to fifo. for times that timer is loaded from a state and
		// Hooks.OnInit wouldn't fire
		writeToFifo(t)
		config.Timer.Hooks.OnInit.Append(writeToFifo)
		config.Timer.Hooks.OnChange.Append(writeToFifo)
		config.Timer.Hooks.OnModeEnd.Append(writeToFifo)
		config.Timer.Hooks.OnModeStart.Append(writeToFifo)
	}
	if config.NtfyAddress != old_config.NtfyAddress {
		ntfySetup(&config)
	}
	if config.Statefile != old_config.Statefile {
		slog.Debug("appending statefile")
		write_to_state_file := func(pt *timer.PomodoroTimer) {
			slog.Debug("writing in state file", "statefile", config.Statefile)
			content, _ := json.Marshal(pt.State)
			if err := os.WriteFile(config.Statefile, content, 0644); err != nil {
				slog.Error("write state file failed", "err", err)
			}
		}
		config.Timer.Hooks.OnQuit.Append(write_to_state_file)
		if config.StatefileKeepUpdated {
			config.Timer.Hooks.OnInit.Append(write_to_state_file)
			config.Timer.Hooks.OnChange.Append(write_to_state_file)
			config.Timer.Hooks.OnModeEnd.Append(write_to_state_file)
			config.Timer.Hooks.OnModeStart.Append(write_to_state_file)
		}
	}
	if config.Activitywatch {
		aw := activitywatch.Watcher{}
		aw.Init()
		aw.AddEventWatchers(&config.Timer)
	}

	slog.Info("checking tcp", "old", old_config.TcpAddress, "new", config.TcpAddress)
	if config.TcpAddress != old_config.TcpAddress {
		slog.Info("restarting tcp", "old", old_config.TcpAddress, "new", config.TcpAddress)
		if tcp_cancel != nil {
			slog.Info("calling tcp cancel")
			tcp_cancel()
		}
		tcp_ctx, tcp_cancel = context.WithCancel(context.Background())
		tcp_daemon := tcpd.Daemon{
			Timer: t,
		}
		if err := tcp_daemon.InitializeListener(config.TcpAddress); err != nil {
			return err
		}
		slog.Info("running tcp daemon", "address", config.TcpAddress)
		go tcp_daemon.Run(tcp_ctx)
	}
	if config.HttpAddress != old_config.HttpAddress {
		if http_cancel != nil {
			http_cancel()
		}
		http_ctx, http_cancel = context.WithCancel(context.Background())
		httpDaemon = &httpd.Daemon{
			Timer:   t,
			Clients: &sync.Map{},
		}
		httpDaemon.Init()
		httpDaemon.SetupEvents()
		httpDaemon.JsonRoutes()
		if !config.NoWebgui {
			runWebgui(config.HttpAddress)
		}
		go httpDaemon.Run(config.HttpAddress, config.Certfile, config.Keyfile, http_ctx)
	}
	if config.Mpris {
		instance, err := mpris.NewInstance(t, &mpris.InstanceOpts{NoInstance: config.MprisNoInstance, WebguiAddress: webguiAddress})
		if err != nil {
			return err
		}
		instance.Start(ctx)
		config.Timer.Hooks.OnQuit.Append(func(pt *timer.PomodoroTimer) {
			instance.Close()
		})
	}
	return nil
}

func runWebgui(address string) {
	slog.Debug("setting up webgui routes")
	httpDaemon.WebguiRoutes(config.CustomCss)
	if !config.NoOpenBrowser {
		// set webguiAddress used in mpris raise
		webguiAddress = address
		go utils.OpenURL(utils.FixHttpAddress(address))
	}
}

func runSystemCommand(t *timer.PomodoroTimer, cmd string) {
	content, _ := json.Marshal(t)
	slog.Info("starting command", "cmd", cmd)
	if err := exec.Command(cmd, string(content)).Run(); err != nil {
		slog.Error("running system command failed", "cmd", cmd, "err", err, "content", content)
	}
	slog.Info("finished command", "cmd", cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
