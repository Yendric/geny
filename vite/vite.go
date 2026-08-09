package vite

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Yendric/geny/common"
)

type Integration struct {
	dev          bool
	devServerURL string
	hotFile      string
	manifestPath string

	manifestOnce sync.Once
	manifest     manifest
	manifestErr  error
}

func New(cfg common.Config) *Integration {
	return &Integration{
		dev:          cfg.DevMode,
		devServerURL: cfg.Vite.DevServerURL,
		hotFile:      cfg.Vite.HotFile,
		manifestPath: filepath.Join(cfg.BuildDir, ".vite", "manifest.json"),
	}
}

// renders the html tags loading the given entry points, either from the
// dev server or from the build manifest
func (i *Integration) Tags(entries ...string) (template.HTML, error) {
	if i.dev {
		if base, ok := i.devServer(); ok {
			return i.devTags(base, entries), nil
		}
	}
	return i.prodTags(entries)
}

func (i *Integration) devServer() (string, bool) {
	data, err := os.ReadFile(i.hotFile)
	if err != nil {
		return "", false
	}
	url := strings.TrimSpace(string(data))
	if url == "" {
		url = i.devServerURL
	}
	return url, true
}

func (i *Integration) devTags(base string, entries []string) template.HTML {
	var b strings.Builder
	b.WriteString(scriptTag(base + "/@vite/client"))
	for _, entry := range entries {
		b.WriteString(scriptTag(base + "/" + entry))
	}
	return template.HTML(b.String())
}

func (i *Integration) prodTags(entries []string) (template.HTML, error) {
	i.manifestOnce.Do(func() {
		i.manifest, i.manifestErr = loadManifest(i.manifestPath)
	})
	if i.manifestErr != nil {
		return "", i.manifestErr
	}
	m := i.manifest

	var b strings.Builder
	seenCSS := map[string]bool{}
	for _, entry := range entries {
		chunk, ok := m[entry]
		if !ok {
			return "", fmt.Errorf("vite: entry %q not found in %s", entry, i.manifestPath)
		}

		for _, css := range m.cssFor(entry, map[string]bool{}) {
			if !seenCSS[css] {
				seenCSS[css] = true
				b.WriteString(linkTag("/" + css))
			}
		}
		if strings.HasSuffix(chunk.File, ".css") {
			if !seenCSS[chunk.File] {
				seenCSS[chunk.File] = true
				b.WriteString(linkTag("/" + chunk.File))
			}
		} else {
			b.WriteString(scriptTag("/" + chunk.File))
		}
	}
	return template.HTML(b.String()), nil
}

type chunk struct {
	File    string   `json:"file"`
	CSS     []string `json:"css"`
	Imports []string `json:"imports"`
}

type manifest map[string]chunk

// dfs to determine all css dependencies of a given entry point
func (m manifest) cssFor(name string, seen map[string]bool) []string {
	if seen[name] {
		return nil
	}
	seen[name] = true

	chunk, ok := m[name]
	if !ok {
		return nil
	}

	css := append([]string{}, chunk.CSS...)
	for _, imp := range chunk.Imports {
		css = append(css, m.cssFor(imp, seen)...)
	}
	return css
}

func loadManifest(path string) (manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("vite: reading manifest %s: %w", path, err)
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("vite: parsing manifest %s: %w", path, err)
	}
	return m, nil
}

func scriptTag(src string) string {
	return fmt.Sprintf("<script type=\"module\" src=\"%s\"></script>", src)
}

func linkTag(href string) string {
	return fmt.Sprintf("<link rel=\"stylesheet\" href=\"%s\">", href)
}
