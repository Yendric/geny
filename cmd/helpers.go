package cmd

import (
	"fmt"
	"net/http"
	"os"
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

	color.Yellow(step + "..." + strings.Repeat(" ", padding) + "[Busy]")
	err := f()
	if err != nil {
		color.Red(step + "..." + strings.Repeat(" ", padding) + "[Fail]")
		fmt.Print("Something went wrong: ", err)
		if quitOnFail {
			os.Exit(1)
		}
	}
	color.Green(step + "..." + strings.Repeat(" ", padding) + "[Done]")
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
