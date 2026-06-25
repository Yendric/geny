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
	ContentDir   string `yaml:"contentDir"`
	TemplatesDir string `yaml:"templatesDir"`
	BuildDir     string `yaml:"buildDir"`
	PublicDir    string `yaml:"publicDir"`
}

func DefaultConfig() Config {
	return Config{
		ContentDir:   "content",
		TemplatesDir: "templates",
		BuildDir:     "build",
		PublicDir:    "public",
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
