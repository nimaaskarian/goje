package requests

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/nimaaskarian/pomodoro/timer"
)

const (
	Pause     = "pause"
	Seek      = "seek"
	Reset     = "reset"
	CycleMode = "cyclemode"
	Skip      = "skip"
	Init      = "init"
	Timer     = "timer"
	Commands  = "commands"
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

func ParseInput(timer *timer.Timer, input string) (string, string, error) {
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
	case CycleMode:
	case Skip:
		out, err = cycleModeCmd(timer, splited)
	case Commands:
		out, err = fmt.Sprintf(`command: %s
command: %s
command: %s
command: %s
command: %s
command: %s
command: %s
`, Pause, Seek, Reset, Init, CycleMode, Skip, Commands), nil
	default:
		out, err = "", errors.New(fmt.Sprintf("command not found %q", splited[0]))
		cmd = ""
	}
	return cmd, out, err
}

func pauseCmd(timer *timer.Timer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer.Paused = !timer.Paused
	case 2:
		var err error
		timer.Paused, err = parseBool(args[1])
		if err != nil {
			return "", err
		}
	default:
		return "", TooManyArgsError{args[0]}
	}
	return "", nil
}

func seekCmd(timer *timer.Timer, args []string) (string, error) {
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

func resetCmd(timer *timer.Timer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer.Reset()
	default:
		return "", TooManyArgsError{args[0]}
	}
	return "", nil
}

func cycleModeCmd(timer *timer.Timer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer.CycleMode()
	default:
		return "", TooManyArgsError{args[0]}
	}
	return "", nil
}

func timerCmd(timer *timer.Timer, args []string) (string, error) {
	switch len(args) {
	case 1:
		timer_value := reflect.ValueOf(*timer)
		typ := timer_value.Type()
		var out string
		for i := range timer_value.NumField() {
			name := typ.Field(i).Name
			if name == "Config" {
				continue
			}
			value := timer_value.Field(i).Interface()
			out += fmt.Sprintln(name+":", value)
		}
		return out, nil
  case 2:
    name := args[1]
		timer_value := reflect.ValueOf(*timer)
    field := timer_value.FieldByName(name)
    if field.IsValid() {
      return fmt.Sprintln(field.Interface()), nil
    } else {
      return "", errors.New(fmt.Sprintf("field doesn't exist on timer: %q", name))
    }
	default:
		return "", TooManyArgsError{args[0]}
	}
}

func initCmd(timer *timer.Timer, args []string) (string, error) {
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
