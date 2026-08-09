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
	"github.com/Yendric/geny/vite"
)

type Generator struct {
	cfg       common.Config
	templates *template.Template
}

func New(cfg common.Config) (*Generator, error) {
	templates, err := parseTemplates(cfg)
	if err != nil {
		return nil, err
	}

	return &Generator{
		cfg:       cfg,
		templates: templates,
	}, nil
}

func buildFuncMap(cfg common.Config) template.FuncMap {
	v := vite.New(cfg)
	return template.FuncMap{
		"stripTags":      util.StripTags,
		"truncate":       util.Truncate,
		"getCurrentYear": util.GetCurrentYear,
		"vite":           v.Tags,
	}
}

func (g *Generator) GenerateFiles(contentFiles []content.ContentFile) error {
	collections := generateCollections(contentFiles)

	for _, contentFile := range contentFiles {
		contentFile.Collections = collections

		if err := g.generateFile(contentFile); err != nil {
			return err
		}
	}

	return nil
}

func parseTemplates(cfg common.Config) (*template.Template, error) {
	var templateFiles []string
	for _, pattern := range []string{cfg.TemplatesDir + "/*.html", cfg.TemplatesDir + "/**/*.html"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("finding templates: %w", err)
		}
		templateFiles = append(templateFiles, matches...)
	}

	templates, err := template.New("").Funcs(buildFuncMap(cfg)).ParseFiles(templateFiles...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	return templates, nil
}

func (g *Generator) generateFile(contentFile content.ContentFile) error {
	whereTo := util.StripHidden(contentFile.Path)
	whereTo = strings.ReplaceAll(whereTo, g.cfg.ContentDir, g.cfg.BuildDir)
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

	if err := g.templates.ExecuteTemplate(buildFile, contentFile.Template.Name+".html", contentFile); err != nil {
		buildFile.Close()
		return fmt.Errorf("rendering %s: %w", contentFile.Path, err)
	}

	return buildFile.Close()
}
