package site

import (
	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/generator"
	"github.com/Yendric/geny/indexer"
	"github.com/Yendric/geny/islands"
)

// state shared across builds of one site
type Site struct {
	cfg     common.Config
	indexer *indexer.Indexer
}

func New(cfg common.Config) *Site {
	return &Site{
		cfg:     cfg,
		indexer: indexer.New(cfg),
	}
}

type Result struct {
	Islands *islands.Registry
	Usages  []islands.Usage
}

// indexes the content and renders it into the build directory,
// templates and islands are rescanned on every call
func (s *Site) Generate() (Result, error) {
	islandRegistry, err := islands.Scan(s.cfg.IslandsDir)
	if err != nil {
		return Result{}, err
	}

	contentFiles, err := s.indexer.IndexContent(islandRegistry)
	if err != nil {
		return Result{}, err
	}

	gen, err := generator.New(s.cfg, islandRegistry)
	if err != nil {
		return Result{}, err
	}

	usages, err := gen.GenerateFiles(contentFiles)
	if err != nil {
		return Result{}, err
	}
	return Result{Islands: islandRegistry, Usages: usages}, nil
}

func (r Result) CheckIslands() error {
	return islands.Check(common.StateDir, r.Islands, r.Usages)
}
