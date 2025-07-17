//go:build !windows && darwin && !unix
// +build !windows,darwin,!unix

package utils

import "os/exec"

const DEFAULT_EDITORS = "vim"

func OpenURL(url string) error {
	return exec.Command("open", url).Run()
}

func ConfigDir() string {
	return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "goje")
}

func MakeFifo(path string) {
	syscall.Mkfifo(path, 0644)
}
