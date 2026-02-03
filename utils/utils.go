package utils

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func EditorCmd(path string) (*exec.Cmd, error) {
	if os.Getenv("SNAP_REVISION") != "" {
		return nil, fmt.Errorf("Did you install with Snap? ydo is sandboxed and unable to open an editor. Please install ydo with Go or another package manager to enable editing.")
	}
	editor, args := getEditor()
	args = append(args, path)
	return exec.Command(editor, args...), nil
}

func getEditor() (string, []string) {
	editor := strings.Fields(os.Getenv("EDITOR"))
	if len(editor) > 1 {
		return editor[0], editor[1:]
	}
	if len(editor) == 1 {
		return editor[0], []string{}
	}
	return "vim", []string{}
}

// set the command's stdout, stderr and stdin to os's
func CmdStdOs(c *exec.Cmd) {
	c.Stdout = os.Stdout
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
}

func FixHttpAddress(address string) string {
	if strings.HasPrefix(address, "http://") || strings.HasPrefix(address, "https://") {
		return address
	}
	if strings.HasPrefix(address, ":") {
		return "http://localhost" + address
	} else {
		slog.Warn("address doesn't specify either http or https. http is assumed. omit this warning by specifying the protocol in url.", "address", address)
		return "http://" + address
	}
}

type ExpandUser struct {
	home      string
	sep       string
	tilde_sep string
}

// create inner expanduser cache
func NewExpandUser() (*ExpandUser, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sep := string(os.PathSeparator)
	return &ExpandUser{
		home,
		sep,
		"~" + sep,
	}, nil
}

// apply expanduser on a path
func (eu *ExpandUser) Expand(path string) string {
	if !strings.HasPrefix(path, eu.tilde_sep) {
		if path == "~" {
			return eu.home
		}
		return path
	}
	return eu.home + eu.sep + path[len(eu.tilde_sep):]
}
