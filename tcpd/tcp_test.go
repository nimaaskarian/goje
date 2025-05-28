package tcpd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/nimaaskarian/tom/timer"
)

type Command struct {
  Name, Cmd string
}

func genCommands() (chan Command, error) {
  ch := make(chan Command)
  file, err := os.Open("./tcp.go")
  if err != nil {
    return nil, err
  }
  reader := bufio.NewReader(file)
  read := false
  go func () {
    defer close(ch)
    for  {
      line_bytes, _, err := reader.ReadLine();
      if err != nil {
        continue
      }
      line := string(bytes.TrimSpace(line_bytes))
      if read && line == ")" {
        read = false
        break
      }
      if read {
        splited_unfiltered := strings.Split(line, " ")
        splited := make([]string, 0, len(splited_unfiltered))
        for _, item := range splited_unfiltered {
          if item != "" {
            splited = append(splited, item)
          }
        }
        cmd_quoted := splited[2]
        cmd_name := splited[0]
        var cmd string
        fmt.Sscanf(cmd_quoted, "%q", &cmd)

        ch <- Command { Name: cmd_name, Cmd: cmd }
      }
      if line == "const (" {
        read = true
      }
    }
  }()
  return ch, nil
}

func TestHandleAllCommands(t * testing.T) {
  pomodoro_timer := timer.Timer {
    Config: timer.DefaultConfig,
  }
  pomodoro_timer.Init()
  cmds, err := genCommands()
  if err != nil {
    t.Fatal(err)
  }
  for cmd := range cmds {
    cmd_out, _, _ := ParseInput(&pomodoro_timer, cmd.Cmd)
    if cmd_out == "" {
      t.Fatalf("command %s (%q) isn't matched in the ParseInput function", cmd.Name, cmd.Cmd)
    }
    if cmd_out != cmd.Cmd {
      t.Fatalf("command %s (%q) isn't returning the same value in the ParseInput function: %q", cmd.Name, cmd, cmd_out)
    }
  }
}

func TestPrintAllCommands(t * testing.T) {
  pomodoro_timer := timer.Timer {
    Config: timer.DefaultConfig,
  }
  pomodoro_timer.Init()
  cmds, err := genCommands()
  if err != nil {
    t.Fatal(err)
  }
  _, out, _ := ParseInput(&pomodoro_timer, Commands)
  command_outputs := make([]string, 0)
  for line := range strings.Lines(out) {
    command_outputs = append(command_outputs, strings.TrimSpace(line[9:]))
  }
  for cmd := range cmds {
    if !slices.Contains(command_outputs, cmd.Cmd) {
      t.Fatalf("command %s (%q) isn't printed in the \"Commands\" command", cmd.Name, cmd.Cmd)
    }
  }
}
