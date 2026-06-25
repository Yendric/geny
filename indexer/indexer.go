package indexer

import (
	"fmt"
	"os"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/indexer/content"
	"github.com/Yendric/geny/indexer/template"
	"github.com/Yendric/geny/util"
)

func IndexContent() ([]content.ContentFile, error) {
	indexedContent, err := indexDirectory(common.CONTENT_DIR)
	if err != nil {
		return nil, err
	}

	return indexedContent, nil
}

func indexDirectory(directory string) ([]content.ContentFile, error) {
	content := []content.ContentFile{}

	files, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		filePath := util.GeneratePath(directory, file.Name())
		if file.IsDir() {
			indexedDirectory, err := indexDirectory(filePath)
			if err != nil {
				return nil, err
			}

			content = append(content, indexedDirectory...)
		} else {
			indexedFile, err := indexFile(filePath)
			if err != nil {
				return nil, err
			}

			content = append(content, indexedFile)
		}
	}

	return content, nil
}

func indexFile(filePath string) (content.ContentFile, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("reading %s: %w", filePath, err)
	}
	fileStats, err := os.Stat(filePath)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("reading %s: %w", filePath, err)
	}

	metaData, renderedContent, err := ParseMdFile(fileContent)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("parsing markdown in %s: %w", filePath, err)
	}

	templateName, found := metaData["template"].(string)
	if !found {
		return content.ContentFile{}, fmt.Errorf("no template declared in %s", filePath)
	}

	template, err := template.GetByName(templateName)
	if err != nil {
		return content.ContentFile{}, fmt.Errorf("%s: %w", filePath, err)
	}

	file := content.ContentFile{
		MetaData:   metaData,
		Content:    renderedContent,
		RawContent: fileContent,
		Path:       filePath,
		FileName:   fileStats.Name(),
		Url:        util.GenerateContentUrl(filePath),
		Template:   template,
	}

	return file, nil
}
