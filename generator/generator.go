package generator

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/indexer/content"
	"github.com/Yendric/geny/util"
)

var funcMap = template.FuncMap{
	"stripTags":      util.StripTags,
	"truncate":       util.Truncate,
	"getCurrentYear": util.GetCurrentYear,
}

func GenerateFiles(contentFiles []content.ContentFile) error {
	templates, err := parseTemplates()
	if err != nil {
		return err
	}

	collections := generateCollections(contentFiles)

	for _, contentFile := range contentFiles {
		contentFile.Collections = collections

		if err := generateFile(templates, contentFile); err != nil {
			return err
		}
	}

	return nil
}

func parseTemplates() (*template.Template, error) {
	var templateFiles []string
	for _, pattern := range []string{common.TEMPLATES_DIR + "/*.html", common.TEMPLATES_DIR + "/**/*.html"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("finding templates: %w", err)
		}
		templateFiles = append(templateFiles, matches...)
	}

	templates, err := template.New("").Funcs(funcMap).ParseFiles(templateFiles...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	return templates, nil
}

func generateFile(templates *template.Template, contentFile content.ContentFile) error {
	whereTo := util.StripHidden(contentFile.Path)
	whereTo = strings.ReplaceAll(whereTo, common.CONTENT_DIR, common.BUILD_DIR)
	whereTo = util.StripExtension(whereTo)
	whereTo = util.StripEmpty(whereTo)

	var (
		buildFile *os.File
		err       error
	)
	if contentFile.FileName == "index.md" || contentFile.FileName == "404.md" {
		buildFile, err = os.Create(util.GeneratePath(whereTo + ".html"))
		if err != nil {
			return fmt.Errorf("creating %s: %w", whereTo+".html", err)
		}
	} else {
		err = os.MkdirAll(whereTo, os.ModePerm)
		if err != nil {
			return fmt.Errorf("creating directory %s: %w", whereTo, err)
		}

		buildFile, err = os.Create(util.GeneratePath(whereTo, "index.html"))
		if err != nil {
			return fmt.Errorf("creating %s: %w", util.GeneratePath(whereTo, "index.html"), err)
		}
	}

	if err := templates.ExecuteTemplate(buildFile, contentFile.Template.Name+".html", contentFile); err != nil {
		buildFile.Close()
		return fmt.Errorf("rendering %s: %w", contentFile.Path, err)
	}

	return buildFile.Close()
}
