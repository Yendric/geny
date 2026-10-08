package site

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Yendric/geny/common"
)

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func testSite(t *testing.T, template string) common.Config {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	cfg := common.DefaultConfig()
	cfg.Vite.Enabled = true

	writeFile(t, filepath.Join(cfg.IslandsDir, "Counter.tsx"), "")
	writeFile(t, filepath.Join(cfg.TemplatesDir, "page.html"), template)
	writeFile(t, filepath.Join(cfg.TemplatesDir, "shared", "fallback.html"), `{{ define "counter-fallback" }}<p>{{ .MetaData.title }}</p>{{ end }}`)
	writeFile(t, filepath.Join(cfg.BuildDir, ".vite", "manifest.json"), `{
		"virtual:geny/islands": {"file": "assets/islands-abc.js", "isEntry": true}
	}`)
	return cfg
}

func readBuild(t *testing.T, cfg common.Config, name string) string {
	t.Helper()
	out, err := os.ReadFile(filepath.Join(cfg.BuildDir, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestTemplateIslandChain(t *testing.T) {
	cfg := testSite(t, `<html><body>{{ island "Counter" | props "start" 5 | props "label" .MetaData.title | client "visible" | fallback (include "counter-fallback" .) | render }}</body></html>`)
	writeFile(t, filepath.Join(cfg.ContentDir, "index.md"), "---\ntemplate: page\ntitle: Home\n---\n")

	result, err := New(cfg).Generate()
	if err != nil {
		t.Fatal(err)
	}

	out := readBuild(t, cfg, "index.html")
	want := `<body><geny-island component="Counter" client="visible" props="{&#34;start&#34;:5,&#34;label&#34;:&#34;Home&#34;}"><p>Home</p></geny-island><script type="module" src="/assets/islands-abc.js"></script></body>`
	if !strings.Contains(out, want) {
		t.Errorf("got:\n%s\nwant to contain:\n%s", out, want)
	}
	if len(result.Usages) != 1 || result.Usages[0].Line != 0 {
		t.Errorf("usages: %+v", result.Usages)
	}
}

func TestMarkdownIslandInjectsRuntime(t *testing.T) {
	cfg := testSite(t, `<html><body>{{ .Content }}</body></html>`)
	writeFile(t, filepath.Join(cfg.ContentDir, "index.md"), "---\ntemplate: page\n---\n\n<Counter start={1} />\n")
	writeFile(t, filepath.Join(cfg.ContentDir, "plain.md"), "---\ntemplate: page\n---\n\nNo islands.\n")

	result, err := New(cfg).Generate()
	if err != nil {
		t.Fatal(err)
	}

	if out := readBuild(t, cfg, "index.html"); !strings.Contains(out, `src="/assets/islands-abc.js"`) {
		t.Errorf("expected runtime script, got:\n%s", out)
	}
	if out := readBuild(t, cfg, "plain/index.html"); strings.Contains(out, "<script") {
		t.Errorf("pages without islands must not load the runtime, got:\n%s", out)
	}
	if len(result.Usages) != 1 || result.Usages[0].Line != 5 || !strings.HasSuffix(result.Usages[0].Source, "index.md") {
		t.Errorf("usages: %+v", result.Usages)
	}
}

func TestUnknownTemplateIsland(t *testing.T) {
	cfg := testSite(t, `{{ island "Missing" | render }}`)
	writeFile(t, filepath.Join(cfg.ContentDir, "index.md"), "---\ntemplate: page\n---\n")

	if _, err := New(cfg).Generate(); err == nil || !strings.Contains(err.Error(), "island Missing not found") {
		t.Fatalf("expected unknown island error, got %v", err)
	}
}

func TestIslandsRequireVite(t *testing.T) {
	cfg := testSite(t, `{{ .Content }}`)
	cfg.Vite.Enabled = false
	writeFile(t, filepath.Join(cfg.ContentDir, "index.md"), "---\ntemplate: page\n---\n\n<Counter />\n")

	if _, err := New(cfg).Generate(); err == nil || !strings.Contains(err.Error(), "vite.enabled") {
		t.Fatalf("expected vite error, got %v", err)
	}
}
