package indexer

import (
	"fmt"
	"os"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/indexer/content"
	"github.com/Yendric/geny/indexer/template"
	"github.com/Yendric/geny/islands"
	"github.com/Yendric/geny/util"
	"github.com/yuin/goldmark"
)

type Indexer struct {
	cfg common.Config
	md  goldmark.Markdown
}

func New(cfg common.Config) *Indexer {
	return &Indexer{
		cfg: cfg,
		md:  newMarkdown(),
	}
}

type registries struct {
	templates *template.Registry
	islands   *islands.Registry
}

func (i *Indexer) IndexContent(islandRegistry *islands.Registry) ([]content.ContentFile, error) {
	return i.indexDirectory(registries{
		templates: template.NewRegistry(i.cfg.TemplatesDir),
		islands:   islandRegistry,
	}, i.cfg.ContentDir)
}

func (i *Indexer) indexDirectory(reg registries, directory string) ([]content.ContentFile, error) {
	indexed := []content.ContentFile{}

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filePath := util.GeneratePath(directory, file.Name())
		if file.IsDir() {
			indexedDirectory, err := i.indexDirectory(reg, filePath)
			if err != nil {
				return nil, err
			}

			indexed = append(indexed, indexedDirectory...)
		} else {
			indexedFile, err := i.indexFile(reg, filePath)
			if err != nil {
				return nil, err
			}

			indexed = append(indexed, indexedFile)
		}
	}

	return indexed, nil
}

func (i *Indexer) indexFile(reg registries, filePath string) (content.ContentFile, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("reading %s: %w", filePath, err)
	}
	fileStats, err := os.Stat(filePath)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("reading %s: %w", filePath, err)
	}

	parsed, err := i.parseMdFile(reg.islands, fileContent)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("parsing markdown in %s: %w", filePath, err)
	}
	for n := range parsed.islands {
		parsed.islands[n].Source = filePath
	}

	templateName, found := parsed.metaData["template"].(string)
	if !found {
		return content.ContentFile{}, fmt.Errorf("no template declared in %s", filePath)
	}

	fileTemplate, err := reg.templates.GetByName(templateName)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("%s: %w", filePath, err)
	}

	file := content.ContentFile{
		MetaData:   parsed.metaData,
		Content:    parsed.html,
		Islands:    parsed.islands,
		Headings:   parsed.headings,
		RawContent: fileContent,
		Path:       filePath,
		FileName:   fileStats.Name(),
		Url:        util.GenerateContentUrl(i.cfg.ContentDir, filePath),
		Template:   fileTemplate,
	}

	return file, nil
}
