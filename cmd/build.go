package cmd

import (
	"fmt"
	"os"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/generator"
	"github.com/Yendric/geny/indexer"
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

		runCmd, err := cmd.Flags().GetString("run")
		if err != nil {
			return err
		}
		if runCmd != "" {
			runStepQuit("Running custom build command", func() error {
				return runBuildCommand(runCmd)
			})
		}

		runStepQuit("Generating html", func() error {
			content, err := indexer.IndexContent(cfg)
			if err != nil {
				return err
			}

			return generator.GenerateFiles(cfg, content)
		})

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

		runStepQuit(fmt.Sprintf("Serving the site on port %d", port), func() error { return serve(cfg.BuildDir, port) })
		return nil
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolP("serve", "s", false, "Serve the site on a local webserver")
	buildCmd.Flags().IntP("port", "p", 8080, "Change the local webserver port from the default 8080")
	buildCmd.Flags().String("run", "", "Run a command before building the site")
}
