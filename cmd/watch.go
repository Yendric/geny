package cmd

import (
	"fmt"
	"log"
	"os"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := common.LoadConfig()
		if err != nil {
			return err
		}

		runCmd, err := cmd.Flags().GetString("run")
		if err != nil {
			return err
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
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
					rebuild(cfg, runCmd)
				}
			}
		}()

		if err := addWatchersRecursive(watcher, cfg.ContentDir); err != nil {
			return err
		}

		if err := addWatchersRecursive(watcher, cfg.TemplatesDir); err != nil {
			return err
		}

		shouldServe, err := cmd.Flags().GetBool("serve")
		if err != nil {
			return err
		}
		if !shouldServe {
			// Loop forever
			<-make(chan struct{})
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		runStepQuit(fmt.Sprintf("Serving the site on port %d", port), func() error { return serve(cfg.BuildDir, port) })
		return nil
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().BoolP("serve", "s", false, "Serve the site on a local webserver")
	watchCmd.Flags().IntP("port", "p", 8080, "Change the local webserver port from the default 8080")
	watchCmd.Flags().String("run", "", "Run a command before building the site")
}

func rebuild(cfg common.Config, runCmd string) {
	runStepRecover("Rebuilding...", func() error {
		if err := os.RemoveAll(cfg.BuildDir); err != nil {
			return err
		}

		if err := copy.Copy(cfg.PublicDir, cfg.BuildDir); err != nil {
			return err
		}

		if runCmd != "" {
			if err := runBuildCommand(runCmd); err != nil {
				return err
			}
		}

		content, err := indexer.IndexContent(cfg)
		if err != nil {
			return err
		}

		return generator.GenerateFiles(cfg, content)
	})
}
