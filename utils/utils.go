package utils

import (
	"fmt"
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
