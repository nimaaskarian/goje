//go:build !windows && !darwin && unix
// +build !windows,!darwin,unix

package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const DEFAULT_EDITORS = "vim"

func OpenURL(url string) error {
	cmd := exec.Command("xdg-open", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd.Run()
}

func ConfigDir() string {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(os.Getenv("HOME"), ".config")
	}
	return filepath.Join(base, "goje")
}
