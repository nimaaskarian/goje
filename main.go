package main

import (
	"os"
	"os/exec"

	"github.com/nimaaskarian/goje/cmd"
	"github.com/nimaaskarian/goje/utils"
)

func main() {
  if utils.IsRunningInTerminal() {
    cmd.Execute()
  } else {
    exec.Command("cmd.exe", "/c", os.Args[0])
  }
}
