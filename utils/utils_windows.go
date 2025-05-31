//go:build windows && !darwin && !unix
// +build windows,!darwin,!unix

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
)

func OpenURL(url string) error {
	return exec.Command("cmd.exe", "/C", "start "+url).Run()
}

func ConfigDir() string {
  return filepath.Join(os.Getenv("APPDATA"), "goje")
}

func IsRunningInTerminal() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
