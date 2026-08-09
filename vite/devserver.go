package vite

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/util"
	"github.com/fatih/color"
)

// launches Vite dev server & writes hot file
// returned stop function terminates server and removes hot file
func StartDevServer(cfg common.Config) (func(), error) {
	hotFile := cfg.Vite.HotFile
	if err := os.MkdirAll(filepath.Dir(hotFile), 0o755); err != nil {
		return nil, fmt.Errorf("vite: creating hot file directory: %w", err)
	}

	cmd := util.ShellCommand(cfg.Vite.DevCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	setProcessGroup(cmd)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("vite: starting dev server: %w", err)
	}

	if err := os.WriteFile(hotFile, []byte(cfg.Vite.DevServerURL), 0o644); err != nil {
		_ = terminate(cmd)
		return nil, fmt.Errorf("vite: writing hot file: %w", err)
	}

	var once sync.Once
	stopping := make(chan struct{})
	stop := func() {
		once.Do(func() {
			close(stopping)
			_ = terminate(cmd)
			_ = os.Remove(hotFile)
		})
	}

	go func() {
		err := cmd.Wait()
		select {
		case <-stopping:
		default:
			if err != nil {
				color.Red("The vite dev server exited unexpectedly: %v", err)
			} else {
				color.Red("The vite dev server exited unexpectedly")
			}
			_ = os.Remove(hotFile)
		}
	}()

	return stop, nil
}
