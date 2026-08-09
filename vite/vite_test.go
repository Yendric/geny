package vite

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Yendric/geny/common"
)

func testConfig(t *testing.T) common.Config {
	t.Helper()
	dir := t.TempDir()
	cfg := common.DefaultConfig()
	cfg.BuildDir = dir
	cfg.Vite.HotFile = filepath.Join(dir, "hot")
	return cfg
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestTagsDevMode(t *testing.T) {
	cfg := testConfig(t)
	cfg.DevMode = true
	writeFile(t, cfg.Vite.HotFile, "http://localhost:5180")

	got, err := New(cfg).Tags("src/main.ts")
	if err != nil {
		t.Fatal(err)
	}

	out := string(got)
	for _, want := range []string{
		`<script type="module" src="http://localhost:5180/@vite/client"></script>`,
		`<script type="module" src="http://localhost:5180/src/main.ts"></script>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dev tags missing %q\ngot: %s", want, out)
		}
	}
}

func TestTagsDevModeFallsBackToConfiguredURL(t *testing.T) {
	cfg := testConfig(t)
	cfg.DevMode = true
	writeFile(t, cfg.Vite.HotFile, "") // present but empty

	got, err := New(cfg).Tags("src/main.ts")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), cfg.Vite.DevServerURL+"/@vite/client") {
		t.Errorf("expected fallback to %s\ngot: %s", cfg.Vite.DevServerURL, got)
	}
}

func TestTagsProdMode(t *testing.T) {
	cfg := testConfig(t)
	writeFile(t, filepath.Join(cfg.BuildDir, ".vite", "manifest.json"), `{
		"src/main.ts": {
			"file": "assets/main-abc123.js",
			"css": ["assets/main-def456.css"],
			"imports": ["_vendor-xyz.js"]
		},
		"_vendor-xyz.js": {
			"file": "assets/vendor-xyz.js",
			"css": ["assets/vendor-789.css"]
		}
	}`)

	got, err := New(cfg).Tags("src/main.ts")
	if err != nil {
		t.Fatal(err)
	}

	out := string(got)
	for _, want := range []string{
		`<script type="module" src="/assets/main-abc123.js"></script>`,
		`<link rel="stylesheet" href="/assets/main-def456.css">`,
		`<link rel="stylesheet" href="/assets/vendor-789.css">`, // pulled in via imports
	} {
		if !strings.Contains(out, want) {
			t.Errorf("prod tags missing %q\ngot: %s", want, out)
		}
	}
}

func TestTagsBuildModeIgnoresHotFile(t *testing.T) {
	cfg := testConfig(t)
	writeFile(t, cfg.Vite.HotFile, "http://localhost:5173") // stale, DevMode is false
	writeFile(t, filepath.Join(cfg.BuildDir, ".vite", "manifest.json"), `{
		"src/main.ts": {"file": "assets/main-abc123.js"}
	}`)

	got, err := New(cfg).Tags("src/main.ts")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "@vite/client") {
		t.Errorf("build mode must not emit dev tags\ngot: %s", got)
	}
	if !strings.Contains(string(got), "/assets/main-abc123.js") {
		t.Errorf("expected manifest asset\ngot: %s", got)
	}
}

func TestTagsDevModeCssEntry(t *testing.T) {
	cfg := testConfig(t)
	cfg.DevMode = true
	writeFile(t, cfg.Vite.HotFile, "http://localhost:5173")

	got, err := New(cfg).Tags("src/style.css", "src/main.js")
	if err != nil {
		t.Fatal(err)
	}

	out := string(got)
	for _, want := range []string{
		`<link rel="stylesheet" href="http://localhost:5173/src/style.css">`,
		`<script type="module" src="http://localhost:5173/src/main.js"></script>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("dev tags missing %q\ngot: %s", want, out)
		}
	}
	if strings.Contains(out, `<script type="module" src="http://localhost:5173/src/style.css"></script>`) {
		t.Errorf("css entry must not be a script tag in dev\ngot: %s", out)
	}
}

func TestTagsProdModeCssEntry(t *testing.T) {
	cfg := testConfig(t)
	writeFile(t, filepath.Join(cfg.BuildDir, ".vite", "manifest.json"), `{
		"src/style.css": {"file": "assets/style-abc123.css", "isEntry": true}
	}`)

	got, err := New(cfg).Tags("src/style.css")
	if err != nil {
		t.Fatal(err)
	}

	out := string(got)
	if !strings.Contains(out, `<link rel="stylesheet" href="/assets/style-abc123.css">`) {
		t.Errorf("expected stylesheet link for css entry\ngot: %s", out)
	}
	if strings.Contains(out, "<script") {
		t.Errorf("css entry must not emit a script tag\ngot: %s", out)
	}
}

func TestTagsProdModeUnknownEntry(t *testing.T) {
	cfg := testConfig(t)
	writeFile(t, filepath.Join(cfg.BuildDir, ".vite", "manifest.json"), `{}`)

	if _, err := New(cfg).Tags("src/missing.ts"); err == nil {
		t.Fatal("expected error for entry missing from manifest")
	}
}
