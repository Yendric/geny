package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yendric/geny/util"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

func runStepQuit(step string, f func() error) {
	runStep(step, f, true)
}

func runStepRecover(step string, f func() error) {
	runStep(step, f, false)
}

func runStep(step string, f func() error, quitOnFail bool) {
	if err := runStepE(step, f); err != nil {
		fmt.Println("Something went wrong:", err)
		if quitOnFail {
			os.Exit(1)
		}
	}
}

func runStepE(step string, f func() error) error {
	padding := 40 - len(step)
	if padding < 0 {
		padding = 0
	}
	prefix := step + "..." + strings.Repeat(" ", padding)

	color.Yellow(prefix + "[Busy]")
	if err := f(); err != nil {
		color.Red(prefix + "[Fail]")
		return err
	}
	color.Green(prefix + "[Done]")
	return nil
}

func runBuildCommand(runCmd string) error {
	out, err := util.ShellCommand(runCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("running %q: %w\n%s", runCmd, err, out)
	}
	return nil
}

func clearDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o755)
	}
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(dir, entry.Name())); err != nil {
			return err
		}
	}
	return nil
}

func addWatchersRecursive(watcher *fsnotify.Watcher, dir string) error {
	err := watcher.Add(dir)
	if err != nil {
		return err
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			err = addWatchersRecursive(watcher, dir+"/"+file.Name())
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func serve(ctx context.Context, buildDir string, port int) error {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(buildDir)))
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
