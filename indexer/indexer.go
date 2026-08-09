package indexer

import (
	"fmt"
	"os"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/indexer/content"
	"github.com/Yendric/geny/indexer/template"
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

func (i *Indexer) IndexContent() ([]content.ContentFile, error) {
	templates := template.NewRegistry(i.cfg.TemplatesDir)
	return i.indexDirectory(templates, i.cfg.ContentDir)
}

func (i *Indexer) indexDirectory(templates *template.Registry, directory string) ([]content.ContentFile, error) {
	indexed := []content.ContentFile{}

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filePath := util.GeneratePath(directory, file.Name())
		if file.IsDir() {
			indexedDirectory, err := i.indexDirectory(templates, filePath)
			if err != nil {
				return nil, err
			}

			indexed = append(indexed, indexedDirectory...)
		} else {
			indexedFile, err := i.indexFile(templates, filePath)
			if err != nil {
				return nil, err
			}

			indexed = append(indexed, indexedFile)
		}
	}

	return indexed, nil
}

func (i *Indexer) indexFile(templates *template.Registry, filePath string) (content.ContentFile, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("reading %s: %w", filePath, err)
	}
	fileStats, err := os.Stat(filePath)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("reading %s: %w", filePath, err)
	}

	metaData, renderedContent, err := i.parseMdFile(fileContent)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("parsing markdown in %s: %w", filePath, err)
	}

	templateName, found := metaData["template"].(string)
	if !found {
		return content.ContentFile{}, fmt.Errorf("no template declared in %s", filePath)
	}

	fileTemplate, err := templates.GetByName(templateName)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("%s: %w", filePath, err)
	}

	file := content.ContentFile{
		MetaData:   metaData,
		Content:    renderedContent,
		RawContent: fileContent,
		Path:       filePath,
		FileName:   fileStats.Name(),
		Url:        util.GenerateContentUrl(i.cfg.ContentDir, filePath),
		Template:   fileTemplate,
	}

	return file, nil
}
