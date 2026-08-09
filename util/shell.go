package util

import (
	"os/exec"
	"runtime"
)

// run shell command (cross platform support)
func ShellCommand(command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd", "/C", command)
	}
	return exec.Command("sh", "-c", command)
}
