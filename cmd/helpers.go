package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/Yendric/geny/common"
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
	padding := 40 - len(step)
	if padding < 0 {
		padding = 0
	}
	prefix := step + "..." + strings.Repeat(" ", padding)

	color.Yellow(prefix + "[Busy]")
	if err := f(); err != nil {
		color.Red(prefix + "[Fail]")
		fmt.Println("Something went wrong:", err)
		if quitOnFail {
			os.Exit(1)
		}
		return
	}
	color.Green(prefix + "[Done]")
}

func runBuildCommand(runCmd string) error {
	out, err := exec.Command("sh", "-c", runCmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("running %q: %w\n%s", runCmd, err, out)
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

func serve(port int) error {
	fs := http.FileServer(http.Dir(common.BUILD_DIR))
	http.Handle("/", fs)

	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	return err
}
