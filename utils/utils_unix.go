//go:build !windows && !darwin && unix
// +build !windows,!darwin,unix

package utils

import (
	"os/exec"
	"syscall"
)

func OpenURL(url string) error {
	cmd := exec.Command("xdg-open", url)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd.Run()
}
