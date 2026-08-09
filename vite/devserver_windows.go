//go:build windows

package vite

import (
	"os/exec"
	"strconv"
)

func setProcessGroup(cmd *exec.Cmd) {}

func terminate(cmd *exec.Cmd) error {
	if cmd.Process == nil {
		return nil
	}
	return exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
}
