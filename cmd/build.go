package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

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
	Run: func(cmd *cobra.Command, args []string) {
		runStepQuit("Removing old builds", func() error {
			err := os.RemoveAll(common.BUILD_DIR)
			return err
		})

		runStepQuit("Copying assets", func() error {
			err := copy.Copy(common.PUBLIC_DIR, common.BUILD_DIR)
			return err
		})

		runCmd, err := cmd.Flags().GetString("run")
		if err != nil {
			log.Fatal(err)
		}
		if runCmd != "" {
			runStepQuit("Running custom build command", func() error {
				err := exec.Command("sh", "-c", runCmd).Run()
				return err
			})
		}

		runStepQuit("Generating html", func() error {
			content, err := indexer.IndexContent()
			if err != nil {
				return err
			}

			err = generator.GenerateFiles(content)
			return err
		})

		color.New(color.BgGreen).Println("Your site has been generated!")

		shouldServe, err := cmd.Flags().GetBool("serve")
		if err != nil || !shouldServe {
			return
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			log.Fatal(err)
		}

		runStepQuit(fmt.Sprintf("Serving the site on port %d", port), func() error { return serve(port) })
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().BoolP("serve", "s", false, "Serve the site on a local webserver")
	buildCmd.Flags().IntP("port", "p", 8080, "Change the local webserver port from the default 8080")
	buildCmd.Flags().String("run", "", "Run a command before building the site")
}
