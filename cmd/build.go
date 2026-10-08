package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/site"
	"github.com/Yendric/geny/vite"
	"github.com/fatih/color"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:     "build",
	Aliases: []string{"b", "generate", "run"},
	Short:   "Generates the static site",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := common.LoadConfig()
		if err != nil {
			return err
		}

		runStepQuit("Removing old builds", func() error {
			return os.RemoveAll(cfg.BuildDir)
		})

		runStepQuit("Copying assets", func() error {
			return copy.Copy(cfg.PublicDir, cfg.BuildDir)
		})

		if cfg.Vite.Enabled {
			runStepQuit("Building assets with Vite", func() error {
				return vite.Build(cfg)
			})
		}

		var result site.Result
		runStepQuit("Generating html", func() error {
			var err error
			result, err = site.New(cfg).Generate()
			return err
		})

		if len(result.Usages) > 0 {
			runStepQuit("Checking island props", result.CheckIslands)
		}

		color.New(color.BgGreen).Println("Your site has been generated!")

		shouldServe, err := cmd.Flags().GetBool("serve")
		if err != nil {
			return err
		}
		if !shouldServe {
			return nil
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		runStepQuit(fmt.Sprintf("Serving the site on port %d", port), func() error { return serve(cmd.Context(), http.FileServer(http.Dir(cfg.BuildDir)), port) })
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolP("serve", "s", false, "Serve the site on a local webserver")
	buildCmd.Flags().IntP("port", "p", 8080, "Change the local webserver port from the default 8080")
}
