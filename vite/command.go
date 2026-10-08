package vite

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/util"
)

func command(cfg common.Config, run string) *exec.Cmd {
	cmd := util.ShellCommand(run)
	cmd.Env = append(os.Environ(),
		"GENY_CONTENT_DIR="+cfg.ContentDir,
		"GENY_TEMPLATES_DIR="+cfg.TemplatesDir,
		"GENY_PUBLIC_DIR="+cfg.PublicDir,
		"GENY_ISLANDS_DIR="+cfg.IslandsDir,
	)
	return cmd
}

func Build(cfg common.Config) error {
	out, err := command(cfg, cfg.Vite.BuildCommand).CombinedOutput()
	if err != nil {
		return fmt.Errorf("running %q: %w\n%s", cfg.Vite.BuildCommand, err, out)
	}
	return nil
}
