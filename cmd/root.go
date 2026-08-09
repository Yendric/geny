package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// stamped by release build via ldflags
var version = "dev"

var rootCmd = &cobra.Command{
	Use:          "geny",
	Short:        "Geny static site generator.",
	Long:         `Geny static site generator.`,
	Version:      version,
	SilenceUsage: true,
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}

func init() {

}
