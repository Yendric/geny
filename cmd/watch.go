package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/generator"
	"github.com/Yendric/geny/indexer"
	"github.com/fsnotify/fsnotify"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:     "watch",
	Aliases: []string{"watch", "w", "run"},
	Short:   "Continously generates the static site when files change",
	Run: func(cmd *cobra.Command, args []string) {
		runCmd, err := cmd.Flags().GetString("run")
		if err != nil {
			log.Fatal(err)
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		go func() {
			timer := time.NewTimer(0)
			for {
				select {
				case _, ok := <-watcher.Events:
					if !ok {
						return
					}
					timer.Reset(time.Millisecond * 100)
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				case <-timer.C:
					rebuild(runCmd)
				}
			}
		}()

		err = addWatchersRecursive(watcher, common.CONTENT_DIR)
		if err != nil {
			log.Fatal(err)
		}

		err = addWatchersRecursive(watcher, common.TEMPLATES_DIR)
		if err != nil {
			log.Fatal(err)
		}

		shouldServe, err := cmd.Flags().GetBool("serve")
		if err != nil || !shouldServe {
			// Loop forever
			<-make(chan struct{})
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.Fatal(err)
		}

		runStepQuit(fmt.Sprintf("Serving the site on port %d", port), func() error { return serve(port) })
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().BoolP("serve", "s", false, "Serve the site on a local webserver")
	watchCmd.Flags().IntP("port", "p", 8080, "Change the local webserver port from the default 8080")
	watchCmd.Flags().String("run", "", "Run a command before building the site")
}

func rebuild(runCmd string) {
	runStepRecover("Rebuilding...", func() error {
		err := os.RemoveAll(common.BUILD_DIR)
		if err != nil {
			return err
		}

		err = copy.Copy(common.PUBLIC_DIR, common.BUILD_DIR)
		if err != nil {
			return err
		}

		if runCmd != "" {
			err := exec.Command("sh", "-c", runCmd).Run()
			if err != nil {
				return err
			}
		}

		content, err := indexer.IndexContent()
		if err != nil {
			return err
		}

		err = generator.GenerateFiles(content)
		return err
	})
}
