package islands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func testRegistry(t *testing.T, names ...string) *Registry {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		writeFile(t, filepath.Join(dir, name+".tsx"), "export default function X() { return null }")
	}
	reg, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func TestScanNamesNestedIslands(t *testing.T) {
	reg := testRegistry(t, "Counter", "ui/Button")
	for _, name := range []string{"Counter", "ui/Button"} {
		if !reg.Has(name) {
			t.Errorf("expected island %s", name)
		}
	}
}

func TestScanMissingDirIsEmpty(t *testing.T) {
	reg, err := Scan(filepath.Join(t.TempDir(), "missing"))
	if err != nil {
		t.Fatal(err)
	}
	if reg.Has("Counter") {
		t.Error("expected empty registry")
	}
}

func TestScanRejectsHTMLElementNames(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "button.tsx"), "")
	if _, err := Scan(dir); err == nil {
		t.Fatal("expected error for island named like an HTML element")
	}
}

func TestScanRejectsDuplicates(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "Counter.tsx"), "")
	writeFile(t, filepath.Join(dir, "Counter.jsx"), "")
	if _, err := Scan(dir); err == nil {
		t.Fatal("expected error for island defined twice")
	}
}

func TestParseTag(t *testing.T) {
	reg := testRegistry(t, "Counter", "ui/Button")

	tests := []struct {
		src         string
		props       string
		directive   Directive
		selfClosing bool
	}{
		{`<Counter />`, `{}`, Load, true},
		{`<Counter/>`, `{}`, Load, true},
		{`<Counter client:visible start={5} />`, `{"start":5}`, Visible, true},
		{`<Counter label="Hi" other='x' open>`, `{"label":"Hi","other":"x","open":true}`, Load, false},
		{"<Counter\n  items={[1, 2]}\n  nested={{\"a\": \"}\"}}\n/>", `{"items":[1,2],"nested":{"a":"}"}}`, Load, true},
		{`<ui/Button client:load />`, `{}`, Load, true},
	}
	for _, tt := range tests {
		got, err := parseTag([]byte(tt.src), 0, reg)
		if err != nil {
			t.Errorf("%s: %v", tt.src, err)
			continue
		}
		if p := got.island.PropsJSON(); p != tt.props {
			t.Errorf("%s: props %s, want %s", tt.src, p, tt.props)
		}
		if got.island.Directive != tt.directive {
			t.Errorf("%s: directive %s, want %s", tt.src, got.island.Directive, tt.directive)
		}
		if got.selfClosing != tt.selfClosing {
			t.Errorf("%s: selfClosing %v, want %v", tt.src, got.selfClosing, tt.selfClosing)
		}
		if got.end != len(tt.src) {
			t.Errorf("%s: end %d, want %d", tt.src, got.end, len(tt.src))
		}
	}
}

func TestParseTagNotIsland(t *testing.T) {
	reg := testRegistry(t, "Counter")
	for _, src := range []string{`<div>`, `<counter />`, `</Counter>`, `<Other />`} {
		if _, err := parseTag([]byte(src), 0, reg); err != errNotIsland {
			t.Errorf("%s: got %v, want errNotIsland", src, err)
		}
	}
}

func TestParseTagErrors(t *testing.T) {
	reg := testRegistry(t, "Counter")
	for _, src := range []string{
		`<Counter client:hover />`,
		`<Counter client:load client:visible />`,
		`<Counter start={5 />`,
		`<Counter start={nope} />`,
		`<Counter start=5 />`,
		`<Counter a="1" a="2" />`,
		`<Counter client:visible="x" />`,
		`<Counter foo:bar />`,
		`<Counter`,
	} {
		if _, err := parseTag([]byte(src), 0, reg); err == nil || err == errNotIsland {
			t.Errorf("%s: expected a syntax error, got %v", src, err)
		}
	}
}

func TestIslandHTMLEscapes(t *testing.T) {
	island, err := New("Counter").WithProp("label", []byte(`"<b>&'"`))
	if err != nil {
		t.Fatal(err)
	}
	island.Fallback = "<p>Loading</p>"

	want := `<geny-island component="Counter" client="load" props="{&#34;label&#34;:&#34;&lt;b&gt;&amp;&#39;&#34;}"><p>Loading</p></geny-island>`
	if got := string(island.HTML()); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestWithPropRejectsDuplicates(t *testing.T) {
	island, err := New("Counter").WithProp("start", []byte("1"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := island.WithProp("start", []byte("2")); err == nil {
		t.Fatal("expected duplicate prop error")
	}
}

func TestCheckSourceMapsLines(t *testing.T) {
	reg := testRegistry(t, "Counter")
	stateDir := t.TempDir()
	usages := []Usage{
		{Island: New("Counter"), Source: "content/a.md", Line: 3},
		{Island: New("Counter"), Source: "content/b.md", Line: 7},
	}

	source, lines, err := checkSource(stateDir, reg, usages)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Split(strings.TrimSpace(string(source)), "\n")
	if len(got) != 4 {
		t.Fatalf("expected 4 lines, got:\n%s", source)
	}
	if !strings.HasPrefix(got[1], "import type Island0 from ") {
		t.Errorf("line 2: %s", got[1])
	}
	if got[3] != "({}) satisfies ComponentProps<typeof Island0>;" {
		t.Errorf("line 4: %s", got[3])
	}

	err = describe(diagnostic{file: filepath.ToSlash(filepath.Join(stateDir, checkFile)), line: 4, message: "bad"}, filepath.ToSlash(filepath.Join(stateDir, checkFile)), lines)
	if want := "content/b.md:7: island Counter: bad"; err.Error() != want {
		t.Errorf("got %q, want %q", err, want)
	}
}

func TestDedupeKeepsDistinctMarkdownLines(t *testing.T) {
	template := Usage{Island: New("Counter"), Source: "page (template x)"}
	usages := dedupe([]Usage{
		template,
		template,
		{Island: New("Counter"), Source: "content/a.md", Line: 3},
		{Island: New("Counter"), Source: "content/a.md", Line: 5},
	})
	if len(usages) != 3 {
		t.Errorf("expected 3 usages, got %d", len(usages))
	}
}

func TestParseDiagnostics(t *testing.T) {
	out := []byte(".geny/islands.check.ts(4,2): error TS2322: Type 'string' is not assignable to type 'number'.\n" +
		"islands/Counter.tsx(3,1): error TS2353: Object literal may only specify known properties.\n" +
		"  Extra detail.\n")
	got := parseDiagnostics(out)
	if len(got) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(got))
	}
	if got[0].file != ".geny/islands.check.ts" || got[0].line != 4 {
		t.Errorf("first diagnostic: %+v", got[0])
	}
	if !strings.HasSuffix(got[1].message, "\n  Extra detail.") {
		t.Errorf("continuation not attached: %q", got[1].message)
	}
}
