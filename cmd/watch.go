package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/generator"
	"github.com/Yendric/geny/indexer"
	"github.com/Yendric/geny/vite"
	"github.com/fsnotify/fsnotify"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:     "watch",
	Aliases: []string{"watch", "w", "run"},
	Short:   "Continously generates the static site when files change",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		cfg, err := common.LoadConfig()
		if err != nil {
			return err
		}
		cfg.DevMode = true

		if err := clearDir(cfg.BuildDir); err != nil {
			return err
		}

		if cfg.Vite.Enabled {
			stop, err := vite.StartDevServer(cfg)
			if err != nil {
				return err
			}
			defer stop()
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
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}

					if event.Op == fsnotify.Chmod {
						continue
					}

					// watch newly created directories
					if event.Op.Has(fsnotify.Create) {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							if err := addWatchersRecursive(watcher, event.Name); err != nil {
								log.Println("error:", err)
							}
						}
					}
					timer.Reset(time.Millisecond * 100)
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				case <-timer.C:
					rebuild(cfg)
				}
			}
		}()

		if err := addWatchersRecursive(watcher, cfg.ContentDir); err != nil {
			return err
		}

		if err := addWatchersRecursive(watcher, cfg.TemplatesDir); err != nil {
			return err
		}

		if _, err := os.Stat(cfg.PublicDir); err == nil {
			if err := addWatchersRecursive(watcher, cfg.PublicDir); err != nil {
				return err
			}
		}

		shouldServe, err := cmd.Flags().GetBool("serve")
		if err != nil {
			return err
		}
		if !shouldServe {
			// Run until interrupted.
			<-ctx.Done()
			return nil
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		return runStepE(fmt.Sprintf("Serving the site on port %d", port), func() error { return serve(ctx, cfg.BuildDir, port) })
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().BoolP("serve", "s", false, "Serve the site on a local webserver")
	watchCmd.Flags().IntP("port", "p", 8080, "Change the local webserver port from the default 8080")
}

// rebuild clears the build directory's contents instead of deleting it:
// watchers (e.g. Vite's) holding the directory open would not survive that.
func rebuild(cfg common.Config) {
	runStepRecover("Rebuilding...", func() error {
		if err := clearDir(cfg.BuildDir); err != nil {
			return err
		}

		if err := copy.Copy(cfg.PublicDir, cfg.BuildDir); err != nil {
			return err
		}

		content, err := indexer.IndexContent(cfg)
		if err != nil {
			return err
		}

		return generator.GenerateFiles(cfg, content)
	})
}
