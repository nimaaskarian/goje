package cmd

import (
	"log/slog"
	"errors"

	"github.com/spf13/cobra"
)

type LogLevel slog.Level
var loglevel = LogLevel(slog.LevelError)

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

func (level * LogLevel) String() string {
	switch *level {
	case LogLevel(slog.LevelWarn):
		return "warn"
	case LogLevel(slog.LevelError):
		return "error"
	case LogLevel(slog.LevelDebug):
		return "debug"
	case LogLevel(slog.LevelInfo):
		return "info"
	}
	return ""
}

func (level * LogLevel) Type() string {
	return "loglevel"
}

func logLevelCompletion(cmd *cobra.Command, args []string, to_complete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return []string{"error", "warn", "debug", "info"}, cobra.ShellCompDirectiveNoFileComp
}
