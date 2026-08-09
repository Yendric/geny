package site

import (
	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/generator"
	"github.com/Yendric/geny/indexer"
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

// indexes the content and renders it into the build directory,
// templates are re-parsed on every call
func (s *Site) Generate() error {
	contentFiles, err := s.indexer.IndexContent()
	if err != nil {
		return err
	}

	gen, err := generator.New(s.cfg)
	if err != nil {
		return err
	}

	return gen.GenerateFiles(contentFiles)
}
