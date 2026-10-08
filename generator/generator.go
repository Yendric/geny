package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yendric/geny/common"
	"github.com/Yendric/geny/indexer/content"
	"github.com/Yendric/geny/islands"
	"github.com/Yendric/geny/util"
	"github.com/Yendric/geny/vite"
)

type Generator struct {
	cfg       common.Config
	templates *template.Template
	islands   *islands.Registry
	vite      *vite.Integration
}

func New(cfg common.Config, islandRegistry *islands.Registry) (*Generator, error) {
	g := &Generator{
		cfg:     cfg,
		islands: islandRegistry,
		vite:    vite.New(cfg),
	}

	templates, err := g.parseTemplates()
	if err != nil {
		return nil, err
	}
	g.templates = templates

	return g, nil
}

func (g *Generator) funcMap() template.FuncMap {
	return template.FuncMap{
		"stripTags":      util.StripTags,
		"truncate":       util.Truncate,
		"getCurrentYear": util.GetCurrentYear,
		"vite":           g.vite.Tags,
		"island":         g.island,
		"props":          props,
		"client":         client,
		"fallback":       fallback,
	}
}

func (g *Generator) island(name string) (islands.Island, error) {
	if !g.islands.Has(name) {
		return islands.Island{}, fmt.Errorf("island %s not found in %s", name, g.cfg.IslandsDir)
	}
	return islands.New(name), nil
}

func props(key string, value interface{}, island islands.Island) (islands.Island, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return island, fmt.Errorf("island %s: prop %s: %w", island.Name, key, err)
	}
	island, err = island.WithProp(key, raw)
	if err != nil {
		return island, fmt.Errorf("island %s: %w", island.Name, err)
	}
	return island, nil
}

func client(directive string, island islands.Island) (islands.Island, error) {
	d, err := islands.ParseDirective(directive)
	if err != nil {
		return island, fmt.Errorf("island %s: %w", island.Name, err)
	}
	island.Directive = d
	return island, nil
}

func fallback(html template.HTML, island islands.Island) islands.Island {
	island.Fallback = html
	return island
}

type page struct {
	templates *template.Template
	source    string
	usages    []islands.Usage
}

func (p *page) funcs() template.FuncMap {
	return template.FuncMap{
		"render":  p.render,
		"include": p.include,
	}
}

func (p *page) render(island islands.Island) template.HTML {
	p.usages = append(p.usages, islands.Usage{Island: island, Source: p.source})
	return island.HTML()
}

func (p *page) include(name string, data interface{}) (template.HTML, error) {
	var buf bytes.Buffer
	if err := p.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil
}

func (g *Generator) GenerateFiles(contentFiles []content.ContentFile) ([]islands.Usage, error) {
	collections := generateCollections(contentFiles)

	var usages []islands.Usage
	for _, contentFile := range contentFiles {
		contentFile.Collections = collections

		fileUsages, err := g.generateFile(contentFile)
		if err != nil {
			return nil, err
		}
		usages = append(usages, fileUsages...)
	}

	return usages, nil
}

func (g *Generator) parseTemplates() (*template.Template, error) {
	cfg := g.cfg
	var templateFiles []string
	for _, pattern := range []string{cfg.TemplatesDir + "/*.html", cfg.TemplatesDir + "/**/*.html"} {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("finding templates: %w", err)
		}
		templateFiles = append(templateFiles, matches...)
	}

	templates, err := template.New("").Funcs(g.funcMap()).Funcs((&page{}).funcs()).ParseFiles(templateFiles...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	return templates, nil
}

func (g *Generator) generateFile(contentFile content.ContentFile) ([]islands.Usage, error) {
	templates, err := g.templates.Clone()
	if err != nil {
		return nil, err
	}
	p := &page{templates: templates, source: contentFile.Path + " (template " + contentFile.Template.Name + ")"}
	templates.Funcs(p.funcs())

	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, contentFile.Template.Name+".html", contentFile); err != nil {
		return nil, fmt.Errorf("rendering %s: %w", contentFile.Path, err)
	}

	usages := append(append([]islands.Usage{}, contentFile.Islands...), p.usages...)
	html := buf.Bytes()
	if len(usages) > 0 {
		if !g.cfg.Vite.Enabled {
			return nil, fmt.Errorf("%s: islands need vite.enabled: true in %s", contentFile.Path, common.ConfigFile)
		}
		tags, err := g.vite.IslandTags()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", contentFile.Path, err)
		}
		html = injectBeforeBodyEnd(html, []byte(tags))
	}

	whereTo := util.StripHidden(contentFile.Path)
	whereTo = strings.ReplaceAll(whereTo, g.cfg.ContentDir, g.cfg.BuildDir)
	whereTo = util.StripExtension(whereTo)
	whereTo = util.StripEmpty(whereTo)

	outPath := util.GeneratePath(whereTo, "index.html")
	if contentFile.FileName == "index.md" || contentFile.FileName == "404.md" {
		outPath = util.GeneratePath(whereTo + ".html")
	} else if err := os.MkdirAll(whereTo, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating directory %s: %w", whereTo, err)
	}

	if err := os.WriteFile(outPath, html, 0o644); err != nil {
		return nil, fmt.Errorf("creating %s: %w", outPath, err)
	}
	return usages, nil
}

func injectBeforeBodyEnd(html, tags []byte) []byte {
	i := bytes.LastIndex(bytes.ToLower(html), []byte("</body>"))
	if i < 0 {
		return append(html, tags...)
	}
	out := make([]byte, 0, len(html)+len(tags))
	out = append(out, html[:i]...)
	out = append(out, tags...)
	return append(out, html[i:]...)
}
