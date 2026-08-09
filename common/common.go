package common

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v2"
)

const ConfigFile = "geny.yaml"

type Config struct {
	ContentDir   string     `yaml:"contentDir"`
	TemplatesDir string     `yaml:"templatesDir"`
	BuildDir     string     `yaml:"buildDir"`
	PublicDir    string     `yaml:"publicDir"`
	Vite         ViteConfig `yaml:"vite"`

	// DevMode is set by `geny watch`, used to make sure
	// a stale hot file cannot leak into `geny build` output
	DevMode bool `yaml:"-"`
}

type ViteConfig struct {
	Enabled      bool   `yaml:"enabled"`
	BuildCommand string `yaml:"buildCommand"`
	DevCommand   string `yaml:"devCommand"`
	DevServerURL string `yaml:"devServerURL"`
	// HotFile marks dev mode while present, contains dev server URL
	HotFile string `yaml:"hotFile"`
}

func DefaultConfig() Config {
	return Config{
		ContentDir:   "content",
		TemplatesDir: "templates",
		BuildDir:     "build",
		PublicDir:    "public",
		Vite: ViteConfig{
			Enabled:      false,
			BuildCommand: "npm run build",
			DevCommand:   "npm run dev",
			DevServerURL: "http://localhost:5173",
			HotFile:      ".geny/hot",
		},
	}
}

func LoadConfig() (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(ConfigFile)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, fmt.Errorf("reading %s: %w", ConfigFile, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing %s: %w", ConfigFile, err)
	}
	return cfg, nil
}
