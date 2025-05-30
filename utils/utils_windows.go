//go:build windows && !darwin && !unix
// +build windows,!darwin,!unix

package utils

import "os/exec"

func OpenURL(url string) error {
	return exec.Command("cmd.exe", "/C", "start "+url).Run()
}
