package tcpd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/nimaaskarian/goje/timer"
)

const (
	Pause          = "pause"
	Seek           = "seek"
	Reset          = "reset"
	Next           = "next"
	Prev           = "prev"
	Skip           = "skip"
	Init           = "init"
	Sessions       = "sessions"
	ConfigSessions = "config-sessions"
	Timer          = "timer"
	Commands       = "commands"
)

type TooManyArgsError struct {
	cmd string
}

func (e TooManyArgsError) Error() string {
	return "too many arguments for \"" + e.cmd + "\""
}

type WrongNumberOfArgsError struct {
	cmd string
}

func (e WrongNumberOfArgsError) Error() string {
	return "wrong number of arguments for \"" + e.cmd + "\""
}

func ParseInput(timer *timer.PomodoroTimer, input string) (string, string, error) {
	splited := strings.Split(input, " ")
	cmd := splited[0]
	var err error
	var out string
	switch splited[0] {
	case Pause:
		out, err = pauseCmd(timer, splited)
	case Seek:
		out, err = seekCmd(timer, splited)
	case Reset:
		out, err = resetCmd(timer, splited)
	case Init:
		out, err = initCmd(timer, splited)
	case Timer:
		out, err = timerCmd(timer, splited)
	case Prev:
		out, err = prevModeCmd(timer, splited)
	case Skip, Next:
		out, err = nextModeCmd(timer, splited)
	case Sessions:
		out, err = sessionsCmd(timer, splited)
	case ConfigSessions:
		out, err = configSessionsCmd(timer, splited)
	case Commands:
		out, err = fmt.Sprintf(`command: %s
command: %s
command: %s
command: %s
command: %s
command: %s
command: %s
command: %s
command: %s
command: %s
command: %s
`, Pause, Seek, Reset, Init, Prev, Next, Skip, Sessions, Timer, ConfigSessions, Commands), nil
	default:
		out, err = "", fmt.Errorf("command not found %q", splited[0])
		cmd = ""
	}
	return cmd, out, err
}

func pauseCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer.Pause()
	case 2:
		var err error
		timer.State.Paused, err = parseBool(args[1])
		if err != nil {
			return "", err
		}
		if !timer.Config.OnSet.Run(timer) {
			timer.Config.OnPause.Run(timer)
		}
	default:
		return "", TooManyArgsError{args[0]}
	}
	return "", nil
}

func sessionsCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 2:
		err := parseRelativeNumber(args[1], &timer.State.FinishedSessions)
		if err != nil {
			return "", err
		}
		if !timer.Config.OnSet.Run(timer) {
			timer.Config.OnChange.Run(timer)
		}
	default:
		return "", WrongNumberOfArgsError{args[0]}
	}
	return "", nil
}

func configSessionsCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 2:
		err := parseRelativeNumber(args[1], &timer.Config.Sessions)
		if err != nil {
			return "", err
		}
		if !timer.Config.OnSet.Run(timer) {
			timer.Config.OnChange.Run(timer)
		}
	default:
		return "", WrongNumberOfArgsError{args[0]}
	}
	return "", nil
}

func parseRelativeNumber(input string, output *uint) (err error) {
	if strings.HasPrefix(input, "+") || strings.HasPrefix(input, "-") {
		var count uint64
		count, err = strconv.ParseUint(input[1:], 10, 32)
		if err != nil {
			return err
		}
		if input[0] == '+' {
			*output += uint(count)
		} else {
			*output -= uint(count)
		}
	} else {
		var count uint64
		count, err = strconv.ParseUint(input, 10, 32)
		if err != nil {
			return err
		}
		*output = uint(count)
	}
	return nil
}

func seekCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 2:
		if strings.HasPrefix(args[1], "+") || strings.HasPrefix(args[1], "-") {
			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return "", err
			}
			timer.SeekAdd(duration)
		} else {
			duration, err := time.ParseDuration(args[1])
			if err != nil {
				return "", err
			}
			timer.SeekTo(duration)
		}
	default:
		return "", WrongNumberOfArgsError{args[0]}
	}
	return "", nil
}

func resetCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer.Reset()
	default:
		return "", TooManyArgsError{args[0]}
	}
	return "", nil
}

func nextModeCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer.SwitchNextMode()
	default:
		return "", TooManyArgsError{args[0]}
	}
	return "", nil
}

func prevModeCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer.SwitchPrevMode()
	default:
		return "", TooManyArgsError{args[0]}
	}
	return "", nil
}

func timerCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer_value := reflect.ValueOf(*timer)
		typ := timer_value.Type()
		var out string
		for i := range timer_value.NumField() {
			name := typ.Field(i).Name
			if name[0] >= 'A' && name[0] <= 'Z' {
				obj := timer_value.Field(i).Interface()
				out += fmt.Sprintf("%s: %v\n", name, obj)
			}
		}
		return out, nil
	case 2:
		name := args[1]
		timer_value := reflect.ValueOf(*timer)
		field := timer_value.FieldByName(name)
		if field.IsValid() {
			return fmt.Sprintln(field.Interface()), nil
		} else {
			return "", fmt.Errorf("field doesn't exist on timer: %q", name)
		}
	default:
		return "", TooManyArgsError{args[0]}
	}
}

func initCmd(timer *timer.PomodoroTimer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer.Init()
	default:
		return "", TooManyArgsError{args[0]}
	}
	return "", nil
}

func parseBool(input string) (bool, error) {
	if input != "1" && input != "0" {
		return false, errors.New("boolean (0/1) expected: \"" + input + "\"")
	}
	return input == "1", nil
}

type Daemon struct {
	Timer    *timer.PomodoroTimer
	Listener net.Listener
}

func (d *Daemon) InitializeListener(address string) error {
	var err error
	d.Listener, err = net.Listen("tcp", address)
	if err != nil {
		return err
	}
	return nil
}

func (d *Daemon) handleConnection(conn net.Conn) {
	conn.Write([]byte("OK goje "+timer.VERSION+"\n"))
	defer conn.Close()
	reader := bufio.NewReader(conn);
	for {
		buff, err := reader.ReadString('\n')
		if err != nil && errors.Is(err, io.EOF) {
			break
		}
		cmd, out, err := ParseInput(d.Timer, strings.TrimSpace(buff))
		if err != nil {
			slog.Error("command throw error", "err", err)
			conn.Write(fmt.Appendf(nil, "ACK {%s} %s\n", cmd, err))
		} else {
			conn.Write(append([]byte(out), []byte("OK\n")...))
		}
	}
}

func (d *Daemon) Run(ctx context.Context) {
	for {
		select {
			case <-ctx.Done():
				slog.Info("closing tcpd connection")
				d.Listener.Close()
				return
			default:
				slog.Info("accepting connection...")
				conn, err := d.Listener.Accept()
				if err != nil {
					slog.Warn("connection throw error", "err", err)
					return
				}
				slog.Info("connection added!")
				go d.handleConnection(conn)
		}
	}
}
