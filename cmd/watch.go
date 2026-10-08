package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/islands"
	"github.com/Yendric/geny/site"
	"github.com/Yendric/geny/vite"
	"github.com/fatih/color"
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

		s := site.New(cfg)
		checks := startIslandChecker(ctx)
		gate := newBuildGate()
		var last site.Result

		go func() {
			timer := time.NewTimer(0)
			needsRebuild := true
			islandsTouched := false
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}

					if event.Op == fsnotify.Chmod {
						continue
					}
					gate.pending()

					// watch newly created directories
					if event.Op.Has(fsnotify.Create) {
						if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
							if err := addWatchersRecursive(watcher, event.Name); err != nil {
								log.Println("error:", err)
							}
						}
					}
					if isInDir(event.Name, cfg.IslandsDir) {
						islandsTouched = true
					} else {
						needsRebuild = true
					}
					timer.Reset(time.Millisecond * 100)
				case err, ok := <-watcher.Errors:
					if !ok {
						return
					}
					log.Println("error:", err)
				case <-timer.C:
					// island edits only change types, vite hot-reloads the code itself
					if islandsTouched && !needsRebuild {
						needsRebuild = islandSetChanged(cfg, last)
					}
					islandsTouched = false
					if needsRebuild {
						needsRebuild = false
						if result, ok := rebuild(cfg, s); ok {
							last = result
						}
					}
					gate.done()
					checks.request(last)
				}
			}
		}()

		if err := addWatchersRecursive(watcher, cfg.ContentDir); err != nil {
			return err
		}

		if err := addWatchersRecursive(watcher, cfg.TemplatesDir); err != nil {
			return err
		}

		if _, err := os.Stat(cfg.IslandsDir); err == nil {
			if err := addWatchersRecursive(watcher, cfg.IslandsDir); err != nil {
				return err
			}
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

		return runStepE(fmt.Sprintf("Serving the site on port %d", port), func() error { return serve(ctx, gate.hold(http.FileServer(http.Dir(cfg.BuildDir))), port) })
	},
}

func init() {
	rootCmd.AddCommand(watchCmd)

	watchCmd.Flags().BoolP("serve", "s", false, "Serve the site on a local webserver")
	watchCmd.Flags().IntP("port", "p", 8080, "Change the local webserver port from the default 8080")
}

// rebuild clears the build directory's contents instead of deleting it:
// watchers (e.g. Vite's) holding the directory open would not survive that.
func rebuild(cfg common.Config, s *site.Site) (site.Result, bool) {
	var result site.Result
	err := runStepE("Rebuilding...", func() error {
		if err := clearDir(cfg.BuildDir); err != nil {
			return err
		}

		if err := copy.Copy(cfg.PublicDir, cfg.BuildDir); err != nil {
			return err
		}

		var err error
		result, err = s.Generate()
		return err
	})
	if err != nil {
		fmt.Println("Something went wrong:", err)
		return result, false
	}
	return result, true
}

func islandSetChanged(cfg common.Config, last site.Result) bool {
	current, err := islands.Scan(cfg.IslandsDir)
	return err != nil || last.Islands == nil || !current.SameNames(last.Islands)
}

func isInDir(path, dir string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// holds page requests while a rebuild is pending, so a browser reload
// triggered by the source change itself never sees the previous build
type buildGate struct {
	mu    sync.Mutex
	ready chan struct{}
}

func newBuildGate() *buildGate {
	return &buildGate{ready: make(chan struct{})}
}

func (g *buildGate) pending() {
	g.mu.Lock()
	defer g.mu.Unlock()
	select {
	case <-g.ready:
		g.ready = make(chan struct{})
	default:
	}
}

func (g *buildGate) done() {
	g.mu.Lock()
	defer g.mu.Unlock()
	select {
	case <-g.ready:
	default:
		close(g.ready)
	}
}

func (g *buildGate) hold(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.mu.Lock()
		ready := g.ready
		g.mu.Unlock()
		select {
		case <-ready:
			next.ServeHTTP(w, r)
		case <-r.Context().Done():
		}
	})
}

type islandChecker struct {
	results chan site.Result
}

// runs island prop checks in the background, only the latest request is kept
func startIslandChecker(ctx context.Context) islandChecker {
	c := islandChecker{results: make(chan site.Result, 1)}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case result := <-c.results:
				if err := result.CheckIslands(); err != nil {
					color.Yellow("Island props do not type check:\n%v", err)
				}
			}
		}
	}()
	return c
}

func (c islandChecker) request(result site.Result) {
	if len(result.Usages) == 0 {
		return
	}
	select {
	case <-c.results:
	default:
	}
	c.results <- result
}
