package requests

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/nimaaskarian/pomodoro/timer"
)

func TestHandleAllCommands(t * testing.T) {
  file, err := os.Open("./requests.go")
  if err != nil {
    t.Fatal(err)
  }
  reader := bufio.NewReader(file)
  read := false
  pomodoro_timer := timer.Timer {
    Config: timer.DefaultConfig,
  }
  pomodoro_timer.Init()
  for  {
    line_bytes, _, err := reader.ReadLine();
    if err != nil {
      t.Fatal("Error when reading from the reader")
      break
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
      cmdo, _, _ := ParseInput(&pomodoro_timer, cmd)
      if cmdo == "" {
        t.Fatalf("command %s (%q) isn't matched in the ParseInput function", cmd_name, cmd)
      }
      if cmdo != cmd {
        t.Fatalf("command %s (%q) isn't returning the same value in the ParseInput function: %q", cmd_name, cmd, cmdo)
      }
    }
    if line == "const (" {
      read = true
    }
  }
}
